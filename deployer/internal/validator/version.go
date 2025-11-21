package validator

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Regular expression to validate Semantic Versioning (SemVer) format
const versionPattern = `^v?\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?$`

// Custom validator for checking valid RollupVersion
func version(fl validator.FieldLevel) bool {
	// Get the value of the RollupVersion field
	version := strings.TrimSpace(fl.Field().String())
	// Compile the regular expression
	re := regexp.MustCompile(versionPattern)

	// Check if the version matches the pattern
	return re.MatchString(version)
}
