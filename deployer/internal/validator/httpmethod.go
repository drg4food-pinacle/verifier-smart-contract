package validator

import (
	"net/http"

	"github.com/go-playground/validator/v10"
)

func httpMethod(fl validator.FieldLevel) bool {
	method := fl.Field().String()
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete,
		http.MethodPatch, http.MethodHead, http.MethodOptions, http.MethodConnect, http.MethodTrace:
		return true
	default:
		return false
	}
}
