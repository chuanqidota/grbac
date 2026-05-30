package user

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"grbac/internal/middleware"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/response"
	userRepo "grbac/internal/repository/user"
	userService "grbac/internal/service/user"
)

// StatusUpdateRequest holds the new status value for a user.
type StatusUpdateRequest struct {
	Status int8 `json:"status" binding:"required"`
}

// SuperAdminUpdateRequest holds the new super-admin flag for a user.
type SuperAdminUpdateRequest struct {
	IsSuperAdmin int8 `json:"is_super_admin"`
}

// ResetPasswordRequest holds the new password for a user.
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// Handler handles user CRUD HTTP requests.
type Handler struct {
	userSvc *userService.Service
}

// NewHandler creates a new Handler.
func NewHandler(userSvc *userService.Service) *Handler {
	return &Handler{userSvc: userSvc}
}

// GetMe returns the current authenticated user's info.
func (h *Handler) GetMe(c *gin.Context) {
	userID, exists := c.Get(middleware.CtxUserID)
	if !exists {
		response.Unauthorized(c, errors.ErrTokenInvalid)
		return
	}

	user, err := h.userSvc.GetByID(userID.(int64))
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, user)
}

// Create registers a new user.
func (h *Handler) Create(c *gin.Context) {
	var req userService.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrUsernameExists)
		return
	}

	user, err := h.userSvc.Create(&req)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, user)
}

// GetByID retrieves a user by their ID.
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrUserNotFound)
		return
	}

	user, err := h.userSvc.GetByID(id)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, user)
}

// List returns a paginated list of users.
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Optional filter: is_super_admin (0 or 1)
	var q *userRepo.ListQuery
	if v := c.Query("is_super_admin"); v != "" {
		val, err := strconv.Atoi(v)
		if err == nil && (val == 0 || val == 1) {
			q = &userRepo.ListQuery{IsSuperAdmin: &val}
		}
	}

	users, total, err := h.userSvc.List(page, pageSize, q)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKPage(c, total, users)
}

// Update modifies editable profile fields of an existing user.
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrUserNotFound)
		return
	}

	var req userService.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	user, err := h.userSvc.Update(id, &req)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, user)
}

// Delete removes a user by their ID.
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrUserNotFound)
		return
	}

	if err := h.userSvc.Delete(id); err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKMessage(c)
}

// UpdateStatus changes the active/disabled status of a user.
func (h *Handler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrUserNotFound)
		return
	}

	var req StatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	if err := h.userSvc.UpdateStatus(id, req.Status); err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKMessage(c)
}

// ResetPassword changes a user's password (admin operation).
func (h *Handler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrUserNotFound)
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	if err := h.userSvc.ResetPassword(id, req.NewPassword); err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKMessage(c)
}

// UpdateSuperAdmin sets or unsets the super-admin flag for a user.
func (h *Handler) UpdateSuperAdmin(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrUserNotFound)
		return
	}

	var req SuperAdminUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	if req.IsSuperAdmin != 0 && req.IsSuperAdmin != 1 {
		response.Fail(c, errors.ErrInternal.Wrap("is_super_admin 必须为 0 或 1"))
		return
	}

	if err := h.userSvc.UpdateSuperAdmin(id, req.IsSuperAdmin); err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKMessage(c)
}

// GetUserRoles returns all roles assigned to a user across all systems.
func (h *Handler) GetUserRoles(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrUserNotFound)
		return
	}

	roles, err := h.userSvc.GetUserRoles(id)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKPage(c, int64(len(roles)), roles)
}
