package addresses

import (
	"deployer/internal/directory"
	"deployer/internal/types"
	"deployer/internal/validator"
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type Addresses struct {
	mu sync.RWMutex
	*types.Contracts
}

func NewAddresses() *Addresses {
	return &Addresses{
		Contracts: &types.Contracts{
			Contracts: make(map[string]*types.Contract),
		},
	}
}

// AddContract appends a new contract with its name and address to the list.
func (a *Addresses) AddContract(name string, address common.Address) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	contract := &types.Contract{
		Name:    name,
		Address: address,
	}

	// Validate the structs
	if err := validator.ValidateStruct(contract); err != nil {
		return fmt.Errorf("failed to validate contract: %w", err)
	}

	a.Contracts.Contracts[name] = contract

	// Validate the structs
	if err := validator.ValidateStruct(a.Contracts); err != nil {
		return fmt.Errorf("failed to validate contracts: %w", err)
	}
	return nil
}

// SaveToFile writes the Addresses' contract information to the specified file path.
// It uses the directory.SaveToFile function to perform the actual file writing.
// Returns an error if the operation fails, including the file path and the underlying error.
func (a *Addresses) SaveToFile(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	err := directory.SaveToFile(path, a.Contracts)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", path, err)
	}
	return nil
}

// LoadFromFile loads account data from a JSON file into an Accounts instance.
func (a *Addresses) LoadFromFile(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	err := directory.LoadFromFile(path, a)
	if err != nil {
		return fmt.Errorf("failed load addresses from file: %w", err)
	}
	return nil
}

// GetContractAddressByName retrieves a contract address by its name.
func (a *Addresses) GetContractAddressByName(name string) (common.Address, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.Contracts == nil || a.Contracts.Contracts == nil {
		return common.Address{}, errors.New("no contracts loaded")
	}

	contract, ok := a.Contracts.Contracts[name]
	if !ok || contract == nil {
		return common.Address{}, errors.New("contract not found: " + name)
	}

	return contract.Address, nil
}
