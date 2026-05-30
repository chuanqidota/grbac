package role

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/response"
	roleService "grbac/internal/service/role"
)

// Request structs

type UpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AssignMenusRequest struct {
	MenuIDs []int64 `json:"menu_ids" binding:"required"`
}

type AssignPermissionsRequest struct {
	PermissionIDs []int64 `json:"permission_ids" binding:"required"`
}

type AssignUsersRequest struct {
	UserIDs []int64 `json:"user_ids" binding:"required"`
}

// Handler handles role management HTTP requests.
type Handler struct {
	roleSvc *roleService.Service
}

// NewHandler creates a new Handler.
func NewHandler(roleSvc *roleService.Service) *Handler {
	return &Handler{roleSvc: roleSvc}
}

// Create registers a new role within a system.
func (h *Handler) Create(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	var req roleService.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	role, err := h.roleSvc.Create(c.Request.Context(), sid, &req)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, role)
}

// List returns all roles belonging to a system.
func (h *Handler) List(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	roles, err := h.roleSvc.ListBySystem(sid)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKPage(c, int64(len(roles)), roles)
}

// Update modifies the name and description of an existing role.
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	role, err := h.roleSvc.Update(c.Request.Context(), id, req.Name, req.Description)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, role)
}

// Delete removes a role by its ID.
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	if err := h.roleSvc.Delete(c.Request.Context(), id); err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKMessage(c)
}

// AssignMenus replaces the menu set of a role.
func (h *Handler) AssignMenus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	var req AssignMenusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[DEBUG] AssignMenus bind error: %v", err)
		response.Fail(c, errors.ErrInternal)
		return
	}

	log.Printf("[DEBUG] AssignMenus: roleID=%d, menuIDs=%v", id, req.MenuIDs)
	if err := h.roleSvc.AssignMenus(c.Request.Context(), id, req.MenuIDs); err != nil {
		log.Printf("[DEBUG] AssignMenus service error: %v", err)
		response.RespondError(c, err)
		return
	}

	log.Printf("[DEBUG] AssignMenus success: roleID=%d", id)
	response.OKMessage(c)
}

// GetRoleMenus returns the menu IDs assigned to a role.
func (h *Handler) GetRoleMenus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	menuIDs, err := h.roleSvc.GetRoleMenus(id)
	if err != nil {
		log.Printf("[DEBUG] GetRoleMenus error: roleID=%d, err=%v", id, err)
		response.RespondError(c, err)
		return
	}

	log.Printf("[DEBUG] GetRoleMenus: roleID=%d, count=%d, ids=%v", id, len(menuIDs), menuIDs)
	response.OKPage(c, int64(len(menuIDs)), menuIDs)
}

// AssignPermissions replaces the permission set of a role.
func (h *Handler) AssignPermissions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	if err := h.roleSvc.AssignPermissions(c.Request.Context(), id, req.PermissionIDs); err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKMessage(c)
}

// GetRolePermissions returns the permission IDs assigned to a role.
func (h *Handler) GetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	permIDs, err := h.roleSvc.GetRolePermissions(id)
	if err != nil {
		log.Printf("[DEBUG] GetRolePermissions error: roleID=%d, err=%v", id, err)
		response.RespondError(c, err)
		return
	}

	log.Printf("[DEBUG] GetRolePermissions: roleID=%d, count=%d, ids=%v", id, len(permIDs), permIDs)
	response.OKPage(c, int64(len(permIDs)), permIDs)
}

// AssignUsers replaces the user set of a role.
func (h *Handler) AssignUsers(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	var req AssignUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[DEBUG] AssignUsers bind error: %v", err)
		response.Fail(c, errors.ErrInternal)
		return
	}

	log.Printf("[DEBUG] AssignUsers: roleID=%d, userIDs=%v", id, req.UserIDs)
	if err := h.roleSvc.AssignUsers(c.Request.Context(), id, req.UserIDs); err != nil {
		log.Printf("[DEBUG] AssignUsers service error: %v", err)
		response.RespondError(c, err)
		return
	}

	log.Printf("[DEBUG] AssignUsers success: roleID=%d", id)
	response.OKMessage(c)
}

// RemoveUser removes a single user from a role.
func (h *Handler) RemoveUser(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	userID, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrUserNotFound)
		return
	}

	if err := h.roleSvc.RemoveUser(c.Request.Context(), roleID, userID); err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKMessage(c)
}

// GetRoleUsers returns all users assigned to a role.
func (h *Handler) GetRoleUsers(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrRoleNotFound)
		return
	}

	users, err := h.roleSvc.GetRoleUsers(id)
	if err != nil {
		log.Printf("[DEBUG] GetRoleUsers error: roleID=%d, err=%v", id, err)
		response.RespondError(c, err)
		return
	}

	log.Printf("[DEBUG] GetRoleUsers: roleID=%d, count=%d", id, len(users))
	response.OKPage(c, int64(len(users)), users)
}
