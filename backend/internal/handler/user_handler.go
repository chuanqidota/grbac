package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/response"
	"grbac/internal/service"
)

// ---------- Request structs ----------

// StatusUpdateRequest holds the new status value for a user.
type StatusUpdateRequest struct {
	Status int8 `json:"status" binding:"required"`
}

// ---------- UserHandler ----------

// UserHandler handles user CRUD HTTP requests.
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Create registers a new user.
func (h *UserHandler) Create(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrUsernameExists)
		return
	}

	user, err := h.userService.Create(&req)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, user)
}

// GetByID retrieves a user by their ID.
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrUserNotFound)
		return
	}

	user, err := h.userService.GetByID(id)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, user)
}

// List returns a paginated list of users.
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := h.userService.List(page, pageSize)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OKPage(c, total, users)
}

// Update modifies editable profile fields of an existing user.
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrUserNotFound)
		return
	}

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	user, err := h.userService.Update(id, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, user)
}

// Delete removes a user by their ID.
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrUserNotFound)
		return
	}

	if err := h.userService.Delete(id); err != nil {
		respondError(c, err)
		return
	}

	response.OKMessage(c)
}

// UpdateStatus changes the active/disabled status of a user.
func (h *UserHandler) UpdateStatus(c *gin.Context) {
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

	if err := h.userService.UpdateStatus(id, req.Status); err != nil {
		respondError(c, err)
		return
	}

	response.OKMessage(c)
}
