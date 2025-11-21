package validator

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validPathRegex = regexp.MustCompile(`^\/[a-zA-Z0-9\/\-\_\:\*]*$`) // simple valid path check

func httpPath(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	return validPathRegex.MatchString(path) && strings.HasPrefix(path, "/")
}
