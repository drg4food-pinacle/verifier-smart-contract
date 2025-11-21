package abigen

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-contracts/internal/config"
	"go-contracts/internal/directory"
	"go-contracts/internal/logger"
	"go-contracts/internal/types"
	"go-contracts/internal/validator"

	"io/fs"
	"os"
	"path/filepath"
)

func Generate(cfg *config.Config) error {
	base := cfg.ContractsBaseDir
	if err := directory.CheckDirExists(base); err != nil {
		logger.Logger.Fatal().Msgf("Base directory does not exist: %s", base)
		return nil
	}

	bindingsDir := cfg.AbigenOutputDir
	// Check if the bindings directory exists
	err := directory.DeleteDir(bindingsDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.Logger.Fatal().Msgf("Failed to delete bindings directory: %s", err)
		return nil
	}

	// Create the build directory if it doesn't exist
	if err := directory.CreateDirIfNotExists(bindingsDir); err != nil {
		logger.Logger.Fatal().Msgf("Failed to create bindings directory: %s", err)
		return nil
	}

	// Check solc output directory exists
	solcOut := cfg.SolcOutputDir
	if err := directory.CheckDirExists(solcOut); err != nil {
		logger.Logger.Fatal().Msgf("Solc Output directory does not exists: %s", err)
		return nil
	}

	// Walk through the solc output directory and find JSON files
	var jsonFiles []string
	err = filepath.WalkDir(solcOut, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // propagate error
		}
		if !d.IsDir() && filepath.Ext(d.Name()) == ".json" {
			jsonFiles = append(jsonFiles, path)
		}
		return nil
	})
	if err != nil {
		logger.Logger.Fatal().Msgf("Failed to walk to contracts bin directory: %s", err)
		return nil
	}

	logger.Logger.Info().Msgf("✅ Found %d JSON compiled files", len(jsonFiles))

	// Initialize the Abigen compiler
	abigen := New(Version(cfg.AbigenVersion))

	// Init once
	abigen.once.Do(abigen.init)
	if abigen.err != nil {
		logger.Logger.Fatal().Msgf("Failed initialize Abigen package: %s", err)
		return nil
	}

	for _, jsonPath := range jsonFiles {
		logger.Logger.Info().Msgf("✅ Found %s JSON compiled file", filepath.Base(jsonPath))

		// Read the JSON file
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			logger.Logger.Fatal().Msgf("Failed read JSON file %s: %s", jsonPath, err)
			return nil
		}

		// Parse JSON
		var contractJSON *types.Contract
		if err := json.Unmarshal(data, &contractJSON); err != nil {
			logger.Logger.Fatal().Msgf("Failed read JSON file %s: %s", jsonPath, err)
			return nil
		}

		// Validate the struct
		if err := validator.ValidateStruct(contractJSON); err != nil {
			logger.Logger.Fatal().Msgf("Failed to validate contract %s: %s", jsonPath, err)
			return nil
		}

		var opts []Option
		outputDir := filepath.Join(cfg.AbigenOutputDir, contractJSON.ContractName)
		if err := directory.CreateDirIfNotExists(outputDir); err != nil {
			logger.Logger.Fatal().Msgf("Failed to create directory %s: %s", outputDir, err)
			return nil
		}
		output := filepath.Join(outputDir, fmt.Sprintf("%s.go", contractJSON.ContractName))
		opts = append(opts, WithOutput(output), WithPackage(contractJSON.ContractName))
		err = abigen.Generate(jsonPath, opts...)
		if err != nil {
			logger.Logger.Fatal().Msgf("Failed to Execute Abigen binary: %s", err)
			return nil
		}
	}

	// When Clean Mode is activated delete after
	if cfg.CleanMode {
		// Remove the .abigen directory
		logger.Logger.Info().Msgf("✅ Removed abigen directory")
		// Remove the .abigen directory
		if err := directory.DeleteDir(".abigen"); err != nil {
			logger.Logger.Fatal().Msgf("Failed to remove abigen directory: %s", err)
			return nil
		}
	}

	return nil
}
