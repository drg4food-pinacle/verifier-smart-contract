package validator

import (
	"os"

	"github.com/go-playground/validator/v10"
)

// isExecutable checks if a file is executable (e.g., abigen, solc).
func isExecutable(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	if path == "" {
		return true // Skip validation if empty
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	mode := info.Mode()
	return !info.IsDir() && mode&0111 != 0
}
