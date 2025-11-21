package validator

import (
	"os"

	"github.com/go-playground/validator/v10"
)

// fileExists checks if a file exists.
func fileExists(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	if path == "" {
		return true // If empty, skip the validation (optional)
	}
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
