package types

import "math/big"

const (
	REGISTERS   = 4
	STRIDE      = 8
	NUM_STRIDES = 256 / STRIDE // = 32
)

type T [NUM_STRIDES][1 << STRIDE][2]Registers
type U [2]Registers
type Registers [REGISTERS]*big.Int

type Signature struct {
	S *Registers `json:"s" validate:"required,dive,required"`
	R *Registers `json:"r" validate:"required,dive,required"`
	V uint8      `json:"v" validate:"required,oneof=27 28"`
}

type Precomputes struct {
	Signature *Signature `json:"signature" validate:"required,dive"`
	T         *T         `json:"t" validate:"required,dive"` // Keep required if you want non-nil pointer
	U         *U         `json:"u" validate:"required,dive"` // dive for inner validation
}

func (Precomputes) CustomErrorMessages() map[string]string {
	return map[string]string{
		"Signature.S.required": "Field 's' is required and must be a valid register set",
		"Signature.R.required": "Field 'r' is required and must be a valid register set",
		"Signature.V.required": "Field 'v' is required",
		"Signature.T.required": "Field 't' is required and must contain precomputed values",
		"Signature.U.required": "Field 'u' is required and must contain u values",
	}
}
