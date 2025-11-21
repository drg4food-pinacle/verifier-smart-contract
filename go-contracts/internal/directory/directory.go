package directory

import (
	"encoding/json"
	"fmt"
	"go-contracts/internal/logger"
	"os"
	"path/filepath"
)

// checkDirExists checks if a directory exists and is a directory.
func CheckDirExists(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("failed to check directory %s: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory: %s", path)
	}
	return nil
}

// createDirIfNotExists creates a directory if it does not exist.
// It returns nil if the directory already exists.
func CreateDirIfNotExists(path string) error {
	if err := CheckDirExists(path); err == nil {
		return nil // already exists
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

func DeleteDir(file string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	for {
		path := filepath.Join(dir, file)
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			// Found it, delete
			if err := os.RemoveAll(path); err != nil {
				return err
			}
			logger.Logger.Info().Msgf("⚠️  Deleted: %s", path)
			return nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return os.ErrNotExist
}

// SaveToFile serializes any data structure as indented JSON and writes it to the specified file path.
func SaveToFile(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}
