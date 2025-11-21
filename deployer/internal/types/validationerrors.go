package types

import "regexp"

var indexPattern = regexp.MustCompile(`\[\d+\]`)

// Interface for types that provide custom validation errors
type CustomValidationErrors interface {
	CustomErrorMessages() map[string]string
}

// NormalizeFieldPath replaces numeric indices with []
func NormalizeFieldPath(path string) string {
	return indexPattern.ReplaceAllString(path, "[]")
}
