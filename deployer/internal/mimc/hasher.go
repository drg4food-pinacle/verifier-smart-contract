package mimc

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
)

// MultiHash computes the MiMC hash of multiple inputs
// and returns the result as a slice of fr.Element.
// The function takes a slice of big.Int inputs, a key of type fr.Element,
// and the number of outputs to generate.
// If the key is nil, it uses a zero-initialized key.
func (m *MiMCSponge) MultiHash(inputs []*big.Int, key *fr.Element, numOutputs int) ([]fr.Element, error) {
	// Check if the number of inputs is valid
	if len(inputs) == 0 {
		// Return an error if inputs are empty
		return nil, ErrEmptyInputs
	}

	// Check if the number of outputs is valid
	if numOutputs <= 0 {
		// Return an error if numOutputs is less than or equal to 0
		return nil, ErrInvalidOutputs
	}

	// If key is nil, use a zero-initialized key
	if key == nil {
		key = &zero
	}

	// Initialize the result and carry elements
	var R, C, el fr.Element

	for _, input := range inputs {
		el.SetZero()
		el.SetBigInt(input)

		R.Add(&R, &el)
		R, C = m.hash(&R, &C, key, false)
	}

	outputs := []fr.Element{R}
	for i := 1; i < numOutputs; i++ {
		R, C = m.hash(&R, &C, key, false)
		outputs = append(outputs, R)
	}

	return outputs, nil
}

// HashLeftRight computes the MiMC hash of two inputs and returns it as a string.
func (m *MiMCSponge) HashLeftRight(left, right *big.Int) (fr.Element, error) {
	// Call the hash function with the left and right inputs
	hash, err := m.MultiHash([]*big.Int{left, right}, nil, 1)
	if err != nil {
		// Handle the error if multiHash fails
		return fr.Element{}, fmt.Errorf("Error during multihash: " + err.Error())
	}
	return hash[0], nil
}

// Encrypt function encrypts the left and right inputs using the MiMC hash function
func (m *MiMCSponge) Encrypt(xL, xR, key *fr.Element) (fr.Element, fr.Element) {
	return m.hash(xL, xR, key, false)
}

// Decrypt function decrypts the left and right inputs using the MiMC hash function
func (m *MiMCSponge) Decrypt(xL, xR, key *fr.Element) (fr.Element, fr.Element) {
	return m.hash(xL, xR, key, true)
}

// HashAddress computes the MiMC hash of the given Ethereum address using the MiMCSponge instance.
// It parses the address as a big integer, applies the MiMC hash function, and returns the resulting hash as a *big.Int.
// The function returns a new *big.Int containing the hash value.
func (m *MiMCSponge) HashAddress(address *common.Address) *big.Int {
	leftBigInt := new(big.Int)
	resultBigInt := new(big.Int)
	var element fr.Element

	// Parse hex string without "0x"
	leftBigInt.SetString(address.Hex(), 0)

	// Compute Mimc hash
	element, _ = m.HashLeftRight(leftBigInt, RightBigInt)

	// Convert fr.Element to *big.Int
	element.BigInt(resultBigInt)

	// Return a copy
	return new(big.Int).Set(resultBigInt)
}

// Hash applies the MiMC Feistel permutation for encryption or decryption.
// If rev == false: encryption mode.
// If rev == true: decryption mode.
func (m *MiMCSponge) hash(xL_in, xR_in, key *fr.Element, rev bool) (fr.Element, fr.Element) {
	xL := *xL_in
	xR := *xR_in
	k := *key

	var t, t2, c, round fr.Element

	index := 0
	step := 1
	if rev {
		// Reverse the order of the constants for decryption
		index = m.NumRounds - 1
		step = -1
	}

	for i := 0; i < m.NumRounds; i++ {
		c.SetZero()
		c.Set(&m.Constants[index]) // Attach the constant to the current round
		index += step

		t.SetZero()
		t.Add(&xL, &k).Add(&t, &c)
		if i == 0 {
			t.Sub(&t, &c)
		}

		// Compute t^5
		t2.Square(&t)   // t^2
		t2.Square(&t2)  // t^4
		t2.Mul(&t2, &t) // t^5

		xRtmp := xR

		round.SetZero()
		if rev {
			t2.Neg(&t2) // Make t^5 negative when decrypting
		}
		round.Add(&xRtmp, &t2) // Always add t^5 (negative or positive based on rev)

		if i < (m.NumRounds - 1) {
			xR = xL
			xL = round
		} else {
			xR = round
		}
	}

	return xL, xR
}
