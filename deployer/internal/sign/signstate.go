package sign

import (
	"deployer/internal/types"
	"sync"
)

type Precomputes struct {
	mu sync.RWMutex
	*types.Precomputes
}

// newPrecomputes creates and returns a new instance of the Precomputes struct.
// This function initializes the Precomputes structure with default zero values.
func newPrecomputes() *Precomputes {
	return &Precomputes{
		Precomputes: &types.Precomputes{
			Signature: &types.Signature{},
			T:         &types.T{},
			U:         &types.U{},
		},
	}
}

// SetSignature safely sets the Signature field in the Precomputes struct.
// It acquires a write lock to ensure thread-safe access.
func (p *Precomputes) SetSignature(sig *types.Signature) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.setSignature(sig)
}

// SetT safely sets the T field in the Precomputes struct.
// It acquires a write lock to ensure thread-safe access.
func (p *Precomputes) SetT(t *types.T) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.setT(t)
}

// SetU safely sets the value of U in the Precomputes struct.
// It acquires a lock to ensure thread-safe access and then delegates
// the actual setting logic to the internal setU method.
func (p *Precomputes) SetU(u *types.U) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.setU(u)
}

// GetSignature safely retrieves the Signature field from the Precomputes struct.
// It acquires a read lock to ensure thread-safe access.
func (p *Precomputes) GetSignature() *types.Signature {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.getSignature()
}

// GetSignatureS safely retrieves the signature S register values from the Precomputes instance.
// It acquires a read lock to ensure thread-safe access and returns a pointer to types.Registers.
func (p *Precomputes) GetSignatureS() *types.Registers {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.getSignatureS()
}

// GetSignatureR safely retrieves the signature registers (R) from the Precomputes instance.
// It acquires a read lock to ensure thread-safe access to the underlying data.
// Returns a pointer to types.Registers containing the signature R values.
func (p *Precomputes) GetSignatureR() *types.Registers {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.getSignatureR()
}

// GetSignatureV safely retrieves the signature V value from the Precomputes struct.
// It acquires a read lock to ensure concurrent access is handled correctly.
func (p *Precomputes) GetSignatureV() uint8 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.getSignatureV()
}

// GetT safely retrieves the current value of the T field from the Precomputes struct.
// It acquires a read lock to ensure thread-safe access and returns a pointer to types.T.
func (p *Precomputes) GetT() *types.T {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.getT()
}

// GetU safely retrieves the current value of the U field from the Precomputes struct.
// It acquires a read lock to ensure thread-safe access and returns a pointer to types.U.
func (p *Precomputes) GetU() *types.U {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.getU()
}

// setSignature assigns the provided signature to the Precomputes struct.
// It updates the Signature field with the given *types.Signature value.
func (p *Precomputes) setSignature(sig *types.Signature) {
	p.Precomputes.Signature = sig
}

// setT sets the T field of the Precomputes struct to the provided value.
// It updates the internal Precomputes.T reference with the given *types.T.
//
// Parameters:
//   - t: A pointer to a types.T instance to assign to Precomputes.T.
func (p *Precomputes) setT(t *types.T) {
	p.Precomputes.T = t
}

// setU sets the U field of the Precomputes struct to the provided value.
// It updates the Precomputes.U with the given *types.U instance.
func (p *Precomputes) setU(u *types.U) {
	p.Precomputes.U = u
}

// getSignature returns the precomputed signature from the Precomputes struct.
// It provides access to the stored *types.Signature value.
func (p *Precomputes) getSignature() *types.Signature {
	return p.Precomputes.Signature
}

// getSignatureS returns the S register values from the Signature field of the Precomputes struct.
// It provides access to the stored *types.Registers representing the S values.
func (p *Precomputes) getSignatureS() *types.Registers {
	return p.getSignature().S
}

// getSignatureR returns the R register values from the Signature field of the Precomputes struct.
// It provides access to the stored *types.Registers representing the R values.
func (p *Precomputes) getSignatureR() *types.Registers {
	return p.getSignature().R
}

// getSignatureV returns the 'V' value from the Signature field of the Precomputes struct.
// The 'V' value is typically used as part of an ECDSA signature in Ethereum-based systems.
func (p *Precomputes) getSignatureV() uint8 {
	return p.getSignature().V
}

// getT returns the T value from the Precomputes struct.
// It provides access to the embedded types.T instance.
func (p *Precomputes) getT() *types.T {
	return p.Precomputes.T
}

// getU returns the U field from the Precomputes struct.
// It provides access to the precomputed value of type *types.U.
func (p *Precomputes) getU() *types.U {
	return p.Precomputes.U
}

// ! Attention when calling this. It will wipe out the pointers. Only if you want to reuse the same object
// Clear resets the Precomputes struct by setting its signature, T, and U fields to nil.
// It acquires a read lock to ensure thread-safe access during the reset operation.
func (p *Precomputes) Clear() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	p.setSignature(nil)
	p.setT(nil)
	p.setU(nil)
}
