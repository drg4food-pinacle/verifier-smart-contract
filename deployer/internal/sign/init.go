package sign

import (
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

var (
	// SECP256K1_N is the order of the secp256k1 curve (N)
	SECP256K1_N *big.Int

	// HalfOrder is SECP256K1_N / 2, used in canonical signature checks (to ensure low-S values)
	HalfOrder *big.Int

	max *big.Int
)

func init() {
	const maxUint64 = ^uint64(0)
	max = new(big.Int).SetUint64(maxUint64)

	SECP256K1_N = crypto.S256().Params().N
	HalfOrder = new(big.Int).Rsh(SECP256K1_N, 1)
}
