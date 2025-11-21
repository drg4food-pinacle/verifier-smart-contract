package validator

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

func notNil(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.Func {
		return true // Not applicable
	}
	return !fl.Field().IsNil()
}
