package ethutil

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
)

// FindPrivateKey searches the given directory for the first Ethereum
// keystore file whose name starts with "UTC--", which is the standard prefix
// for geth keystore filenames.
//
// Parameters:
//   - dir: the path to the directory containing keystore files.
//
// Returns:
//   - string: the full path to the first matching keystore file.
//   - error:  an error if the directory walk fails or no keyfile is found.
func FindPrivateKey(dir string) (string, error) {
	var found string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasPrefix(filepath.Base(path), "UTC--") {
			found = path
			return filepath.SkipDir // Stop the walk once the first match is found
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("walk keystore: %w", err)
	}
	if found == "" {
		return "", errors.New("no UTC-- keyfile found in keystore")
	}
	return found, nil
}

// DecryptKeyfile reads an Ethereum keystore file from the given path,
// decrypts it using the provided password, and returns the ECDSA private key.
//
// Parameters:
//   - path:     the full path to the keystore file (e.g. "UTC--2022...").
//   - password: the password used to decrypt the keystore file.
//
// Returns:
//   - *ecdsa.PrivateKey: the extracted private key if successful.
//   - error:             an error if reading or decryption fails.
func DecryptKeyfile(path, password string) (*ecdsa.PrivateKey, error) {
	keyJSON, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read keyfile: %w", err)
	}
	key, err := keystore.DecryptKey(keyJSON, password)
	if err != nil {
		return nil, fmt.Errorf("decrypt keyfile: %w", err)
	}
	return key.PrivateKey, nil
}

// NewTransactorFromKeystore creates a new transaction options object using the provided
// ECDSA private key and chain ID. This function is a wrapper around the
// bind.NewKeyedTransactorWithChainID function, which generates a TransactOpts
// instance for signing Ethereum transactions.
//
// Parameters:
//   - privateKey: The ECDSA private key used to sign transactions.
//   - chainId: The chain ID of the Ethereum network to ensure replay protection.
//
// Returns:
//   - *bind.TransactOpts: A transaction options object configured with the provided
//     private key and chain ID.
//   - error: An error if the transaction options could not be created.
func NewTransactorFromKeystore(privateKey *ecdsa.PrivateKey, chainId *big.Int) (*bind.TransactOpts, error) {
	return bind.NewKeyedTransactorWithChainID(privateKey, chainId)
}
