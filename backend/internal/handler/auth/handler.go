package auth

import (
	"github.com/gin-gonic/gin"
	"grbac/internal/middleware"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/response"
	authService "grbac/internal/service/auth"
)

// Request structs

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// Handler handles authentication-related HTTP requests.
type Handler struct {
	authSvc *authService.Service
}

// NewHandler creates a new Handler.
func NewHandler(authSvc *authService.Service) *Handler {
	return &Handler{authSvc: authSvc}
}

// Login authenticates a user and returns a token pair.
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInvalidPassword)
		return
	}

	resp, err := h.authSvc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, resp)
}

// Refresh validates a refresh token and issues a new token pair.
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrTokenInvalid)
		return
	}

	resp, err := h.authSvc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, resp)
}

// Logout blacklists the current access token.
func (h *Handler) Logout(c *gin.Context) {
	accessToken, exists := c.Get(middleware.CtxAccessToken)
	if !exists {
		response.Fail(c, errors.ErrTokenInvalid)
		return
	}

	if err := h.authSvc.Logout(c.Request.Context(), accessToken.(string)); err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKMessage(c)
}

// ChangePassword verifies the old password and sets a new one.
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxUserID)
	if !exists {
		response.Fail(c, errors.ErrTokenInvalid)
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInvalidPassword)
		return
	}

	if err := h.authSvc.ChangePassword(c.Request.Context(), userID.(int64), req.OldPassword, req.NewPassword); err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKMessage(c)
}
