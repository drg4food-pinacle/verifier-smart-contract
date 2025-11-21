package sign

import (
	"deployer/internal/types"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec/v2"
)

const (
	REGISTERS   = types.REGISTERS
	STRIDE      = types.STRIDE
	NUM_STRIDES = types.NUM_STRIDES
)

var (
	ErrXMustBe32Bytes    = errors.New("x must be exactly 32 bytes")
	ErrRNoModularInverse = errors.New("r has no modular inverse")
)

var zero = new(big.Int).SetInt64(0)

func pointPrecomputes(P *btcec.PublicKey) *types.T {
	var gPowers types.T // concrete array value, not pointer

	power := new(big.Int)
	l := new(big.Int)

	for i := 0; i < NUM_STRIDES; i++ {
		power.Lsh(big.NewInt(1), uint(i*STRIDE)) // 2^(i * STRIDE)

		for j := 0; j < 1<<STRIDE; j++ {
			l.Mul(big.NewInt(int64(j)), power)
			x, y := btcec.S256().ScalarMult(P.X(), P.Y(), l.Bytes())

			xRegs := BigIntToRegisters(x)
			yRegs := BigIntToRegisters(y)

			for r := 0; r < REGISTERS; r++ {
				gPowers[i][j][0][r] = xRegs[r]
				gPowers[i][j][1][r] = yRegs[r]
			}
		}
	}
	return &gPowers
}

func pointU(Ux, Uy *big.Int) *types.U {
	return &types.U{
		*BigIntToRegisters(Ux),
		*BigIntToRegisters(Uy),
	}
}

// splitToRegisters splits a hex string or big.Int into four 64-bit registers (big.Int values as strings).
func BigIntToRegisters(value *big.Int) *types.Registers {
	var registers types.Registers // concrete array value, not pointer

	if value == nil || value.Sign() == 0 {
		return &types.Registers{zero, zero, zero, zero}
	}

	// pad to 64 hex characters (256 bits)
	hex := fmt.Sprintf("%064x", value)

	for k := 0; k < REGISTERS; k++ {
		start := k * 16
		end := (k + 1) * 16
		p := hex[start:end]
		partInt := new(big.Int)
		partInt.SetString(p, 16)

		// Insert in reverse order like unshift in JS
		registers[REGISTERS-1-k] = partInt
	}

	return &registers
}

// pointFromX returns the public key point given X and Y's parity bit (isYOdd)
// RecoverYFromX recovers the Y coordinate on secp256k1 given X and the isYOdd parity bit
func pointFromX(xHex string, isOdd bool) (*btcec.PublicKey, error) {
	xBytes, err := hex.DecodeString(xHex)
	if err != nil {
		return nil, err
	}
	if len(xBytes) != 32 {
		return nil, ErrXMustBe32Bytes
	}

	var prefix byte = 0x02 // even
	if isOdd {
		prefix = 0x03 // odd
	}

	compressed := append([]byte{prefix}, xBytes...) // 33 bytes
	return btcec.ParsePubKey(compressed)
}

// Get r point from x coordinate and y parity (like curve.pointFromX)
func rInv(rHex string) (*big.Int, error) {
	rBytes, err := hex.DecodeString(rHex)
	if err != nil {
		return nil, err
	}
	r := new(big.Int).SetBytes(rBytes)
	rInv := new(big.Int).ModInverse(r, SECP256K1_N)
	if rInv == nil {
		return nil, ErrRNoModularInverse
	}
	return rInv, nil
}

// T computes the elliptic curve point T = rInv * R, where rInv is a scalar and R is a public key.
// It performs scalar multiplication on the secp256k1 curve and returns the resulting public key.
func T(rInv *big.Int, R *btcec.PublicKey) *btcec.PublicKey {
	// Perform scalar multiplication: rInv * R
	Tx, Ty := btcec.S256().ScalarMult(R.X(), R.Y(), rInv.Bytes())

	// Convert *big.Int to *field.Val
	var fx, fy btcec.FieldVal
	fx.SetByteSlice(Tx.Bytes())
	fy.SetByteSlice(Ty.Bytes())

	// Construct a new public key from FieldVals
	return btcec.NewPublicKey(&fx, &fy)
}

// U computes the elliptic curve point U = w * G, where G is the secp256k1 generator.
// It returns the X and Y coordinates of the resulting point as *big.Int values.
func U(w *big.Int) (*big.Int, *big.Int) {
	// G is the generator point for secp256k1
	curve := btcec.S256()

	// Multiply the curve's generator G by the scalar w (point multiplication)
	// This gives you the point U = w * G on the elliptic curve.
	return curve.ScalarBaseMult(w.Bytes())
}

// W computes w = -rInv * hashedMessage mod N, where rInv is the modular inverse of r,
// hashedMessage is the message hash as a big.Int, and N is the order of the secp256k1 curve.
// This is typically used in ECDSA signature verification and related cryptographic operations.
func W(rInv, hashedMessage *big.Int) *big.Int {
	w := new(big.Int).Mul(rInv, hashedMessage) // rInv * msg
	w.Neg(w)                                   // -(rInv * msg)
	w.Mod(w, SECP256K1_N)                      // mod N to ensure in [0, N)
	if w.Sign() < 0 {                          // ensure unsigned mod result
		w.Add(w, SECP256K1_N)
	}
	return w
}
