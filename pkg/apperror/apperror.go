// Package apperror provides apperror functionality for the IMB platform.
package apperror

import "net/http"

// AppError is a structured error carrying an HTTP status code and a
// human-readable message. It is used by pkg/response to produce consistent
// error envelopes and can be returned from any layer that needs to signal
// a specific HTTP status to the handler.
type AppError struct {
	Code    int
	Message string
}

// Error implements the corresponding interface or provides the named functionality.
func (e *AppError) Error() string { return e.Message }

// New creates an AppError with the given HTTP status code and message.
func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Common constructors for frequently used HTTP error statuses.

// BadRequest implements the corresponding interface or provides the named functionality.
func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, message)
}

// Unauthorized implements the corresponding interface or provides the named functionality.
func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, message)
}

// Forbidden implements the corresponding interface or provides the named functionality.
func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, message)
}

// NotFound implements the corresponding interface or provides the named functionality.
func NotFound(message string) *AppError {
	return New(http.StatusNotFound, message)
}

// Conflict implements the corresponding interface or provides the named functionality.
func Conflict(message string) *AppError {
	return New(http.StatusConflict, message)
}

// Internal implements the corresponding interface or provides the named functionality.
func Internal() *AppError {
	return New(http.StatusInternalServerError, "internal server error")
}
