// Package response provides response functionality for the IMB platform.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// envelope is the standard JSON response shape used by every endpoint.
type envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// OK writes a 200 response with the given message and data payload.
func OK(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, envelope{Success: true, Message: message, Data: data})
}

// Created writes a 201 response with the given message and data payload.
func Created(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, envelope{Success: true, Message: message, Data: data})
}

// BadRequest writes a 400 response with the given error message.
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, envelope{Success: false, Message: message, Data: nil})
}

// Unauthorized writes a 401 response with the given error message.
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, envelope{Success: false, Message: message, Data: nil})
}

// Forbidden writes a 403 response with the given error message.
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, envelope{Success: false, Message: message, Data: nil})
}

// NotFound writes a 404 response with the given error message.
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, envelope{Success: false, Message: message, Data: nil})
}

// Conflict writes a 409 response with the given error message.
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, envelope{Success: false, Message: message, Data: nil})
}

// InternalServerError writes a 500 response with a generic message.
func InternalServerError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, envelope{
		Success: false,
		Message: "internal server error",
		Data:    nil,
	})
}
