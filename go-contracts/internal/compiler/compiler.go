package compiler

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go-contracts/internal/config"
	"go-contracts/internal/directory"
	"go-contracts/internal/logger"
	"go-contracts/internal/types"
	"go-contracts/internal/validator"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	solc "github.com/lmittmann/go-solc"
)

// Regex for matching multiple delimiters (comma, pipe, semicolon, space)
var delimiterRegex = regexp.MustCompile(`[,\|;\s]+`)

func Compile(cfg *config.Config) error {
	base := cfg.ContractsBaseDir
	if err := directory.CheckDirExists(base); err != nil {
		logger.Logger.Fatal().Msgf("Base directory does not exist: %s", base)
		return nil
	}

	// If multiple contracts are allowed, we split using the group delimiter
	var groups []string
	contractMap := make(map[string][]string)

	// Check if we are allowing multiple contracts
	if cfg.ContractsAllowMultiple {
		groupDelimiter := cfg.ContractsGroupDelimeter
		contractDirGroups := strings.Split(cfg.ContractsDir, groupDelimiter)
		contractNameGroups := strings.Split(cfg.ContractsNames, groupDelimiter)

		if len(contractDirGroups) != len(contractNameGroups) {
			logger.Logger.Fatal().Msgf("Mismatch between contract directory groups and contract name groups")
			return nil
		}

		// Further split each group by the subgroup delimiter if ContractsAllowMultiple is true
		for i, _ := range contractDirGroups {
			// Split by subgroup delimiter
			subgroupDelimeter := cfg.ContractsSubGroupDelimeter
			contractDirs := strings.Split(contractDirGroups[i], subgroupDelimeter)
			contractNames := strings.Split(contractNameGroups[i], subgroupDelimeter)

			// Check if the number of directories are empty
			if len(contractDirs) != 1 || len(contractNames) < 1 {
				logger.Logger.Fatal().Msgf("No contract directories or names found in subgroup %d", i)
				return nil
			}

			// Check if the contracts directory exists
			contractPath := filepath.Join(base, contractDirs[0])
			if err := directory.CheckDirExists(contractPath); err != nil {
				logger.Logger.Fatal().Msgf("Contracts directory does not exist: %s", contractDirs[0])
				return nil
			}

			groups = append(groups, contractPath)
			contractMap[contractPath] = contractNames
		}

	} else {
		// If comma-separated values are provided, log a warning or error
		if delimiterRegex.MatchString(cfg.ContractsDir) || delimiterRegex.MatchString(cfg.ContractsNames) {
			logger.Logger.Fatal().Msgf("Multiple contracts provided, but multiple contracts are not allowed")
			return nil
		}

		// Check if the contracts directory exists
		contractPath := filepath.Join(base, cfg.ContractsDir)
		if err := directory.CheckDirExists(contractPath); err != nil {
			logger.Logger.Fatal().Msgf("Contracts directory does not exist: %s", cfg.ContractsDir)
			return nil
		}
		groups = []string{contractPath}

		contractMap[contractPath] = []string{cfg.ContractsNames}
	}

	buildDir := cfg.SolcOutputDir
	// Check if the build directory exists
	err := directory.DeleteDir(buildDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.Logger.Fatal().Msgf("Failed to delete build directory: %s", err)
		return nil
	}

	if err := directory.CreateDirIfNotExists(buildDir); err != nil {
		logger.Logger.Fatal().Msgf("Failed to create build directory: %s", err)
		return nil
	}

	// Solc options
	var solcOptions []solc.Option

	// Parse remappings from environment variable
	remappings, err := parseRemappings(cfg.ContractsRemappings)
	if err != nil {
		logger.Logger.Fatal().Msgf("Failed to parse remappings: %s", err)
		return nil
	}

	if len(remappings) > 0 {
		for _, remapping := range remappings {
			logger.Logger.Info().Msgf("Using remapping: %s", remapping)
		}
		solcOptions = append(solcOptions, solc.WithRemappings(remappings))
	}

	// Parse output selection from environment variable
	outputSelection, err := parseOutputSelection(cfg.OutputSelection)
	if err != nil {
		logger.Logger.Fatal().Msgf("Failed to parse output selection: %s", err)
		return nil
	}

	if len(outputSelection) > 0 {
		logger.Logger.Info().Msgf("Using output selection: %s", outputSelection)
		solcOptions = append(solcOptions, solc.WithOutputSelection(outputSelection))
	}

	// Include static options
	solcOptions = append([]solc.Option{
		solc.WithOptimizer(&solc.Optimizer{Enabled: cfg.SolcOptimizer, Runs: uint64(cfg.SolcOptimizerRuns)}),
		solc.WithEVMVersion(solc.EVMVersionPetersburg),
		solc.WithViaIR(cfg.SolcViaIR),
	}, solcOptions...)

	// Initialize the Solidity compiler
	c := solc.New(solc.Version(cfg.SolcVersion))

	for _, contractDir := range groups {
		for _, name := range contractMap[contractDir] {
			contracName := fmt.Sprintf("%s.sol", name)
			mainContractFile := filepath.Join(contractDir, contracName)

			if _, err := os.Stat(mainContractFile); os.IsNotExist(err) {
				logger.Logger.Fatal().Msgf("Contract file does not exist: %s", mainContractFile)
				return nil
			}

			logger.Logger.Info().Msgf("🚀 Compiling %s", contracName)

			// Compile the contract
			c, err := c.Compile(
				contractDir, name,
				solcOptions...,
			)
			if err != nil {
				logger.Logger.Fatal().Msgf("Compilation failed for %s: %s", contracName, err)
				return nil
			}

			contract := &types.Contract{
				ContractName: name,
				ABI:          c.ABI,
				Bytecode:     "0x" + hex.EncodeToString(c.Constructor),
			}

			// Validate the struct
			if err := validator.ValidateStruct(contract); err != nil {
				logger.Logger.Fatal().Msgf("Failed to validate contract: %s", err)
				return nil
			}

			// Write ABI
			output := filepath.Join(buildDir, name+".json")
			if err := directory.SaveToFile(output, contract); err != nil {
				return fmt.Errorf("Fail to write bin to file: %s", err)
			}

			logger.Logger.Info().Msgf("✅ Compiled contract in %s.json", name)
		}
	}

	// When Clean Mode is activated delete after
	if cfg.CleanMode {
		// Remove the .solc directory
		logger.Logger.Info().Msgf("✅ Removed solc directory")
		// Remove the .solc directory
		if err := directory.DeleteDir(".solc"); err != nil {
			logger.Logger.Fatal().Msgf("Failed to remove solc directory: %s", err)
			return nil
		}
	}

	return nil
}

// ParseRemappings parses the remappings env variable (e.g., "lib=/path1,mimc=/path2")
func parseRemappings(raw string) ([]string, error) {
	// If the raw string is empty, return an empty slice.
	if raw == "" {
		return []string{}, nil
	}

	// Parse the remappings string into a slice
	pairs := strings.Split(raw, ",")
	var remappings []string
	for _, pair := range pairs {
		if !strings.Contains(pair, "=") {
			return nil, fmt.Errorf("invalid remapping format: %s", pair)
		}
		remappings = append(remappings, pair)
	}
	return remappings, nil
}

// ParseOutputSelection parses the JSON output selection from env
func parseOutputSelection(raw string) (map[string]map[string][]string, error) {
	// If the raw string is empty, return an empty map.
	if raw == "" {
		return map[string]map[string][]string{}, nil
	}
	// Parse the JSON string into a map
	var selection map[string]map[string][]string
	if err := json.Unmarshal([]byte(raw), &selection); err != nil {
		return nil, fmt.Errorf("invalid output selection format: %w", err)
	}
	return selection, nil
}
