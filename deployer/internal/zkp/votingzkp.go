package zkp

import (
	"deployer/internal/types"
	"encoding/json"
	"math/big"
	"sync"
)

type PinacleZKP struct {
	mu sync.RWMutex
	*types.PinacleZKP
}

// NewZKP creates and returns a new instance of the ZKP struct with all fields initialized to their default values.
// This includes zero values for numeric fields, empty slices for PathElements and PathIndices, and new instances
// for T and U fields. The function is typically used to generate a fresh ZKP object for further configuration or use.
func NewZKP() *PinacleZKP {
	return &PinacleZKP{
		PinacleZKP: &types.PinacleZKP{
			PathElements: zeroArr,
			PathIndices:  zeroArr,
		},
	}
}

// SetPrivateKey sets the private key for the VotingZKP instance.
// It acquires a read lock to ensure thread-safe access while updating the PrivateKey field.
// The provided key is expected to be a pointer to a types.Registers struct.
func (zkp *PinacleZKP) SetPrivateKey(key *types.Registers) {
	zkp.mu.RLock()
	defer zkp.mu.RUnlock()
	zkp.PrivateKey = key
}

// SetPathElement sets the PathElements field of the ZKP struct to the provided slice of big.Int pointers.
// This method replaces any existing path elements with the new slice.
//
// pathElements: A slice of pointers to big.Int representing the path elements to be set.
func (zkp *PinacleZKP) SetPathElement(pathElements [LEVELS]*big.Int) {
	zkp.mu.RLock()
	defer zkp.mu.RUnlock()
	zkp.PathElements = pathElements
}

// SetPathIndices sets the PathIndices field of the ZKP struct to the provided slice of big.Int pointers.
// This method replaces any existing path indices with the new slice.
//
// pathIndices: A slice of pointers to big.Int representing the path indices to be set.
func (zkp *PinacleZKP) SetPathIndices(pathIndices [LEVELS]*big.Int) {
	zkp.mu.RLock()
	defer zkp.mu.RUnlock()
	zkp.PathIndices = pathIndices
}

// Creates []byte output ready for zkp
func (zkp *PinacleZKP) MarshalJSON() ([]byte, error) {
	zkp.mu.RLock()
	defer zkp.mu.RUnlock()

	// Prepare map[string]interface{} or a struct literal with string fields
	m := map[string]interface{}{}

	// Convert slice of *big.Int S to string slice
	privateKeyStrings := make([]string, len(zkp.getPrivateKey()))
	for i, v := range zkp.getPrivateKey() {
		if v == nil {
			privateKeyStrings[i] = "0"
		} else {
			privateKeyStrings[i] = v.String()
		}
	}
	m["privateKey"] = privateKeyStrings

	// PathElements [LEVELS]*big.Int to []string
	pathElements := make([]string, len(zkp.getPathElements()))
	for i, v := range zkp.getPathElements() {
		if v == nil {
			pathElements[i] = "0"
		} else {
			pathElements[i] = v.String()
		}
	}
	m["pathElements"] = pathElements

	// PathIndices [LEVELS]*big.Int to []string
	pathIndices := make([]string, len(zkp.getPathIndices()))
	for i, v := range zkp.getPathIndices() {
		if v == nil {
			pathIndices[i] = "0"
		} else {
			pathIndices[i] = v.String()
		}
	}
	m["pathIndices"] = pathIndices

	// Marshal the map to JSON bytes
	return json.Marshal(m)
}

func (zkp *PinacleZKP) getPrivateKey() *types.Registers {
	return zkp.PrivateKey
}

func (zkp *PinacleZKP) getPathElements() [LEVELS]*big.Int {
	return zkp.PathElements
}

func (zkp *PinacleZKP) getPathIndices() [LEVELS]*big.Int {
	return zkp.PathIndices
}
