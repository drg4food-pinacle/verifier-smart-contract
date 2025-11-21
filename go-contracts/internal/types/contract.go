package types

import "encoding/json"

type Contract struct {
	ContractName string            `mapstructure:"contractName" validate:"required"`
	ABI          []json.RawMessage `mapstructure:"abi" validate:"required,min=1,dive,required"` // Ensure ABI has at least 1 element
	Bytecode     string            `mapstructure:"bytecode" validate:"required,hexadecimal"`
}

func (c Contract) CustomErrorMessages() map[string]string {
	return map[string]string{
		"Contract.ContractName.required": "Contract name is required",
		// ABI
		"Contract.ABI.required":   "ABI is required",
		"Contract.ABI.min":        "ABI must contain at least one element",
		"Contract.ABI[].required": "Each ABI entry must be non-empty",
		// Bytecode
		"Contract.Bytecode.required":    "Contract bytecode is required",
		"Contract.Bytecode.hexadecimal": "Contract bytecode must be a valid hexadecimal string",
	}
}
