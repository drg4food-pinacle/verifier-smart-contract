package types

import (
	"github.com/ethereum/go-ethereum/common"
)

type Contract struct {
	Name    string         `json:"name" validate:"required"`
	Address common.Address `json:"address" validate:"required,eth_addr"`
}

//	type Contracts struct {
//		Contracts []*Contract `json:"contracts" validate:"required,dive,required"`
//	}
type Contracts struct {
	Contracts map[string]*Contract `json:"contracts" validate:"required,dive,keys,required,endkeys,required"`
}

func (Contract) CustomErrorMessages() map[string]string {
	return map[string]string{
		// Contract
		"Contract.Name.required":    "Contract name is required",
		"Contract.Address.required": "Contract address is required",
		"Contract.Address.eth_addr": "Invalid Ethereum address",
	}
}

func (Contracts) CustomErrorMessages() map[string]string {
	return map[string]string{
		// Contracts
		"Contracts.Contracts.required":           "Contracts list is required",
		"Contracts.Contracts[].required":         "Each contract must be non-nil",
		"Contracts.Contracts[].Address.required": "Contract address is required",
		"Contracts.Contracts[].Address.eth_addr": "Invalid Ethereum address",
		"Contracts.Contracts[].Name.required":    "Contract name is required",
	}
}
