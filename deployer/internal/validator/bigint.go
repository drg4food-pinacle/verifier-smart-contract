package validator

import (
	"math/big"

	"github.com/go-playground/validator/v10"
)

var maxUint256 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)) // 2^256 - 1

func isBigint(fl validator.FieldLevel) bool {
	valIface := fl.Field().Interface()

	var val *big.Int
	switch v := valIface.(type) {
	case *big.Int:
		val = v
	case big.Int:
		val = &v
	default:
		return false
	}

	// nil check
	if val == nil {
		return false
	}

	// Must be >= 0 and <= 2^256 - 1
	if val.Sign() < 0 {
		return false
	}
	if val.Cmp(maxUint256) > 0 {
		return false
	}

	return true
}

func isBigintGte0(fl validator.FieldLevel) bool {
	switch val := fl.Field().Interface().(type) {
	case *big.Int:
		return val != nil && val.Sign() >= 0
	case big.Int:
		return val.Sign() >= 0
	default:
		return false
	}
}
