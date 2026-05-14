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

// GetUserRoles returns the user's identity and their roles within the requesting system.
func (h *Handler) GetUserRoles(c *gin.Context) {
	systemCode := c.GetHeader("X-System-Code")
	if systemCode == "" {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	username := c.Query("username")
	if username == "" {
		response.FailWithMessage(c, errors.ErrInternal, "缺少必填参数 username")
		return
	}

	userInfo, err := h.externalSvc.GetUserInfoByUsername(c.Request.Context(), username, systemCode)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	response.OK(c, userInfo)
}

// GetMenus returns the menus assigned to the user within the requesting system.
func (h *Handler) GetMenus(c *gin.Context) {
	systemCode := c.GetHeader("X-System-Code")
	if systemCode == "" {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	username := c.Query("username")
	if username == "" {
		response.FailWithMessage(c, errors.ErrInternal, "缺少必填参数 username")
		return
	}

	menus, err := h.externalSvc.GetUserMenusByUsername(c.Request.Context(), username, systemCode)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	response.OK(c, menus)
}

// GetUserAPIs returns the API permissions assigned to the user within the requesting system.
func (h *Handler) GetUserAPIs(c *gin.Context) {
	systemCode := c.GetHeader("X-System-Code")
	if systemCode == "" {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	username := c.Query("username")
	if username == "" {
		response.FailWithMessage(c, errors.ErrInternal, "缺少必填参数 username")
		return
	}

	permissions, err := h.externalSvc.GetUserPermissionsByUsername(c.Request.Context(), username, systemCode)
	if err != nil {
		respondExternalError(c, err)
		return
	}

	response.OK(c, permissions)
}

// CheckPermission checks whether the user holds a specific permission (method + path).
func (h *Handler) CheckPermission(c *gin.Context) {
	systemCode := c.GetHeader("X-System-Code")
	if systemCode == "" {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	username := c.Query("username")
	if username == "" {
		response.FailWithMessage(c, errors.ErrInternal, "缺少必填参数 username")
		return
	}

	method := c.Query("method")
	path := c.Query("path")
	if method == "" || path == "" {
		response.Fail(c, errors.ErrPermNotFound)
		return
	}

	valid, err := h.externalSvc.ValidatePermissionByUsername(c.Request.Context(), username, systemCode, method, path)
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
