// Package handler provides handler functionality for the IMB platform.
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

type AuthHandler struct {
	svc service.AuthService
	cfg config.JWTConfig
}

func NewAuthHandler(svc service.AuthService, cfg config.JWTConfig) *AuthHandler {
	return &AuthHandler{svc: svc, cfg: cfg}
}

// Login godoc
// @Summary User Login
// @Description Authenticates a user and returns a JWT access token in the body and a refresh token in an HttpOnly cookie.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Credentials"
// @Success 200 {object} response.SwaggerResponse[dto.LoginResponse] "Login Successful"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Router /auth/login [post]
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
// @Summary Refresh Token
// @Description Reads the refresh token from the HttpOnly cookie, rotates it, and returns a new access token.
// @Tags Authentication
// @Produce json
// @Success 200 {object} response.SwaggerResponse[dto.RefreshResponse] "Token Refreshed"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Router /auth/refresh [post]
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
// @Summary User Logout
// @Description Revokes the refresh token and clears the HttpOnly cookie.
// @Tags Authentication
// @Produce json
// @Success 200 {object} response.SwaggerResponse[any] "Logout Successful"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Router /auth/logout [post]
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
// @Summary Change Password
// @Description Changes the authenticated user's password and revokes all active sessions.
// @Tags Authentication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ChangePasswordRequest true "Change Password Payload"
// @Success 200 {object} response.SwaggerResponse[any] "Password Changed Successfully"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Router /auth/change-password [post]
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


func (h *AuthHandler) clearRefreshTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshTokenCookie, "", -1, "/", "", false, true)
}


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
