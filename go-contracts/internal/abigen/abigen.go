package abigen

import (
	"encoding/json"
	"fmt"
	"go-contracts/internal/logger"
	"go-contracts/internal/types"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"golang.org/x/sync/singleflight"
)

var (
	abigenDir = ".abigen/bin"
	maxTries  = 2
	dg        singleflight.Group // global download group

)

type Abigen struct {
	version Version // Solc version

	once          sync.Once
	abigenAbsPath string // solc absolute path
	err           error  // initialization error
}

func New(version Version) *Abigen {
	return &Abigen{
		version: version,
	}
}

// init initializes the compiler.
func (c *Abigen) init() {
	// check mod root is set
	if Root == "" {
		c.err = fmt.Errorf("abigen: no go.mod detected")
		return
	}

	// check or download abigen version
	c.abigenAbsPath, c.err = checkAbigen(c.version)
}

// Generate command
func (c *Abigen) Generate(file string, opts ...Option) error {
	// Apply options
	args := &abigenArgs{
		ContractPath: file,
	}
	for _, opt := range opts {
		opt(args)
	}

	// Validate contract file
	if _, err := os.Stat(args.ContractPath); err != nil {
		return fmt.Errorf("contract file not found: %w", err)
	}

	logger.Logger.Info().Msgf("🚀 Running abigen for %s", filepath.Base(file))

	// Parse JSON to extract ABI + bytecode
	data, err := os.ReadFile(args.ContractPath)
	if err != nil {
		return fmt.Errorf("read contract file: %w", err)
	}

	var contractJSON *types.Contract
	if err := json.Unmarshal(data, &contractJSON); err != nil {
		return fmt.Errorf("parse contract json: %w", err)
	}

	// Write temp ABI and BIN files
	tmpDir := os.TempDir()
	abiPath := filepath.Join(tmpDir, fmt.Sprintf("%s.abi", contractJSON.ContractName))
	binPath := filepath.Join(tmpDir, fmt.Sprintf("%s.bin", contractJSON.ContractName))

	abiBytes, err := json.Marshal(contractJSON.ABI)
	if err != nil {
		return fmt.Errorf("marshal abi: %w", err)
	}

	if err := os.WriteFile(abiPath, abiBytes, 0644); err != nil {
		return fmt.Errorf("write abi file: %w", err)
	}
	if err := os.WriteFile(binPath, []byte(contractJSON.Bytecode), 0644); err != nil {
		return fmt.Errorf("write bin file: %w", err)
	}

	defer os.Remove(abiPath)
	defer os.Remove(binPath)

	// Build abigen command
	cmdArgs := []string{
		"--abi", abiPath,
		"--bin", binPath,
		"--pkg", args.Pkg,
		"--out", args.OutGo,
	}
	cmdArgs = append(cmdArgs, args.Additional...)

	cmd := exec.Command(c.abigenAbsPath, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
