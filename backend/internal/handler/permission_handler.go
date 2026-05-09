package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/response"
	"github.com/jinang/grbac/internal/service"
)

// ---------- PermissionHandler ----------

// PermissionHandler handles permission management HTTP requests.
type PermissionHandler struct {
	permService *service.PermissionService
}

// NewPermissionHandler creates a new PermissionHandler.
func NewPermissionHandler(permService *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{permService: permService}
}

// Create registers a new API permission within a system.
func (h *PermissionHandler) Create(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("system_id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	var req service.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	perm, err := h.permService.Create(sid, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, perm)
}

// List returns a paginated list of permissions belonging to a system.
func (h *PermissionHandler) List(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("system_id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	perms, total, err := h.permService.ListBySystem(sid, page, pageSize)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OKPage(c, total, perms)
}

// Update modifies the fields of an existing permission.
func (h *PermissionHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrPermNotFound)
		return
	}

	var req service.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	perm, err := h.permService.Update(id, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, perm)
}

// Delete removes a permission by its ID.
func (h *PermissionHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrPermNotFound)
		return
	}

	if err := h.permService.Delete(id); err != nil {
		respondError(c, err)
		return
	}

	response.OKMessage(c)
}
