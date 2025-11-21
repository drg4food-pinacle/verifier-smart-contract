package validator

import (
	"deployer/internal/logger"
	"deployer/internal/types"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// func ValidateStruct[T any](s T) error {
// 	if err := validate.Struct(s); err != nil {
// 		if e, ok := err.(validator.ValidationErrors); ok {
// 			for _, ve := range e {
// 				// TODO
// 				// ! Probably needs better error logging
// 				logger.Logger.Error().
// 					Str("field", ve.StructField()).
// 					Str("rule", ve.Tag()).
// 					Str("value", fmt.Sprintf("%v", ve.Param())).
// 					Msg("Validation failed")
// 			}
// 			return errors.New("Validation Failed")
// 		}
// 		return fmt.Errorf("Validation Error: %w", err)
// 	}
// 	return nil
// }

func ValidateStruct[T any](s T) error {
	if err := validate.Struct(s); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			// Try to get custom error map from struct
			var customErrors map[string]string
			if e, ok := any(s).(types.CustomValidationErrors); ok {
				customErrors = e.CustomErrorMessages()
			}

			for _, fe := range ve {
				structPath := fe.StructNamespace()
				normalizedPath := types.NormalizeFieldPath(structPath)
				key := fmt.Sprintf("%s.%s", normalizedPath, fe.Tag())

				var msg string
				if customErrors != nil {
					if val, exists := customErrors[key]; exists {
						msg = val
					}
				}

				logger.Logger.Error().
					Str("field", structPath).
					Str("rule", fe.Tag()).
					Str("value", fmt.Sprintf("%v", fe.Value())).
					Msg(msg)
			}
			return errors.New("Validation Failed")
		}
		return fmt.Errorf("Validation Error: %w", err)
	}
	return nil
}
