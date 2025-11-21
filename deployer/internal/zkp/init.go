package zkp

import (
	"deployer/internal/types"
	"math/big"
)

type Path string

const LEVELS = types.LEVELS
const VOTING_PUBLIC_SIGNALS = types.PINACLE_PUBLIC_SIGNALS

var emptyRegisters *types.Registers
var zeroArr [LEVELS]*big.Int

func init() {
	emptyRegisters = newEmptyRegisters()
	zeroArr = zeroArray()
}
