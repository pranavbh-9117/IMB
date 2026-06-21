// Package validator provides validator functionality for the IMB platform.
package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Returns Human Readable Errors
func FormatBindingError(err error) string {
	var ve validator.ValidationErrors
	if ok := isValidationErrors(err, &ve); ok {
		msgs := make([]string, 0, len(ve))
		for _, fe := range ve {
			msgs = append(msgs, fieldMessage(fe))
		}
		return strings.Join(msgs, "; ")
	}

	return err.Error()
}

// Check error is of type ValidationError
func isValidationErrors(err error, target *validator.ValidationErrors) bool {
	if ve, ok := err.(validator.ValidationErrors); ok {
		*target = ve
		return true
	}
	return false
}

// convert field error into message
func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("%s is invalid (%s)", fe.Field(), fe.Tag())
	}
}
