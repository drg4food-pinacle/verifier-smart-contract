package validator

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-playground/validator/v10"
)

func isEthAddress(fl validator.FieldLevel) bool {
	addr := fl.Field().Interface().(common.Address)
	return common.IsHexAddress(addr.Hex()) && addr != common.MaxAddress
}
