package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/response"
	"github.com/jinang/grbac/internal/service"
)

// ---------- Request structs ----------

// UpdateRoleRequest holds the fields for updating a role.
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AssignMenusRequest holds the payload for assigning menus to a role.
type AssignMenusRequest struct {
	MenuIDs []int64 `json:"menu_ids" binding:"required"`
}

// AssignPermissionsRequest holds the payload for assigning permissions to a role.
type AssignPermissionsRequest struct {
	PermissionIDs []int64 `json:"permission_ids" binding:"required"`
}

// AssignUsersRequest holds the payload for assigning users to a role.
type AssignUsersRequest struct {
	UserIDs []int64 `json:"user_ids" binding:"required"`
}

// ---------- RoleHandler ----------

// RoleHandler handles role management HTTP requests.
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler creates a new RoleHandler.
func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// Create registers a new role within a system.
func (h *RoleHandler) Create(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	role, err := h.roleService.Create(sid, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, role)
}

// List returns all roles belonging to a system.
func (h *RoleHandler) List(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	roles, err := h.roleService.ListBySystem(sid)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, roles)
}

// Update modifies the name and description of an existing role.
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	role, err := h.roleService.Update(id, req.Name, req.Description)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, role)
}

// Delete removes a role by its ID.
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	if err := h.roleService.Delete(id); err != nil {
		respondError(c, err)
		return
	}

	response.OKMessage(c)
}

// AssignMenus replaces the menu set of a role.
func (h *RoleHandler) AssignMenus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	var req AssignMenusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	if err := h.roleService.AssignMenus(id, req.MenuIDs); err != nil {
		respondError(c, err)
		return
	}

	response.OKMessage(c)
}

// GetRoleMenus returns the menu IDs assigned to a role.
func (h *RoleHandler) GetRoleMenus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	menuIDs, err := h.roleService.GetRoleMenus(id)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, menuIDs)
}

// AssignPermissions replaces the permission set of a role.
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	if err := h.roleService.AssignPermissions(id, req.PermissionIDs); err != nil {
		respondError(c, err)
		return
	}

	response.OKMessage(c)
}

// GetRolePermissions returns the permission IDs assigned to a role.
func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	permIDs, err := h.roleService.GetRolePermissions(id)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, permIDs)
}

// AssignUsers replaces the user set of a role.
func (h *RoleHandler) AssignUsers(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	var req AssignUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	if err := h.roleService.AssignUsers(id, req.UserIDs); err != nil {
		respondError(c, err)
		return
	}

	response.OKMessage(c)
}

// RemoveUser removes a single user from a role.
func (h *RoleHandler) RemoveUser(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	userID, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrUserNotFound)
		return
	}

	if err := h.roleService.RemoveUser(roleID, userID); err != nil {
		respondError(c, err)
		return
	}

	response.OKMessage(c)
}
