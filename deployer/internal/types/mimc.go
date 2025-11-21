package types

import "github.com/consensys/gnark-crypto/ecc/bn254/fr"

// MiMCSponge represents the MiMC sponge hash function
type MiMCSponge struct {
	Constants []fr.Element `validate:"required,dive"`
	Seed      string       `validate:"required"`
	NumRounds int          `validate:"required,min=1,max=220"`
}

func (MiMCSponge) CustomErrorMessages() map[string]string {
	return map[string]string{
		"MiMCSponge.MiMCSponge.Constants[].required": "Constants are required",
		"MiMCSponge.MiMCSponge.Seed.required":        "Seed is required",
		"MiMCSponge.MiMCSponge.NumRounds.required":   "NumRounds is required",
		"MiMCSponge.MiMCSponge.NumRounds.min":        "NumRounds must be at least 1",
		"MiMCSponge.MiMCSponge.NumRounds.max":        "NumRounds must not exceed 220",
	}
}
