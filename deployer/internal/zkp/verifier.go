package zkp

import (
	"fmt"
	"os"
	"sync"

	verifier "github.com/iden3/go-rapidsnark/verifier"
)

type Verifier struct {
	mu   sync.RWMutex
	vkey []byte
}

// NewProve creates a new Prove instance by loading wasm and zkey files.
func NewVerifier(vkey Path) (*Verifier, error) {
	// Load wasm bytes and create calculator
	verificationKey, err := os.ReadFile(string(vkey))
	if err != nil {
		return nil, fmt.Errorf("failed to read vkey file: %w", err)
	}

	return &Verifier{
		vkey: verificationKey,
	}, nil
}

// VerifyProofs verifies the provided zero-knowledge proofs using the Groth16 verifier and the current verification key.
// It acquires a read lock to ensure thread-safe access to the verification key.
// Returns an error if the proof verification fails, otherwise returns nil.
func (v *Verifier) VerifyProofs(proofs *ZKProof) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// Verify the proof
	err := verifier.VerifyGroth16(*proofs.ZKProof, v.getVerificationKey())
	if err != nil {
		return fmt.Errorf("failed to verify proof: %v", err)
	}

	return nil
}

// getVerificationKey returns the verification key as a byte slice.
// This key is typically used in zero-knowledge proof verification processes.
func (v *Verifier) getVerificationKey() []byte {
	return v.vkey
}
