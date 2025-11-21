package types

import "math/big"

type ZKPType uint8

const (
	ZKEthereumAddress ZKPType = 0
	ZKMerkleTree      ZKPType = 1
)

const PINACLE_PUBLIC_SIGNALS = 2

type Groth16Proof struct {
	PiA [2]*big.Int
	PiB [2][2]*big.Int
	PiC [2]*big.Int
}
