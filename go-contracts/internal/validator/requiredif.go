package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// requiredIfValidator checks if a field is required when another field meets a certain condition.
func requiredIf(fl validator.FieldLevel) bool {
	param := fl.Param() // e.g., "TLSEnabled true"
	parts := strings.SplitN(param, " ", 2)
	if len(parts) < 2 {
		return false
	}

	conditionField := parts[0]
	conditionValue := parts[1]

	structVal := fl.Parent()

	condField := structVal.FieldByName(conditionField)
	if !condField.IsValid() {
		return false
	}

	// Support bool, int, string comparisons
	switch condField.Kind() {
	case reflect.Bool:
		expected := conditionValue == "true"
		if condField.Bool() == expected && fl.Field().String() == "" {
			return false
		}
	case reflect.String:
		if condField.String() == conditionValue && fl.Field().String() == "" {
			return false
		}
	case reflect.Int, reflect.Int64:
		// Optional: convert conditionValue to int and compare
	}

	return true
}
