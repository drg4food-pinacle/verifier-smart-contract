package zkp

import (
	"deployer/internal/types"
	"math/big"
)

func newEmptyRegisters() *types.Registers {
	var r types.Registers
	for i := 0; i < types.REGISTERS; i++ {
		r[i] = big.NewInt(0)
	}
	return &r
}

func zeroArray() [LEVELS]*big.Int {
	var s [LEVELS]*big.Int
	for i := range s {
		s[i] = big.NewInt(0)
	}
	return s
}

// convert Registers -> []string
func registersToStrings(r *types.Registers) []string {
	strs := make([]string, len(r))
	for i, b := range r {
		if b == nil {
			strs[i] = "0"
		} else {
			strs[i] = b.String()
		}
	}
	return strs
}

// convert T -> [][][][]string
func tToStrings(t *types.T) [][][][]string {
	out := make([][][][]string, len(t))
	for i := range t {
		out[i] = make([][][]string, len(t[i]))
		for j := range t[i] {
			out[i][j] = make([][]string, len(t[i][j]))
			for k := range t[i][j] {
				out[i][j][k] = registersToStrings(&t[i][j][k])
			}
		}
	}
	return out
}

func uToStrings(u *types.U) [][]string {
	out := make([][]string, len(u))
	for i := range u {
		out[i] = make([]string, len(u[i]))
		for j := range u[i] {
			out[i][j] = u[i][j].String()
		}
	}
	return out
}
