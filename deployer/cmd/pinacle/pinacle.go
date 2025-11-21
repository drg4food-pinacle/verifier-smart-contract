package main

// ⚠️ WARNING: The "input" field is not compatible with older versions of Geth.
// Always use "data" in eth_call requests to maintain backward compatibility.
// Use the internal modified package not the official

import (
	"context"
	zklogin "deployer/internal/abigen/zkLogin"
	"deployer/internal/accounts"
	"deployer/internal/addresses"
	"deployer/internal/banner"
	"deployer/internal/config"
	"deployer/internal/ethutil"
	"deployer/internal/logger"
	"deployer/internal/mimc"
	"deployer/internal/sign"
	"deployer/internal/types"
	"deployer/internal/zkp"
	"fmt"
	"math/big"
	"path/filepath"
	"runtime"
	"sync/atomic"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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

	// Initialize Mimc
	// Mimc is a global variable, so we can use it directly
	mimc, err := mimc.NewMiMCSponge(mimc.Seed, mimc.MimcNbRounds)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to initialize MiMC Sponge")
	}

	// Load Accounts
	accountsBase := cfg.AccountsDir

	// Election Leaders
	foodbanksFilename := "foodBanks"
	foodbanksPath := filepath.Join(accountsBase, fmt.Sprintf("%s.json", foodbanksFilename))
	// Create new Accounts object
	foodbanks := accounts.NewAccounts(foodbanksFilename)
	foodbanks.SetMiMC(mimc) // Set MiMC for hashing addresses
	foodbanks.LoadFromFile(foodbanksPath)

	// Load Addresses
	contractAddressessBase := cfg.AddressesDir

	// Initialize Addresses struct
	contractAddressesFilename := "addresses"
	contractAddressesPath := filepath.Join(contractAddressessBase, fmt.Sprintf("%s.json", contractAddressesFilename))
	// Create new Address object
	contractAddresses := addresses.NewAddresses()
	contractAddresses.LoadFromFile(contractAddressesPath)

	// Load Wasm and Zkey into a prover object
	prover, err := zkp.NewProver(zkp.Path(cfg.WasmFilename), zkp.Path(cfg.ZkeyFilename))
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to initialize ZKP prover")
	}

	// verifier, err := zkp.NewVerifier(zkp.Path(cfg.VerificationKeyFilename))
	// if err != nil {
	// 	logger.Logger.Fatal().Err(err).Msg("Failed to initialize ZKP verifier")
	// }

	// Create a ctx context with timeout
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to EthClient
	client, chainId, err := ethutil.NewEthClient(ctx, cfg.GethNodeUrl)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to connect to Ethereum node")
	}

	// Get MerkleTreeWithHistory contract address
	zkLoginAddress, err := contractAddresses.GetContractAddressByName("zklogin")
	if err != nil {
		logger.Logger.Fatal().Err(err).Str("contract", "ZkLogin").Msg("Failed to get contract address")
	}

	zkloginInstance, err := zklogin.NewZklogin(zkLoginAddress, client.EthClient)
	if err != nil {
		logger.Logger.Fatal().Err(err).Str("contract", "zkLogin").Msg("Failed to conncect to contract")
	}

	// ! FoodbBank
	foodbanksIndex := 0
	// Get Private key Hex at index
	foodbanksPrivateKeyHex, err := foodbanks.GetPrivateKey(foodbanksIndex)
	if err != nil {
		logger.Logger.Fatal().Err(err).Uint32("treeId", uint32(types.RoleFoodBank)).Msg("Failed to fetch private key from accounts")
	}

	// Load Private Key to wallet
	foodbanksPrivateKey := sign.NewECDSA()
	err = foodbanksPrivateKey.LoadPrivateKeyFromHex(foodbanksPrivateKeyHex)
	if err != nil {
		logger.Logger.Fatal().Err(err).Uint32("treeId", uint32(types.RoleFoodBank)).Msg("Failed to load private key")
	}

	// Create a new Transactor
	trOpts, err := ethutil.NewTransactorFromKeystore(foodbanksPrivateKey.GetPrivateKey(), chainId)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to create a new Transactor")
	}

	trOpts.Context = ctx
	trOpts.GasPrice = big.NewInt(0)
	trOpts.Nonce = nil
	trOpts.GasLimit = 47_000_000

	// Create Hex Key to BigInt
	foodbanksPrivateKeyBigInt, _ := new(big.Int).SetString(foodbanksPrivateKeyHex, 16)
	// Create ZK Ethereum Address Input
	input := zkp.NewZKP()
	input.SetPrivateKey(sign.BigIntToRegisters(foodbanksPrivateKeyBigInt))
	// Convert to string and marshal
	inputJSON, err := input.MarshalJSON()
	if err != nil {
		logger.Logger.Fatal().Err(err).Uint32("treeId", uint32(types.RoleFoodBank)).Uint8("zkp", uint8(types.ZKEthereumAddress)).Msg("Failed to convert zkp input to bytes")
	}

	// Generate Proofs and Public Signals (zk Ethereum Address)
	proofs, err := prover.GenerateProofs(inputJSON)
	if err != nil {
		logger.Logger.Fatal().Err(err).Uint32("treeId", uint32(types.RoleFoodBank)).Uint8("zkp", uint8(types.ZKEthereumAddress)).Msg("Failed to generate zk proofs")
	}

	proofsConverted, err := proofs.ConvertProof()
	if err != nil {
		logger.Logger.Fatal().Err(err).Uint32("treeId", uint32(types.RoleFoodBank)).Uint8("zkp", uint8(types.ZKEthereumAddress)).Msg("Failed to convert zk proofs")
	}

	publicSignalsConverted, err := proofs.ConvertPublicSignals()
	if err != nil {
		logger.Logger.Fatal().Err(err).Uint32("treeId", uint32(types.RoleFoodBank)).Uint8("zkp", uint8(types.ZKEthereumAddress)).Msg("Failed to convert zk public signals")
	}

	foodbanksMerkleProofs, err := zkloginInstance.FetchFoodBankMerkleProofs(&bind.CallOpts{From: trOpts.From, Context: ctx}, *proofsConverted, publicSignalsConverted)
	if err != nil {
		logger.Logger.Fatal().Err(err).Uint32("treeId", uint32(types.RoleFoodBank)).Str("call", "fetchFoodBankMerkleProofs").Msg("Failed to call function")
	}

	// Create ZK Merkle Tree
	input = zkp.NewZKP()
	input.SetPrivateKey(sign.BigIntToRegisters(foodbanksPrivateKeyBigInt))
	input.SetPathElement([32]*big.Int(foodbanksMerkleProofs.PathElements))
	input.SetPathIndices([32]*big.Int(foodbanksMerkleProofs.PathIndices))
	// Convert to string and marshal
	inputJSON, err = input.MarshalJSON()
	if err != nil {
		logger.Logger.Fatal().Err(err).Uint32("treeId", uint32(types.RoleFoodBank)).Uint8("zkp", uint8(types.ZKMerkleTree)).Msg("Failed to convert zkp input to bytes")
	}

	// Generate Proofs and Public Signals (zk Ethereum Address)
	proofs, err = prover.GenerateProofs(inputJSON)
	if err != nil {
		logger.Logger.Fatal().Err(err).Uint32("treeId", uint32(types.RoleFoodBank)).Uint8("zkp", uint8(types.ZKMerkleTree)).Msg("Failed to generate zk proofs")
	}

	proofsConverted, err = proofs.ConvertProof()
	if err != nil {
		logger.Logger.Fatal().Err(err).Uint32("treeId", uint32(types.RoleFoodBank)).Uint8("zkp", uint8(types.ZKMerkleTree)).Msg("Failed to convert zk proofs")
	}

	publicSignalsConverted, err = proofs.ConvertPublicSignals()
	if err != nil {
		logger.Logger.Fatal().Err(err).Uint32("treeId", uint32(types.RoleFoodBank)).Uint8("zkp", uint8(types.ZKMerkleTree)).Msg("Failed to convert zk public signals")
	}

	// ! THESE ARE THE ZK MERKLE TREE PROOFS
	// ! THESE SHOULD BE KEPT IN MEMORY
	spew.Dump(proofsConverted)
	spew.Dump(publicSignalsConverted)
	logger.Logger.Info().Msg("ZK Proofs generated successfully")
}
