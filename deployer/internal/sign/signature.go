package sign

import (
	"errors"
	"fmt"
	"math/big"
	"sync"

	"deployer/internal/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	ErrInvalidSignatureLength      = errors.New("invalid signature length")
	ErrInvalidSignatureS           = errors.New("signature 's' value too high")
	ErrInvalidSignatureV           = errors.New("invalid 'v' value")
	ErrPublicKeyRecovery           = errors.New("public key recovery failed")
	ErrSignatureVerificationFailed = errors.New("signature verification failed")
	ErrSigningMessage              = errors.New("error signing message")
	ErrSplitingSignature           = errors.New("error splitting signature")
	ErrSignatureReconstruction     = errors.New("reconstruct signature failed")
)

type Signature struct {
	mu sync.RWMutex
	*types.Signature
}

func NewSignature() *Signature {
	return &Signature{
		Signature: &types.Signature{},
	}
}

// SplitSignature splits a 65-byte Ethereum signature into r, s, and v components.
// Returns an error if the signature length is not exactly 65 bytes.
func (sig *Signature) SplitSignature(signature []byte) error {
	if len(signature) != 65 {
		return ErrInvalidSignatureLength
	}

	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:64])
	v := signature[64] + 27 // Convert to Ethereum's [27,28] style

	// Validate s
	if s.Cmp(HalfOrder) > 0 {
		return ErrInvalidSignatureS
	}

	// Validate v
	if v != 27 && v != 28 {
		return ErrInvalidSignatureV
	}

	sig.mu.RLock()
	defer sig.mu.RUnlock()

	sig.setS(BigIntToRegisters(s))
	sig.setR(BigIntToRegisters(r))
	sig.setV(v)
	return nil
}

// ReconstructSignature reconstructs the 65-byte Ethereum signature from R, S, and V registers.
func (sig *Signature) ReconstructSignature() ([]byte, error) {
	sig.mu.RLock()
	defer sig.mu.RUnlock()

	return sig.reconstructSignature()
}

// VerifySignature verifies the signature against the provided hashed message.
// It reconstructs the signature, recovers the public key from the signature and hashed message,
// and returns the Ethereum address derived from the recovered public key.
// Returns an error if signature reconstruction or public key recovery fails.
func (sig *Signature) VerifySignature(hashedMessage []byte) (common.Address, error) {
	sig.mu.RLock()
	defer sig.mu.RUnlock()

	// Reconstruct signature
	signature, err := sig.reconstructSignature()
	if err != nil {
		return common.Address{}, fmt.Errorf("%s :%s", ErrSignatureReconstruction, err)
	}

	// Recover address to verify
	pubKey, err := crypto.SigToPub(hashedMessage, signature)
	if err != nil {
		return common.Address{}, fmt.Errorf("%s: %s", ErrPublicKeyRecovery, err)
	}

	return crypto.PubkeyToAddress(*pubKey), nil
}

// SetS safely sets the internal state of the Signature using the provided Registers.
// It acquires a read lock before delegating to the internal setS method to ensure
// thread-safe access to the Signature's data.
func (sig *Signature) SetS(s *types.Registers) {
	sig.mu.RLock()
	defer sig.mu.RUnlock()
	sig.setS(s)
}

// SetR safely sets the R field of the Signature struct using the provided
// types.Registers value. It acquires a read lock to ensure thread-safe access
// before delegating the actual assignment to the setR helper method.
func (sig *Signature) SetR(r *types.Registers) {
	sig.mu.RLock()
	defer sig.mu.RUnlock()
	sig.setR(r)
}

// SetV sets the 'v' value of the Signature to the provided uint8 value.
// It validates that 'v' is either 27 or 28, returning ErrInvalidSignatureV if not.
// The method acquires a read lock before setting the value to ensure thread safety.
func (sig *Signature) SetV(v uint8) error {
	if v != 27 && v != 28 {
		return ErrInvalidSignatureV
	}
	sig.mu.RLock()
	defer sig.mu.RUnlock()
	sig.setV(v)
	return nil
}

// GetR safely retrieves the R register from the Signature instance.
// It acquires a read lock to ensure thread-safe access to the underlying data.
// Returns a pointer to types.Registers representing the R value.
func (sig *Signature) GetR() *types.Registers {
	sig.mu.RLock()
	defer sig.mu.RUnlock()
	return sig.getR()
}

// setR sets the R field of the Signature to the provided types.Registers value.
// This method updates the internal state of the Signature with the given register values.
func (sig *Signature) setR(r *types.Registers) {
	sig.Signature.R = r
}

// setS sets the S field of the Signature to the provided types.Registers value.
// This method updates the internal signature's S component, which is typically
// used as part of a cryptographic signature structure.
//
// Parameters:
//   - s: A pointer to a types.Registers instance representing the new S value.
func (sig *Signature) setS(s *types.Registers) {
	sig.Signature.S = s
}

// setV sets the V value of the Signature to the provided uint8 value.
// V is typically used as the recovery identifier in ECDSA signatures.
func (sig *Signature) setV(v uint8) {
	sig.Signature.V = v
}

// getR returns the R component of the signature as a pointer to types.Registers.
func (sig *Signature) getR() *types.Registers {
	return sig.Signature.R
}

// getS returns the S component of the signature as a pointer to types.Registers.
func (sig *Signature) getS() *types.Registers {
	return sig.Signature.S
}

// getV returns the V value from the underlying Signature, which is typically used as part of the ECDSA signature (recovery identifier).
func (sig *Signature) getV() uint8 {
	return sig.Signature.V
}

// reconstructSignature reconstructs the signature byte slice in Ethereum format (R || S || V)
// from the Signature struct. It ensures that R and S are each 32 bytes, and V is either 27 or 28.
// Returns the concatenated signature bytes or an error if reconstruction fails.
func (sig *Signature) reconstructSignature() ([]byte, error) {
	r, err := reconstructBigIntFromRegisters(sig.getR())
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct R: %w", err)
	}

	s, err := reconstructBigIntFromRegisters(sig.getS())
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct S: %w", err)
	}

	// Ensure R and S are 32 bytes each (Ethereum signature format)
	rBytes := r.FillBytes(make([]byte, 32))
	sBytes := s.FillBytes(make([]byte, 32))

	// V must be 0 or 1 in recovery id format, or 27/28 in Ethereum format
	v := sig.getV()
	if v != 27 && v != 28 {
		return nil, fmt.Errorf("%s: %d", ErrInvalidSignatureV, v)
	}

	// Final signature: 32 bytes R + 32 bytes S + 1 byte V
	return append(append(rBytes, sBytes...), v-27), nil
}

// reconstructBigIntFromRegisters reconstructs a big.Int value from a slice of four registers,
// where each register represents a 64-bit segment of the original integer. The function checks
// that each register does not exceed the maximum value for a uint64. It then combines the
// registers by shifting and adding them to form the final big.Int value. Returns an error if
// any register exceeds the uint64 limit.
func reconstructBigIntFromRegisters(reg *types.Registers) (*big.Int, error) {
	for i, v := range reg {
		if v.Cmp(max) > 0 {
			return nil, fmt.Errorf("register[%d] exceeds uint64: %s", i, v.String())
		}
	}

	bigA := new(big.Int).Set(reg[0])
	bigB := new(big.Int).Lsh(reg[1], 64)
	bigC := new(big.Int).Lsh(reg[2], 128)
	bigD := new(big.Int).Lsh(reg[3], 192)

	result := new(big.Int).Add(bigA, bigB)
	result.Add(result, bigC)
	result.Add(result, bigD)

	return result, nil
}
