package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Account struct {
	PrivateKeyHex    string         `mapstructure:"privateKeyHex" validate:"required,len=64,hexadecimal"` // 64 hex chars (32 bytes)
	PrivateKeyBigInt *big.Int       `mapstructure:"privateKeyBigInt" validate:"required,bigint"`          // BigInt as string
	ChecksumAddress  common.Address `mapstructure:"checksumAddress" validate:"required,eth_addr"`         // Custom eth address format
	Nonce            *big.Int       `mapstructure:"nonce" validate:"required,bigint,bigint_gte_0"`        // Should be zero or positive
}

type Accounts struct {
	Name     string           `mapstructure:"name" validate:"required"`
	Accounts map[int]*Account `mapstructure:"accounts" validate:"required,dive"`
}

func (Account) CustomErrorMessages() map[string]string {
	return map[string]string{
		// Account Private Key
		"Account.PrivateKeyHex.required":    "Private key (hex) is required",
		"Account.PrivateKeyHex.len":         "Private key must be exactly 64 hexadecimal characters long",
		"Account.PrivateKeyHex.hexadecimal": "Private key must contain only valid hexadecimal characters",
		// Account Private Key BigInt
		"Account.PrivateKeyBigInt.required": "Private key (big integer) is required",
		"Account.PrivateKeyBigInt.bigint":   "Private key (big integer) must be a valid numeric string",
		// Account Address
		"Account.ChecksumAddress.required": "Checksum Ethereum address is required",
		"Account.ChecksumAddress.eth_addr": "Checksum Ethereum address is not valid",
		// Account Nonce
		"Account.Nonce.required":     "Nonce is required",
		"Account.Nonce.bigint":       "Nonce must be a valid number",
		"Account.Nonce.bigint_gte_0": "Nonce must be zero or a positive number",
	}
}

func (Accounts) CustomErrorMessages() map[string]string {
	return map[string]string{
		// Accounts wrapper
		"Accounts.Accounts.Name.required":       "Account group name is required",
		"Accounts.Accounts.Accounts.required":   "Account list cannot be empty",
		"Accounts.Accounts.Accounts[].required": "Each account entry must not be nil",
	}
}
