package types

import (
	"math/big"
)

const LEVELS = 32

type PinacleZKP struct {
	PrivateKey   *Registers       `json:"privateKey"`
	PathElements [LEVELS]*big.Int `json:"pathElements"`
	PathIndices  [LEVELS]*big.Int `json:"pathIndices"`
}
