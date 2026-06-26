// Package handler provides handler functionality for the IMB platform.
package handler

import (
	"errors"
	"net/http"
	"strings"

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

// ForgotPassword godoc
// @Summary Forgot Password
// @Description Sends a password reset email if the email is registered.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "User Email"
// @Success 200 {object} response.SwaggerResponse[any] "Reset Link Sent"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.FormatBindingError(err))
		return
	}

	if err := h.svc.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.OK(c, "if the email exists, a password reset link has been sent", nil)
}

// ResetPassword godoc
// @Summary Reset Password
// @Description Resets user password using a valid reset token.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Token and New Password"
// @Success 200 {object} response.SwaggerResponse[any] "Password Reset Successfully"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.FormatBindingError(err))
		return
	}

	if err := h.svc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.clearRefreshTokenCookie(c)

	response.OK(c, "password reset successfully", nil)
}





// GoogleLogin godoc
// @Summary Google OAuth Login
// @Description Initiates the Google OAuth authorization code flow.
// @Tags Authentication
// @Produce json
// @Success 302 "Redirects to Google"
// @Router /auth/google/login [get]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url, state := h.svc.GetGoogleLoginURL()

	secure := false

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"oauth_state",
		state,
		600,
		"/",
		"",
		secure,
		true,
	)

	c.Redirect(http.StatusFound, url)
}

// GoogleCallback godoc
// @Summary Google OAuth Callback
// @Description Handles the callback from Google and returns standard login response.
// @Tags Authentication
// @Produce json
// @Param code query string true "OAuth Authorization Code"
// @Param state query string true "OAuth State Parameter"
// @Success 200 {object} response.SwaggerResponse[dto.LoginResponse] "Google Login Successful"
// @Failure 400 {object} response.SwaggerErrorResponse "Bad Request"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized"
// @Failure 403 {object} response.SwaggerErrorResponse "Forbidden"
// @Router /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	cookieState, err := c.Cookie("oauth_state")
	if err != nil || state != cookieState {
		response.BadRequest(c, "invalid oauth state")
		return
	}

	// Delete cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	result, err := h.svc.GoogleCallback(c.Request.Context(), code)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	h.setRefreshTokenCookie(c, result.RefreshToken)

	response.OK(c, "Google login successful", dto.LoginResponse{
		AccessToken: result.AccessToken,
		User: dto.UserResponse{
			ID:    result.User.ID,
			Name:  result.User.Name,
			Email: result.User.Email,
			Role:  result.User.Role,
		},
	})
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
	case errors.Is(err, service.ErrInvalidResetToken):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrWrongPassword):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrGoogleEmailUnverified):
		response.Forbidden(c, err.Error())
	case errors.Is(err, service.ErrAccountNotProvisioned):
		response.Forbidden(c, err.Error())
	case errors.Is(err, service.ErrGoogleProfileMismatch):
		response.Forbidden(c, err.Error())
	case strings.Contains(err.Error(), "validate id_token") || strings.Contains(err.Error(), "invalid issuer") || strings.Contains(err.Error(), "no id_token in response"):
		response.Unauthorized(c, "invalid google identity token")
	default:
		response.InternalServerError(c)
	}
}
