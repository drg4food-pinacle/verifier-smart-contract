package validator

import (
	"os"
	"reflect"

	"github.com/go-playground/validator/v10"
)

func fileExistsIfTls(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	if path == "" {
		return true // Let required_if handle this
	}

	// Param will be the condition field, e.g. "TLSEnabled" or "TLSClientAuthRequired"
	conditionFieldName := fl.Param()
	if conditionFieldName == "" {
		return false
	}

	// Get the parent struct
	parent := fl.Top()
	if parent.Kind() == reflect.Ptr {
		parent = parent.Elem()
	}
	conditionField := parent.FieldByName(conditionFieldName)

	if !conditionField.IsValid() || conditionField.Kind() != reflect.Bool {
		return false
	}

	if !conditionField.Bool() {
		return true // Condition not met → skip file check
	}

	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
