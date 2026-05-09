package handler

import (
	"github.com/gin-gonic/gin"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/response"
	"grbac/internal/service"
)

// Context keys set by the auth middleware.
const (
	CtxUserID      = "user_id"
	CtxAccessToken = "access_token"
)

// ---------- Request structs ----------

// LoginRequest holds the credentials for user login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest holds the refresh token for token renewal.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest holds the old and new passwords for a password change.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ---------- AuthHandler ----------

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login authenticates a user and returns a token pair.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInvalidPassword)
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, resp)
}

// Refresh validates a refresh token and issues a new token pair.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrTokenInvalid)
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, resp)
}

// Logout blacklists the current access token.
func (h *AuthHandler) Logout(c *gin.Context) {
	accessToken, exists := c.Get(CtxAccessToken)
	if !exists {
		response.Fail(c, errors.ErrTokenInvalid)
		return
	}

	if err := h.authService.Logout(c.Request.Context(), accessToken.(string)); err != nil {
		respondError(c, err)
		return
	}

	response.OKMessage(c)
}

// ChangePassword verifies the old password and sets a new one.
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get(CtxUserID)
	if !exists {
		response.Fail(c, errors.ErrTokenInvalid)
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInvalidPassword)
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID.(int64), req.OldPassword, req.NewPassword); err != nil {
		respondError(c, err)
		return
	}

	response.OKMessage(c)
}

// ---------- helpers ----------

// respondError sends the appropriate error response based on the error type.
func respondError(c *gin.Context, err error) {
	if appErr, ok := errors.IsAppError(err); ok {
		response.Fail(c, appErr)
		return
	}
	response.InternalError(c)
}
