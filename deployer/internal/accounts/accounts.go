package accounts

import (
	"deployer/internal/directory"
	"deployer/internal/mimc"
	"deployer/internal/types"
	"deployer/internal/validator"
	"encoding/hex"
	"sync"

	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Accounts struct {
	mu sync.RWMutex
	*mimc.MiMCSponge
	*types.Accounts
}

// NewAccounts initializes and returns an Accounts struct with a name and empty map of accounts.
func NewAccounts(name string) *Accounts {
	return &Accounts{
		Accounts: &types.Accounts{
			Name:     name,
			Accounts: make(map[int]*types.Account),
		},
	}
}

// SetMiMC sets the MiMC pointer
func (a *Accounts) SetMiMC(mimc *mimc.MiMCSponge) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	a.setMiMC(mimc)
}

// CreateAccounts generates `number` Ethereum accounts and writes them to a JSON file.
func (a *Accounts) CreateAccounts(number int) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if number <= 0 {
		return fmt.Errorf("number must be greater than zero")
	}

	for i := 0; i < number; i++ {
		// Generate a random private key
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			return fmt.Errorf("generate private key: %w", err)
		}

		privateKeyBytes := crypto.FromECDSA(privateKey)            // PrivateKey to Bytes
		privateKeyHex := hex.EncodeToString(privateKeyBytes)       // PrivateKey Hex
		privateKeyBigInt := new(big.Int).SetBytes(privateKeyBytes) // PrivateKey BigInt

		// Derive address from the private key
		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		// New Account
		err = a.addAccount(i, privateKeyHex, privateKeyBigInt, big.NewInt(0), address)
		if err != nil {
			return fmt.Errorf("failed to add account: %s", err)
		}
	}

	// Validate Struct
	if err := validator.ValidateStruct(a); err != nil {
		return fmt.Errorf("failed to validate accounts struct: %s", err)
	}
	return nil
}

// ExtractAddresses returns a slice of all account checksum addresses.
// It acquires a lock for thread safety and delegates to the internal extractAddresses method.
func (a *Accounts) ExtractAddresses() []*common.Address {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.extractAddresses()
}

// ExtractHashedAddresses returns a slice of hashed addresses for all accounts.
// It uses the MiMC hash function to hash each account's address (as a big.Int),
// producing a unique *big.Int for each address. The function ensures thread safety
// and returns a copy of each hash to avoid pointer reuse issues.
func (a *Accounts) ExtractHashedAddresses() []*big.Int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.extractHashedAddresses()
}

// GetPrivateKey returns the private key (in hex format) of the account at the specified index.
// It checks if the index is within bounds and if the account exists.
// Returns the private key as a string, or an error if the index is invalid or the account is nil.
func (a *Accounts) GetPrivateKey(index int) (string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	account, ok := a.getAccount(index)
	if !ok || account == nil {
		return "", fmt.Errorf("account not found for key: %d", index)
	}

	return account.PrivateKeyHex, nil
}

// GetHashedAddress returns the MiMC hash of the checksum address for the account at the specified index.
// It acquires a read lock to ensure thread-safe access to the accounts data.
// If the account does not exist, has no checksum address, or the MiMC hasher is not set, an error is returned.
// On success, it returns the hashed address as a *big.Int.
func (a *Accounts) GetHashedAddress(index int) (*big.Int, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	account, ok := a.getAccount(index)
	if !ok || account == nil {
		return nil, fmt.Errorf("account not found for key: %d", index)
	}
	if account.ChecksumAddress == (common.Address{}) {
		return nil, fmt.Errorf("account at index %d has no checksum address", index)
	}
	// Hash the address using MiMC
	if a.getMiMC() == nil {
		return nil, fmt.Errorf("mimc not set")
	}
	return a.hashAddress(&account.ChecksumAddress)
}

// SaveToFile saves the Accounts data to the specified file path.
// It serializes the Accounts instance and writes it to the file.
// Returns an error if the operation fails.
func (a *Accounts) SaveToFile(path string) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Save the Accounts data to the specified file
	err := directory.SaveToFile(path, a.Accounts)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", path, err)
	}
	return nil
}

// LoadFromFile loads account data from a JSON file into an Accounts instance.
func (a *Accounts) LoadFromFile(path string) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	err := directory.LoadFromFile(path, a.Accounts)
	if err != nil {
		return fmt.Errorf("failed load accounts from file: %w", err)
	}
	return nil
}

// extractAddresses returns a slice of all non-nil account checksum addresses
// from the Accounts struct. It acquires a read lock to ensure thread-safe
// access to the underlying accounts data.
func (a *Accounts) extractAddresses() []*common.Address {
	addresses := make([]*common.Address, 0, len(a.getAccounts().Accounts))
	for _, account := range a.getAccounts().Accounts {
		if account != nil {
			addresses = append(addresses, &account.ChecksumAddress)
		}
	}
	return addresses
}

func (a *Accounts) extractHashedAddresses() []*big.Int {
	hashedAddresses := make([]*big.Int, 0, len(a.getAccounts().Accounts))

	for _, account := range a.extractAddresses() {
		if account == nil {
			continue // Skip nil addresses
		}
		// Hash the address using MiMC
		accountBigInt, _ := a.hashAddress(account)

		// Append a **copy** of resultBigInt (because it's reused)
		hashedAddresses = append(hashedAddresses, accountBigInt)
	}

	return hashedAddresses
}

// AddAccount validates and appends a new account.
func (a *Accounts) addAccount(index int, privateKeyHex string, privateKeyBigInt, nonce *big.Int, address common.Address) error {
	account := &types.Account{
		PrivateKeyHex:    privateKeyHex,
		PrivateKeyBigInt: privateKeyBigInt,
		ChecksumAddress:  address,
		Nonce:            nonce,
	}

	// Validate the struct
	if err := validator.ValidateStruct(account); err != nil {
		return fmt.Errorf("failed to validate account: %w", err)
	}

	a.setAccount(index, account)

	// Validate the structs
	if err := validator.ValidateStruct(a.getAccounts()); err != nil {
		return fmt.Errorf("failed to validate accounts: %w", err)
	}

	return nil
}

// getAccount returns the account at the given index if it exists, else returns nil and false.
func (a *Accounts) getAccount(index int) (*types.Account, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	account, ok := a.getAccounts().Accounts[index]
	return account, ok
}

// getAccounts safely returns the Accounts field of the Accounts struct.
// It acquires a read lock to ensure concurrent access is handled correctly.
func (a *Accounts) getAccounts() *types.Accounts {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.Accounts
}

// hashAddress hashes the given Ethereum address using the MiMC hash function.
func (a *Accounts) hashAddress(address *common.Address) (*big.Int, error) {
	if a.getMiMC() == nil {
		return big.NewInt(0), fmt.Errorf("mimc not set")
	}
	// Use MiMC to hash the address
	return a.getMiMC().HashAddress(address), nil
}

// SetMimc sets the MiMCSponge pointer
func (a *Accounts) setAccount(index int, account *types.Account) {
	a.Accounts.Accounts[index] = account
}

// SetMimc sets the MiMCSponge pointer
func (a *Accounts) setMiMC(mimc *mimc.MiMCSponge) {
	a.MiMCSponge = mimc
}

// GetMiMC returns the mimc (read-only)
func (a *Accounts) getMiMC() *mimc.MiMCSponge {
	return a.MiMCSponge
}
