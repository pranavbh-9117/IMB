package response

// SwaggerResponse is a generic wrapper exclusively used for Swaggo OpenAPI generation
// to represent the standard JSON envelope with typed data.
type SwaggerResponse[T any] struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"operation successful"`
	Data    T      `json:"data"`
}

// SwaggerErrorResponse is exclusively used for Swaggo OpenAPI generation
// to represent the standard JSON error envelope.
type SwaggerErrorResponse struct {
	Success bool        `json:"success" example:"false"`
	Message string      `json:"message" example:"error description"`
	Data    interface{} `json:"data"`
}
