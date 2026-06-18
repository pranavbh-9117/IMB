package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pranavbh-9117/IMB/internal/auth/dto"
	"github.com/pranavbh-9117/IMB/internal/auth/service"
	"github.com/pranavbh-9117/IMB/internal/middleware"
	"github.com/pranavbh-9117/IMB/pkg/config"
	"github.com/pranavbh-9117/IMB/pkg/response"
	"github.com/pranavbh-9117/IMB/pkg/validator"
)

const refreshTokenCookie = "refresh_token"

// AuthHandler holds the auth service and JWT config needed by all four
// authentication endpoints.
type AuthHandler struct {
	svc service.AuthService
	cfg config.JWTConfig
}

// NewAuthHandler constructs an AuthHandler. It returns a concrete pointer
// so routes can register individual handler methods.
func NewAuthHandler(svc service.AuthService, cfg config.JWTConfig) *AuthHandler {
	return &AuthHandler{svc: svc, cfg: cfg}
}

// Login godoc
// POST /api/v1/auth/login
// Verifies email+password, issues an access token in the JSON body and a
// refresh token in an HttpOnly cookie.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.FormatBindingError(err))
		return
	}

	result, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.setRefreshTokenCookie(c, result.RefreshToken)

	response.OK(c, "login successful", dto.LoginResponse{
		AccessToken: result.AccessToken,
		User: dto.UserResponse{
			ID:    result.User.ID,
			Name:  result.User.Name,
			Email: result.User.Email,
			Role:  result.User.Role,
		},
	})
}

// Refresh godoc
// POST /api/v1/auth/refresh
// Reads the refresh token from the HttpOnly cookie, rotates it, and returns
// a new access token in the JSON body and a new refresh token in the cookie.
func (h *AuthHandler) Refresh(c *gin.Context) {
	rawToken, err := c.Cookie(refreshTokenCookie)
	if err != nil {
		response.Unauthorized(c, "refresh token cookie is missing")
		return
	}

	pair, err := h.svc.Refresh(c.Request.Context(), rawToken)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.setRefreshTokenCookie(c, pair.RefreshToken)

	response.OK(c, "token refreshed", dto.RefreshResponse{
		AccessToken: pair.AccessToken,
	})
}

// Logout godoc
// POST /api/v1/auth/logout
// Reads the refresh token from the HttpOnly cookie, revokes it, and clears
// the cookie.
func (h *AuthHandler) Logout(c *gin.Context) {
	rawToken, err := c.Cookie(refreshTokenCookie)
	if err != nil {
		response.Unauthorized(c, "refresh token cookie is missing")
		return
	}

	if err := h.svc.Logout(c.Request.Context(), rawToken); err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.clearRefreshTokenCookie(c)

	response.OK(c, "logged out successfully", nil)
}

// ChangePassword godoc
// POST /api/v1/auth/change-password
// Delegates password replacement to the auth service. The caller's identity
// must be populated in the context by the RequireAuth middleware. All active
// sessions are revoked on success.
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.FormatBindingError(err))
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.clearRefreshTokenCookie(c)

	response.OK(c, "password changed successfully", nil)
}

// --- private helpers ---

// setRefreshTokenCookie writes the refresh token as an HttpOnly cookie.
// Secure is set to false for local development; set to true in production
// when serving over HTTPS.
func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, rawToken string) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		refreshTokenCookie,
		rawToken,
		int(h.cfg.RefreshExpiry.Seconds()),
		"/",
		"",
		false,
		true,
	)
}

// clearRefreshTokenCookie expires the refresh token cookie immediately.
func (h *AuthHandler) clearRefreshTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshTokenCookie, "", -1, "/", "", false, true)
}

// handleServiceError maps auth service sentinel errors to HTTP responses.
// Unknown errors are logged as internal server errors without leaking details.
func (h *AuthHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		response.Unauthorized(c, err.Error())
	case errors.Is(err, service.ErrAccountInactive):
		response.Forbidden(c, err.Error())
	case errors.Is(err, service.ErrTokenInvalid):
		response.Unauthorized(c, err.Error())
	case errors.Is(err, service.ErrWrongPassword):
		response.BadRequest(c, err.Error())
	default:
		response.InternalServerError(c)
	}
}
