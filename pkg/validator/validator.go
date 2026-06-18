package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FormatBindingError returns a human-readable summary of a Gin binding or
// validation error. For go-playground/validator.ValidationErrors each failing
// field is described individually; for any other error the raw message is
// returned as-is.
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

// isValidationErrors performs a type-assertion to validator.ValidationErrors.
func isValidationErrors(err error, target *validator.ValidationErrors) bool {
	if ve, ok := err.(validator.ValidationErrors); ok {
		*target = ve
		return true
	}
	return false
}

// fieldMessage converts a single FieldError into a readable sentence.
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
