package zkp

import (
	"fmt"
	"os"
	"sync"

	"github.com/iden3/go-rapidsnark/prover"
	"github.com/iden3/go-rapidsnark/witness/v2"
	"github.com/iden3/go-rapidsnark/witness/wasmer"
)

type Prover struct {
	mu   sync.RWMutex
	calc witness.Calculator
	zkey []byte
}

// NewProve creates a new Prove instance by loading wasm and zkey files.
func NewProver(wasm, zkey Path) (*Prover, error) {
	// Load wasm bytes and create calculator
	wasmBytes, err := os.ReadFile(string(wasm))
	if err != nil {
		return nil, fmt.Errorf("failed to read wasm file: %w", err)
	}

	calc, err := witness.NewCalculator(wasmBytes, witness.WithWasmEngine(wasmer.NewCircom2WitnessCalculator))
	if err != nil {
		return nil, fmt.Errorf("failed to create witness calculator: %w", err)
	}

	// Load zkey bytes
	zkeyBytes, err := os.ReadFile(string(zkey))
	if err != nil {
		return nil, fmt.Errorf("failed to read zkey file: %w", err)
	}

	return &Prover{
		calc: calc,
		zkey: zkeyBytes,
	}, nil
}

func (p *Prover) GenerateProofs(inputJson []byte) (*ZKProof, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	inputs, err := witness.ParseInputs(inputJson)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inputs from JSON: %v", err)
	}

	wtns, err := p.calc.CalculateWTNSBin(inputs, true)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate witness: %w", err)
	}

	// Prove the proof
	proof, err := prover.Groth16Prover(p.zkey, wtns)
	if err != nil {
		return nil, fmt.Errorf("failed to prove: %w", err)
	}

	zkproof := NewZKProof()

	zkproof.mu.RLock()
	zkproof.setProof(proof.Proof)
	zkproof.setPublicSignals(proof.PubSignals)
	zkproof.mu.RUnlock()

	return zkproof, nil
}
