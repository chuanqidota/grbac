package system

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/response"
	systemService "grbac/internal/service/system"
)

// Request structs

type UpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AddMemberRequest struct {
	UserID int64  `json:"user_id" binding:"required"`
	Role   string `json:"role"    binding:"required"`
}

// Handler handles system and member management HTTP requests.
type Handler struct {
	systemSvc *systemService.Service
}

// NewHandler creates a new Handler.
func NewHandler(systemSvc *systemService.Service) *Handler {
	return &Handler{systemSvc: systemSvc}
}

// Create registers a new system.
func (h *Handler) Create(c *gin.Context) {
	var req systemService.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, errors.ErrInternal, "请求参数无效: "+err.Error())
		return
	}

	system, err := h.systemSvc.Create(&req)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, system)
}

// List returns a paginated list of systems.
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	systems, total, err := h.systemSvc.List(page, pageSize)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKPage(c, total, systems)
}

// GetByID retrieves a system by its ID.
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	system, err := h.systemSvc.GetByID(id)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, system)
}

// Update modifies the name and description of an existing system.
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, errors.ErrInternal, "请求参数无效: "+err.Error())
		return
	}

	system, err := h.systemSvc.Update(id, req.Name, req.Description)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, system)
}

// Delete removes a system by its ID.
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	if err := h.systemSvc.Delete(id); err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKMessage(c)
}

// AddMember adds a user to a system with the specified role.
func (h *Handler) AddMember(c *gin.Context) {
	systemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, errors.ErrInternal, "请求参数无效: "+err.Error())
		return
	}

	if err := h.systemSvc.AddMember(systemID, req.UserID, req.Role); err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKMessage(c)
}

// RemoveMember removes a user from a system.
func (h *Handler) RemoveMember(c *gin.Context) {
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

	if err := h.systemSvc.RemoveMember(systemID, userID); err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKMessage(c)
}

// GetMembers returns all members of a system.
func (h *Handler) GetMembers(c *gin.Context) {
	systemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	members, err := h.systemSvc.GetMembers(systemID)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKPage(c, int64(len(members)), members)
}

// GetMemberRoles returns the RBAC roles of a member within a system.
func (h *Handler) GetMemberRoles(c *gin.Context) {
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

	roles, err := h.systemSvc.GetMemberRoles(systemID, userID)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, roles)
}
