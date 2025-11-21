package sign

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
)

var ErrPrivateKeyOutOfRange = errors.New("invalid private key, out of range")

type ECDSA struct {
	mu sync.RWMutex
	*ecdsa.PrivateKey
}

// NewECDSA creates and returns a new instance of the ECDSA struct.
func NewECDSA() *ECDSA {
	return &ECDSA{}
}

// SetPrivateKey sets the ECDSA private key for the receiver.
// It acquires a read lock before delegating to the internal setPrivateKey method.
// Note: This method uses a read lock (RLock), which may not be appropriate for write operations.
func (e *ECDSA) SetPrivateKey(key *ecdsa.PrivateKey) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	e.setPrivateKey(key)
}

// GetPrivateKey returns the underlying ECDSA private key in a thread-safe manner.
// It acquires a read lock to ensure safe concurrent access to the private key.
// The returned *ecdsa.PrivateKey should be handled with care to avoid leaking sensitive information.
func (e *ECDSA) GetPrivateKey() *ecdsa.PrivateKey {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.getPrivateKey()
}

// GetPublicKey returns the ECDSA public key in a thread-safe manner.
// It acquires a read lock to ensure safe concurrent access to the private key.
// If the private key is not set, it returns nil.
func (e *ECDSA) GetPublicKey() *ecdsa.PublicKey {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.getPrivateKey() == nil {
		return nil
	}
	return e.getPublicKey()
}

// LoadPrivateKeyFromHex loads a hex-encoded private key string into an *ecdsa.PrivateKey
func (e *ECDSA) LoadPrivateKeyFromHex(hexkey string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Remove 0x prefix if present
	if len(hexkey) >= 2 && hexkey[0:2] == "0x" {
		hexkey = hexkey[2:]
	}

	// Decode hex string to bytes
	keyBytes, err := hex.DecodeString(hexkey)
	if err != nil {
		return err
	}

	// Use the crypto package to convert bytes to ecdsa.PrivateKey
	privKey, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		return err
	}

	// Optional: verify private key is valid (private key < secp256k1 curve order)
	if privKey.D.Cmp(SECP256K1_N) >= 0 || privKey.D.Sign() <= 0 {
		return ErrPrivateKeyOutOfRange
	}
	e.setPrivateKey(privKey)
	return nil
}

// SignMessage signs the provided hashed message using the ECDSA private key associated with the receiver.
// It returns a Signature object containing the split signature components (R, S, V) or an error if signing fails.
// The method is thread-safe and acquires a read lock during the signing process.
//
// Parameters:
//   - hashed: The hashed message to be signed.
//
// Returns:
//   - *Signature: The resulting signature object.
//   - error: An error if signing or signature splitting fails.
func (e *ECDSA) SignMessage(hashed []byte) (*Signature, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.signMessage(hashed)
}

// Sign takes a plain message string, hashes it using Ethereum's personal message hashing,
// signs it with the ECDSA private key, verifies the signature, and then computes and returns
// precomputed elliptic curve values (T and U) used for further cryptographic operations.
// It ensures the signature is valid and corresponds to the current private key.
// Returns a Precomputes struct containing the signature and computed points, or an error.
func (e *ECDSA) Sign(msg string) (*Precomputes, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Hash the message
	hashed := hashPersonalMessage(msg)

	// Sign message and take signature (register)
	sig, err := e.signMessage(hashed)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %s", err)
	}

	// Verify Signature
	recoveredAddr, err := sig.VerifySignature(hashed)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signature: %s", err)
	}

	originalAddr := crypto.PubkeyToAddress(*e.getPublicKey())
	if recoveredAddr != originalAddr {
		return nil, ErrSignatureVerificationFailed
	}

	// Reconstruct the r part
	r, err := reconstructBigIntFromRegisters(sig.getR())
	if err != nil {
		return nil, err
	}
	// Convert to hex string, split into registers (stubbed function)
	rHex := fmt.Sprintf("%064x", r)

	isYOdd := (sig.getV()-27)%2 == 1

	// Get r point from x coordinate and y parity (like curve.pointFromX)
	rPoint, err := pointFromX(rHex, isYOdd)
	if err != nil {
		return nil, err
	}

	// RInv
	rInv, err := rInv(rHex)
	// Calcualte W
	w := W(rInv, new(big.Int).SetBytes(hashed))
	// Calculate U
	Ux, Uy := U(w)
	// Calculate TPrecomputes
	T := T(rInv, rPoint)

	// Elliptic curve operations
	sig.mu.RLock()
	precomputes := newPrecomputes()
	precomputes.setSignature(sig.Signature)
	precomputes.setT(pointPrecomputes(T))
	precomputes.setU(pointU(Ux, Uy))
	sig.mu.RUnlock()

	return precomputes, nil
}

// setPrivateKey sets the ECDSA instance's private key to the provided ecdsa.PrivateKey.
// This method is intended for internal use to assign or update the private key used for signing operations.
func (e *ECDSA) setPrivateKey(key *ecdsa.PrivateKey) {
	e.PrivateKey = key
}

// getPrivateKey returns the underlying ECDSA private key associated with the ECDSA instance.
// This method provides access to the raw *ecdsa.PrivateKey for cryptographic operations.
// Use with caution to avoid exposing sensitive key material.
func (e *ECDSA) getPrivateKey() *ecdsa.PrivateKey {
	return e.PrivateKey
}

// getPublicKey returns the ECDSA public key associated with the current private key.
// It provides access to the public component of the ECDSA key pair.
func (e *ECDSA) getPublicKey() *ecdsa.PublicKey {
	return &e.PrivateKey.PublicKey
}

// signMessage signs the given hashed message using the ECDSA private key.
// It returns a Signature object containing the split signature components (R, S, V).
// This is an internal helper method and assumes the caller has already acquired the necessary locks.
func (e *ECDSA) signMessage(hashed []byte) (*Signature, error) {
	// Sign the hash and return signature bytes
	sig, err := crypto.Sign(hashed, e.getPrivateKey())
	if err != nil {
		return nil, fmt.Errorf("%s: %s", ErrSigningMessage, err)
	}

	signature := NewSignature()
	err = signature.SplitSignature(sig)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", ErrSplitingSignature, err)
	}
	return signature, nil
}
