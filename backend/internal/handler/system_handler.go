package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/response"
	"grbac/internal/service"
)

// ---------- Request structs ----------

// UpdateSystemRequest holds the fields for updating a system.
type UpdateSystemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AddMemberRequest holds the payload for adding a member to a system.
type AddMemberRequest struct {
	UserID int64  `json:"user_id" binding:"required"`
	Role   string `json:"role"    binding:"required"`
}

// ---------- SystemHandler ----------

// SystemHandler handles system and member management HTTP requests.
type SystemHandler struct {
	systemService *service.SystemService
}

// NewSystemHandler creates a new SystemHandler.
func NewSystemHandler(systemService *service.SystemService) *SystemHandler {
	return &SystemHandler{systemService: systemService}
}

// Create registers a new system.
func (h *SystemHandler) Create(c *gin.Context) {
	var req service.CreateSystemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	system, err := h.systemService.Create(&req)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, system)
}

// List returns a paginated list of systems.
func (h *SystemHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	systems, total, err := h.systemService.List(page, pageSize)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OKPage(c, total, systems)
}

// GetByID retrieves a system by its ID.
func (h *SystemHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	system, err := h.systemService.GetByID(id)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, system)
}

// Update modifies the name and description of an existing system.
func (h *SystemHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	var req UpdateSystemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	system, err := h.systemService.Update(id, req.Name, req.Description)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, system)
}

// Delete removes a system by its ID.
func (h *SystemHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	if err := h.systemService.Delete(id); err != nil {
		respondError(c, err)
		return
	}

	response.OKMessage(c)
}

// AddMember adds a user to a system with the specified role.
func (h *SystemHandler) AddMember(c *gin.Context) {
	systemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	if err := h.systemService.AddMember(systemID, req.UserID, req.Role); err != nil {
		respondError(c, err)
		return
	}

	response.OKMessage(c)
}

// RemoveMember removes a user from a system.
func (h *SystemHandler) RemoveMember(c *gin.Context) {
	systemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	userID, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrUserNotFound)
		return
	}

	if err := h.systemService.RemoveMember(systemID, userID); err != nil {
		respondError(c, err)
		return
	}

	response.OKMessage(c)
}

// GetMembers returns all members of a system.
func (h *SystemHandler) GetMembers(c *gin.Context) {
	systemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	members, err := h.systemService.GetMembers(systemID)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, members)
}
