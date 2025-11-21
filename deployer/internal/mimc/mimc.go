package mimc

import (
	"deployer/internal/types"
	"deployer/internal/validator"
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"

	"golang.org/x/crypto/sha3"
)

const (
	MimcNbRounds = 220          // number of rounds for the MiMC sponge
	Seed         = "mimcsponge" // seed to derive the constants
)

// Params constants for the mimc hash function
var (
	ErrEmptyInputs    = errors.New("inputs cannot be empty")
	ErrEmptySeed      = errors.New("seed cannot be empty")
	ErrInvalidRounds  = errors.New("numRounds must be greater than 0")
	ErrInvalidOutputs = errors.New("numOutputs must be greater than 0")

	zero fr.Element // zero initialized
)

type MiMCSponge struct {
	*types.MiMCSponge
}

// NewMiMCSponge initializes a new MiMCSponge instance
func NewMiMCSponge(seed string, numRounds int) (*MiMCSponge, error) {
	// Check if the number of rounds is valid
	if numRounds <= 0 {
		// Return an error if numRounds is less than or equal to 0
		return nil, ErrInvalidRounds
	}

	// Check if the seed is empty
	if len(seed) == 0 {
		// Return an error if seed is empty
		return nil, ErrEmptySeed
	}

	// Initialize the MiMCSponge instance
	mimcSponge := &MiMCSponge{
		MiMCSponge: &types.MiMCSponge{
			Seed:      seed,
			NumRounds: numRounds,
			Constants: make([]fr.Element, numRounds),
		},
	}

	// Generate constants for the MiMC sponge
	// Initialize the mimcConstants array
	mimcConstants := getConstants([]byte(seed), numRounds)
	for i := 0; i < numRounds; i++ {
		mimcSponge.Constants[i] = mimcConstants[i]
	}

	// Validate the structs
	if err := validator.ValidateStruct(mimcSponge); err != nil {
		return nil, fmt.Errorf("failed to validate MiMC: %w", err)
	}

	return mimcSponge, nil
}

// getConstants generates the MiMC constants using the Keccak-256 hash function
// The constants are derived from the seed and the number of rounds.
func getConstants(seed []byte, nRounds int) []fr.Element {
	c_partial := make([]fr.Element, nRounds)

	// c = keccak256(seed)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(seed)
	c := hash.Sum(nil)

	for i := 1; i < nRounds; i++ {
		// c = keccak256(c)
		hash.Reset()
		hash.Write(c)
		c = hash.Sum(nil)

		c_partial[i].SetBytes(c)
	}

	// Set c_0 and c_last to 0
	c_partial[0].SetZero()
	c_partial[nRounds-1].SetZero()

	return c_partial
}
