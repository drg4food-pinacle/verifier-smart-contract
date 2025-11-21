package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Use `mapstructure` tag names in error output (you can change to `json` if needed)
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("mapstructure"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	// Register custom validation function to check if a file exists
	validate.RegisterValidation("file_exists", fileExists)
	validate.RegisterValidation("is_executable", isExecutable)
	validate.RegisterValidation("version", version)
	validate.RegisterValidation("required_if", requiredIf)
	validate.RegisterValidation("solc_version", solc_version)
}
