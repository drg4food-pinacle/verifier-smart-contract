package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"time"

	verifier "deployer/internal/abigen/Verifier"
	"deployer/internal/abigen/mimc"
	zklogin "deployer/internal/abigen/zkLogin"
	"deployer/internal/accounts"
	"deployer/internal/addresses"
	"deployer/internal/banner"
	"deployer/internal/config"
	"deployer/internal/directory"
	"deployer/internal/ethutil"
	"deployer/internal/logger"
	mimcsponge "deployer/internal/mimc"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

var (
	// Find the number of cpus the system has.
	maxProcs = runtime.NumCPU()
	// Keep track of the last processed block from events
	LatestProcessedBlockNumber atomic.Uint64
)

func main() {
	runtime.GOMAXPROCS(maxProcs)

	// Initialize config first
	cfg := config.NewConfig()

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

	// Delete Old Accounts
	if err := directory.DeleteDir(cfg.AccountsDir); err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.Logger.Fatal().Err(err).Msgf("Failed to delete directory")
	}
	if err := directory.CreateDirIfNotExists(cfg.AccountsDir); err != nil {
		logger.Logger.Fatal().Err(err).Msgf("Failed to create directory")
	}

	// Initialize Mimc
	// Mimc is a global variable, so we can use it directly
	mimcsponge, err := mimcsponge.NewMiMCSponge(mimcsponge.Seed, mimcsponge.MimcNbRounds)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to initialize MiMC Sponge")
	}

	base := cfg.AccountsDir

	// Election Leaders
	foodbanksFilename := "foodBanks"
	foodbanksPath := filepath.Join(base, fmt.Sprintf("%s.json", foodbanksFilename))
	// Create new Accounts object
	foodbanks := accounts.NewAccounts(foodbanksFilename)
	foodbanks.SetMiMC(mimcsponge) // Set MiMC for hashing addresses
	// Create Accounts
	foodbanks.CreateAccounts(cfg.AccountsNumber)
	// Save to file
	foodbanks.SaveToFile(foodbanksPath)

	// Find the private Key and unlock it
	keyfile, err := ethutil.FindPrivateKey(cfg.GethNodeKeystore)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to find Private key in Keystore")
	}

	// Decrypt the encrypted private key
	privateKey, err := ethutil.DecryptKeyfile(keyfile, cfg.GethNodePassword)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to decrypt Private key in Keystore")
	}

	// Initialize Addresses struct
	contractAddresses := addresses.NewAddresses()

	// Create a ctx context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	client, chainId, err := ethutil.NewEthClient(ctx, cfg.GethNodeUrl)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to connect to Ethereum node")
	}
	// Create a new EthClient
	ethclient := client.EthClient
	defer client.Close()

	// Create a new Transactor
	trOpts, err := ethutil.NewTransactorFromKeystore(privateKey, chainId)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to create a new Transactor")
	}
	trOpts.Context = ctx
	trOpts.Nonce = nil
	trOpts.GasLimit = 20_000_000
	trOpts.GasPrice = big.NewInt(0)

	// Deploy Mimc
	mimcAddress, txMimc, _, err := mimc.DeployMimc(trOpts, ethclient)
	if err != nil {
		logger.Logger.Fatal().Err(err).Str("contract", "Mimc").Msgf("Failed to deploy contract")
	}
	_, err = bind.WaitMined(ctx, ethclient, txMimc)
	if err != nil {
		logger.Logger.Fatal().Err(err).Str("contract", "Mimc").Msgf("Failed to mine tx")
	}

	logger.Logger.Info().Str("address", mimcAddress.Hex()).Msg("Mimc")
	contractAddresses.AddContract("mimc", mimcAddress) // Add the address to the contractAddresses object

	// Deploy Verifier
	verifierAddress, txVerifier, _, err := verifier.DeployVerifier(trOpts, ethclient)
	if err != nil {
		logger.Logger.Fatal().Err(err).Str("contract", "Verifier").Msgf("Failed to deploy contract")
	}
	_, err = bind.WaitMined(ctx, ethclient, txVerifier)
	if err != nil {
		logger.Logger.Fatal().Err(err).Str("contract", "Verifier").Msgf("Failed to mine tx")
	}

	logger.Logger.Info().Str("address", verifierAddress.Hex()).Msg("Verifier")
	contractAddresses.AddContract("verifier", verifierAddress) // Add the address to the contractAddresses object

	// Get Foodbank addresses
	fb := derefAddresses(foodbanks.ExtractAddresses())

	// Deploy ZkLogin
	zkLoginAddress, txZkLogin, _, err := zklogin.DeployZklogin(trOpts, ethclient, 2, []uint32{1, 1}, []uint32{32, 32}, mimcAddress, verifierAddress, fb)
	if err != nil {
		logger.Logger.Fatal().Err(err).Str("contract", "ZkLogin").Msgf("Failed to deploy contract")
	}
	_, err = bind.WaitMined(ctx, ethclient, txZkLogin)
	if err != nil {
		logger.Logger.Fatal().Err(err).Str("contract", "ZkLogin").Msgf("Failed to mine tx")
	}
	logger.Logger.Info().Str("address", zkLoginAddress.Hex()).Msg("ZkLogin")
	contractAddresses.AddContract("zklogin", zkLoginAddress) // Add the address to the contractAddresses object

	// Write the contract Address to file
	addressesPath := filepath.Join(cfg.AddressesDir, "addresses.json")
	contractAddresses.SaveToFile(addressesPath)
	logger.Logger.Info().Msg("Deployer finished successfully")
}

func derefAddresses(ptrs []*common.Address) []common.Address {
	addrs := make([]common.Address, 0, len(ptrs))
	for _, ptr := range ptrs {
		if ptr != nil {
			addrs = append(addrs, *ptr)
		}
	}
	return addrs
}
