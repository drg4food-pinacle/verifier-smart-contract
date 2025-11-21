package main

import (
	"runtime"

	"go-contracts/internal/abigen"
	"go-contracts/internal/banner"
	"go-contracts/internal/compiler"
	"go-contracts/internal/config"
	"go-contracts/internal/logger"

	"github.com/spf13/cobra"
)

var (
	// Find the number of cpus the system has.
	maxProcs = runtime.NumCPU()

	// Configuration
	cfg *config.Config

	// Commands
	compileCMD = &cobra.Command{
		Use:   "compile",
		Short: "Compile Solidity contracts",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Logger.Info().Msg("🔨 Starting Solidity compilation...")
			return compiler.Compile(cfg)
		},
	}

	abigenCMD = &cobra.Command{
		Use:   "abigen",
		Short: "Denerate Go bindings for Solidity contracts",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Logger.Info().Msg("🔨 Starting Abigen binding generation...")
			return abigen.Generate(cfg)
		},
	}
)

func init() {
	// Set the max procs to the number of CPUs available
	runtime.GOMAXPROCS(maxProcs)

	// Initialize the config
	cfg = config.NewConfig()

	// Initialize the logger
	logger.SetupLogger(cfg.LoggerMode)
	logger.Logger.Info().Msg("Logger Initialized")
}

func main() {
	// Set GOMAXPROCS to the number of CPUs available
	logger.Logger.Info().Msgf("Setting GOMAXPROCS to %d", maxProcs)
	logger.Logger.Info().Msgf("Starting Rollup Server on %d CPU(s)", maxProcs)
	logger.Logger.Info().Msgf("Go Version: %s", runtime.Version())
	logger.Logger.Info().Msgf("OS: %s", runtime.GOOS)
	logger.Logger.Info().Msgf("Architecture: %s", runtime.GOARCH)

	// Load configuration
	err := cfg.LoadConfig()
	if err != nil {
		// Log the error and exit
		logger.Logger.Fatal().Err(err).Msg("Failed to load Config")
	}

	// Reload Logger if production mode is set
	if cfg.LoggerMode == "production" {
		logger.SetupLogger("production")
		logger.Logger.Info().Msg("Production logger Initialized")
	}

	logger.Logger.Info().Msg("Configuration loaded Successfully")

	// Print Banner
	if !cfg.DisableBanner {
		banner.PrintBanner(cfg.Version)
	}

	rootCmd := &cobra.Command{
		Use:   "contract-cli",
		Short: "Contract CLI for Solidity compilation and Go binding generation",
		Long: `contract-cli is a powerful tool for working with Solidity smart contracts.

It allows you to compile Solidity source files and generate Go bindings using abigen.
This is useful for integrating Ethereum smart contracts into your Go applications.

Examples:
  # Compile Solidity contracts
  contract-cli compile

  # Generate Go bindings from compiled contracts
  contract-cli abigen
`,
		Example: `
  contract-cli compile
  contract-cli abigen
`,
		SilenceUsage:  true, // Avoid showing usage on errors like "flag not found"
		SilenceErrors: true, // Avoid showing errors on command execution
	}
	rootCmd.AddCommand(compileCMD)
	rootCmd.AddCommand(abigenCMD)

	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		logger.Logger.Error().Msgf("Command failed: %v", err)
		logger.Logger.Error().Msgf("Printing Help...")
		_ = cmd.Help()
		return nil
	})

	if err := rootCmd.Execute(); err != nil {
		logger.Logger.Fatal().Msgf("Command failed: %v", err)
	}
}
