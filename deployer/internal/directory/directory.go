package directory

import (
	"deployer/internal/logger"
	"deployer/internal/validator"
	"encoding/json"
	"fmt"
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
			logger.Logger.Info().Str("path", path).Msg("File Deleted")
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
func SaveToFile(filePath string, data any) error {
	// Validate before saving
	if err := validator.ValidateStruct(data); err != nil {
		return fmt.Errorf("invalid accounts file structure: %w", err)
	}

	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Proceed to marshal and save to file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode json: %w", err)
	}

	return nil
}

// LoadFromFile loads a JSON file and unmarshals it into the provided target.
// `target` should be a pointer to the destination struct.
func LoadFromFile(filePath string, target any) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("failed to decode json from file: %w", err)
	}

	// Optional: validate after loading
	if err := validator.ValidateStruct(target); err != nil {
		return fmt.Errorf("invalid accounts file structure: %w", err)
	}
	return nil
}
