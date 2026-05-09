package external

import (
	"github.com/gin-gonic/gin"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/response"
	externalService "grbac/internal/service/external"
)

// Handler handles API requests from external systems.
type Handler struct {
	externalSvc *externalService.Service
}

// NewHandler creates a new Handler.
func NewHandler(externalSvc *externalService.Service) *Handler {
	return &Handler{externalSvc: externalSvc}
}

// Verify validates the JWT carried in the X-User-Token header.
func (h *Handler) Verify(c *gin.Context) {
	token := c.GetHeader("X-User-Token")
	if token == "" {
		response.Fail(c, errors.ErrTokenInvalid)
		return
	}

	claims, err := h.externalSvc.VerifyToken(c.Request.Context(), token)
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

// GetUserInfo returns the user's identity and their roles within the requesting system.
func (h *Handler) GetUserInfo(c *gin.Context) {
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

	claims, err := h.externalSvc.VerifyToken(c.Request.Context(), token)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	userInfo, err := h.externalSvc.GetUserInfo(c.Request.Context(), int64(claims.UserID), systemCode)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	response.OK(c, userInfo)
}

// GetMenus returns the menus assigned to the user within the requesting system.
func (h *Handler) GetMenus(c *gin.Context) {
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

	claims, err := h.externalSvc.VerifyToken(c.Request.Context(), token)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	menus, err := h.externalSvc.GetUserMenus(c.Request.Context(), int64(claims.UserID), systemCode)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	response.OK(c, menus)
}

// GetPermissions returns the permission codes assigned to the user within the requesting system.
func (h *Handler) GetPermissions(c *gin.Context) {
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

	claims, err := h.externalSvc.VerifyToken(c.Request.Context(), token)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	permissions, err := h.externalSvc.GetUserPermissions(c.Request.Context(), int64(claims.UserID), systemCode)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	response.OK(c, permissions)
}

// ValidatePermission checks whether the user holds a specific permission (method + path).
func (h *Handler) ValidatePermission(c *gin.Context) {
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

	method := c.Query("method")
	path := c.Query("path")
	if method == "" || path == "" {
		response.Fail(c, errors.ErrPermNotFound)
		return
	}

	claims, err := h.externalSvc.VerifyToken(c.Request.Context(), token)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	valid, err := h.externalSvc.ValidatePermission(c.Request.Context(), int64(claims.UserID), systemCode, method, path)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	response.OK(c, gin.H{"valid": valid})
}

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
