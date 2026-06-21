// Package response provides response functionality for the IMB platform.
package response

// SwaggerResponse 
type SwaggerResponse[T any] struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"operation successful"`
	Data    T      `json:"data"`
}

// SwaggerErrorResponse 
type SwaggerErrorResponse struct {
	Success bool        `json:"success" example:"false"`
	Message string      `json:"message" example:"error description"`
	Data    interface{} `json:"data"`
}
