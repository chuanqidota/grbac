package permission

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/response"
	permService "grbac/internal/service/permission"
)

// Handler handles permission management HTTP requests.
type Handler struct {
	permSvc *permService.Service
}

// NewHandler creates a new Handler.
func NewHandler(permSvc *permService.Service) *Handler {
	return &Handler{permSvc: permSvc}
}

// Create registers a new API permission within a system.
func (h *Handler) Create(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	var req permService.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	perm, err := h.permSvc.Create(sid, &req)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, perm)
}

// List returns a paginated list of permissions belonging to a system.
func (h *Handler) List(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 10
	}

	method := c.Query("method")
	keyword := c.Query("keyword")

	perms, total, err := h.permSvc.ListBySystem(sid, page, pageSize, method, keyword)
	if err != nil {
		log.Printf("[DEBUG] Permission List error: systemID=%d, err=%v", sid, err)
		response.RespondError(c, err)
		return
	}

	log.Printf("[DEBUG] Permission List: systemID=%d, page=%d, pageSize=%d, total=%d, count=%d", sid, page, pageSize, total, len(perms))
	response.OKPage(c, total, perms)
}

// Update modifies the fields of an existing permission.
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("pid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrPermNotFound)
		return
	}

	var req permService.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	perm, err := h.permSvc.Update(id, &req)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, perm)
}

// Delete removes a permission by its ID.
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("pid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrPermNotFound)
		return
	}

	if err := h.permSvc.Delete(id); err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKMessage(c)
}
