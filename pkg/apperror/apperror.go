// Package apperror provides apperror functionality for the IMB platform.
package apperror

import "net/http"

// AppError returns the status code and message
type AppError struct {
	Code    int
	Message string
}

// Return error message
func (e *AppError) Error() string { return e.Message }

// Creates error with status code and message
func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// BadRequest Error
func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, message)
}

// Unauthorized Error
func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, message)
}

// Forbidden Error
func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, message)
}

// NotFound Error
func NotFound(message string) *AppError {
	return New(http.StatusNotFound, message)
}

// Conflict Error
func Conflict(message string) *AppError {
	return New(http.StatusConflict, message)
}

// Internal Server Error
func Internal() *AppError {
	return New(http.StatusInternalServerError, "internal server error")
}
