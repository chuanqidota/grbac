package handler

import (
	"github.com/gin-gonic/gin"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/response"
	"grbac/internal/service"
)

// Context keys set by external auth middleware.
const (
	CtxExtSystemID   = "system_id"
	CtxExtSystemCode = "system_code"
)

// ---------- Request structs ----------

// ValidatePermissionRequest holds the permission code to validate.
type ValidatePermissionRequest struct {
	PermissionCode string `json:"permission_code" binding:"required"`
}

// ---------- ExternalHandler ----------

// ExternalHandler handles API requests from external systems.
type ExternalHandler struct {
	externalService *service.ExternalService
}

// NewExternalHandler creates a new ExternalHandler.
func NewExternalHandler(externalService *service.ExternalService) *ExternalHandler {
	return &ExternalHandler{externalService: externalService}
}

// Verify validates the JWT carried in the X-User-Token header and
// returns the parsed claims.
func (h *ExternalHandler) Verify(c *gin.Context) {
	token := c.GetHeader("X-User-Token")
	if token == "" {
		response.Fail(c, errors.ErrTokenInvalid)
		return
	}

	claims, err := h.externalService.VerifyToken(c.Request.Context(), token)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	response.OK(c, gin.H{
		"user_id":        claims.UserID,
		"username":       claims.Username,
		"is_super_admin": claims.IsSuperAdmin,
	})
}

// GetUserInfo returns the user's identity and their roles within the
// requesting system.
func (h *ExternalHandler) GetUserInfo(c *gin.Context) {
	token := c.GetHeader("X-User-Token")
	if token == "" {
		response.Fail(c, errors.ErrTokenInvalid)
		return
	}

	systemCode := c.GetHeader("X-System-Code")
	if systemCode == "" {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	// Validate token first.
	claims, err := h.externalService.VerifyToken(c.Request.Context(), token)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	userInfo, err := h.externalService.GetUserInfo(c.Request.Context(), int64(claims.UserID), systemCode)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	response.OK(c, userInfo)
}

// GetMenus returns the menus assigned to the user within the requesting system.
func (h *ExternalHandler) GetMenus(c *gin.Context) {
	token := c.GetHeader("X-User-Token")
	if token == "" {
		response.Fail(c, errors.ErrTokenInvalid)
		return
	}

	systemCode := c.GetHeader("X-System-Code")
	if systemCode == "" {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	claims, err := h.externalService.VerifyToken(c.Request.Context(), token)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	menus, err := h.externalService.GetUserMenus(c.Request.Context(), int64(claims.UserID), systemCode)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	response.OK(c, menus)
}

// GetPermissions returns the permission codes assigned to the user
// within the requesting system.
func (h *ExternalHandler) GetPermissions(c *gin.Context) {
	token := c.GetHeader("X-User-Token")
	if token == "" {
		response.Fail(c, errors.ErrTokenInvalid)
		return
	}

	systemCode := c.GetHeader("X-System-Code")
	if systemCode == "" {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	claims, err := h.externalService.VerifyToken(c.Request.Context(), token)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	permissions, err := h.externalService.GetUserPermissions(c.Request.Context(), int64(claims.UserID), systemCode)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	response.OK(c, permissions)
}

// ValidatePermission checks whether the user holds a specific permission
// code within the requesting system. Returns { "valid": true/false }.
func (h *ExternalHandler) ValidatePermission(c *gin.Context) {
	token := c.GetHeader("X-User-Token")
	if token == "" {
		response.Fail(c, errors.ErrTokenInvalid)
		return
	}

	systemCode := c.GetHeader("X-System-Code")
	if systemCode == "" {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	var req ValidatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrPermNotFound)
		return
	}

	claims, err := h.externalService.VerifyToken(c.Request.Context(), token)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	permissions, err := h.externalService.GetUserPermissions(c.Request.Context(), int64(claims.UserID), systemCode)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	valid := false
	for _, p := range permissions {
		if p == req.PermissionCode {
			valid = true
			break
		}
	}

	response.OK(c, gin.H{"valid": valid})
}

// ---------- helpers ----------

// respondExternalError maps service errors to the appropriate HTTP response.
func respondExternalError(c *gin.Context, err error) {
	if appErr, ok := errors.IsAppError(err); ok {
		switch appErr {
		case errors.ErrTokenExpired, errors.ErrTokenInvalid:
			response.Unauthorized(c, appErr)
		case errors.ErrNoPermission:
			response.Forbidden(c, appErr)
		default:
			response.Fail(c, appErr)
		}
		return
	}
	response.InternalError(c)
}
