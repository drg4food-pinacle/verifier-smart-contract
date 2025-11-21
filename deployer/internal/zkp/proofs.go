package zkp

import (
	"fmt"
	"math/big"
	"sync"

	zklogin "deployer/internal/abigen/zkLogin"
	"deployer/internal/types"

	rapidsnark "github.com/iden3/go-rapidsnark/types"
)

// ZKProof wraps types.ZKProof and provides thread-safe access to its fields.
type ZKProof struct {
	mu sync.RWMutex
	*rapidsnark.ZKProof
}

// NewZKProof creates a new empty ZKProof with thread-safe access.
func NewZKProof() *ZKProof {
	return &ZKProof{
		ZKProof: &rapidsnark.ZKProof{},
	}
}

// SetProof sets the proof byte slice.
func (zkp *ZKProof) SetProof(proof *rapidsnark.ProofData) {
	zkp.mu.Lock()
	defer zkp.mu.Unlock()
	zkp.setProof(proof)
}

// GetProof returns the proof byte slice.
func (zkp *ZKProof) GetProof() *rapidsnark.ProofData {
	zkp.mu.RLock()
	defer zkp.mu.RUnlock()
	return zkp.getProof()
}

// SetPublicSignals sets the public signals slice.
func (zkp *ZKProof) SetPublicSignals(signals []string) {
	zkp.mu.Lock()
	defer zkp.mu.Unlock()
	zkp.setPublicSignals(signals)
}

// GetPublicSignals returns the full list of public signals.
func (zkp *ZKProof) GetPublicSignals() []string {
	zkp.mu.RLock()
	defer zkp.mu.RUnlock()
	return zkp.getPublicSignals()
}

// GetPublicSignal returns the signal at the specified index if valid.
func (zkp *ZKProof) GetPublicSignal(index int) (string, error) {
	zkp.mu.RLock()
	defer zkp.mu.RUnlock()
	publicSignal, err := zkp.getPublicSignal(index)
	if err != nil {
		return "", fmt.Errorf("GetPublicSignal: %w", err)
	}
	return publicSignal, nil
}

// GetPublicSignalsBigInt returns all public signals converted to []*big.Int.
// It acquires a read lock for thread-safe access and fails fast on any conversion error.
func (zkp *ZKProof) GetPublicSignalsBigInt() ([]*big.Int, error) {
	zkp.mu.RLock()
	defer zkp.mu.RUnlock()

	rawSignals := zkp.getPublicSignals()
	bigInts := make([]*big.Int, len(rawSignals))

	for i, hexStr := range rawSignals {
		bi, ok := new(big.Int).SetString(hexStr, 10)
		if !ok {
			return nil, fmt.Errorf("GetPublicSignalsBigInt: failed to parse signal at index %d: %q", i, hexStr)
		}
		bigInts[i] = bi
	}

	return bigInts, nil
}

// GetPublicSignalBigInt retrieves the public signal at the specified index,
// converts it from a hexadecimal string to a *big.Int, and returns it.
// It acquires a read lock to ensure thread-safe access to the underlying data.
// Returns an error if the public signal cannot be retrieved or if the conversion fails.
func (zkp *ZKProof) GetPublicSignalBigInt(index int) (*big.Int, error) {
	zkp.mu.RLock()
	defer zkp.mu.RUnlock()
	publicSignal, err := zkp.getPublicSignal(index)
	if err != nil {
		return nil, fmt.Errorf("GetPublicSignal: %w", err)
	}
	publicSignalBigInt, ok := new(big.Int).SetString(publicSignal, 16)
	if !ok {
		return nil, fmt.Errorf("GetPublicSignalBigInt: %w", err)
	}
	return publicSignalBigInt, nil
}

func (zkp *ZKProof) ConvertProof() (*zklogin.ZkLoginGroth16Proof, error) {
	proof := &zklogin.ZkLoginGroth16Proof{} // allocate struct before use

	// PiA
	if len(zkp.getProof().A) < 2 {
		return proof, fmt.Errorf("invalid PiA length: got %d", len(zkp.getProof().A))
	}
	for i := 0; i < 2; i++ {
		bi, ok := new(big.Int).SetString(zkp.getProof().A[i], 10)
		if !ok {
			return proof, fmt.Errorf("invalid PiA[%d]: %q", i, zkp.getProof().A[i])
		}
		proof.PiA[i] = bi
	}

	// PiB
	if len(zkp.getProof().B) < 2 || len(zkp.getProof().B[0]) < 2 || len(zkp.getProof().B[1]) < 2 {
		return proof, fmt.Errorf("invalid PiB structure")
	}
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			// reverse the inner arrays: [b00, b01] becomes [b01, b00]
			bi, ok := new(big.Int).SetString(zkp.getProof().B[i][1-j], 10)
			if !ok {
				return proof, fmt.Errorf("invalid PiB[%d][%d]: %q", i, j, zkp.getProof().B[i][1-j])
			}
			proof.PiB[i][j] = bi
		}
	}

	// PiC
	if len(zkp.getProof().C) < 2 {
		return proof, fmt.Errorf("invalid PiC length: got %d", len(zkp.getProof().C))
	}
	for i := 0; i < 2; i++ {
		bi, ok := new(big.Int).SetString(zkp.getProof().C[i], 10)
		if !ok {
			return proof, fmt.Errorf("invalid PiC[%d]: %q", i, zkp.getProof().C[i])
		}
		proof.PiC[i] = bi
	}

	return proof, nil
}

func (zkp *ZKProof) ConvertPublicSignals() ([2]*big.Int, error) {
	var arr [2]*big.Int

	if len(zkp.getPublicSignals()) != types.PINACLE_PUBLIC_SIGNALS {
		return arr, fmt.Errorf("expected %d public signals got %d", types.PINACLE_PUBLIC_SIGNALS, len(zkp.getPublicSignals()))
	}

	for i, sig := range zkp.getPublicSignals() {
		bi, ok := new(big.Int).SetString(sig, 10) // or base 16 if your signals are hex strings
		if !ok {
			return arr, fmt.Errorf("invalid public signal at index %d: %s", i, sig)
		}
		arr[i] = bi
	}

	return arr, nil
}

// SetProof sets the proof byte slice.
func (zkp *ZKProof) setProof(proof *rapidsnark.ProofData) {
	zkp.Proof = proof
}

// GetProof returns the proof byte slice.
func (zkp *ZKProof) getProof() *rapidsnark.ProofData {
	return zkp.Proof
}

// SetPublicSignals sets the public signals slice.
func (zkp *ZKProof) setPublicSignals(signals []string) {
	zkp.PubSignals = signals
}

// GetPublicSignals returns the full list of public signals.
func (zkp *ZKProof) getPublicSignals() []string {
	return zkp.PubSignals
}

// GetPublicSignals returns the full list of public signals.
func (zkp *ZKProof) getPublicSignal(index int) (string, error) {
	if index < 0 || index >= len(zkp.getPublicSignals()) {
		return "", fmt.Errorf("invalid index %d: out of bounds", index)
	}
	return zkp.PubSignals[index], nil
}
