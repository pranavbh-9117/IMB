package dto

// LoginRequest carries the credentials for email+password authentication.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// ChangePasswordRequest carries the current and replacement passwords.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UserResponse is the user profile fragment included in the login response.
type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// LoginResponse is the data payload returned by POST /auth/login.
// The refresh token is delivered via an HttpOnly cookie and is absent from
// this struct.
type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

// RefreshResponse is the data payload returned by POST /auth/refresh.
// The new refresh token is delivered via an HttpOnly cookie.
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}
