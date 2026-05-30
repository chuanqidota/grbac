package menu

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/response"
	menuService "grbac/internal/service/menu"
)

// Handler handles menu management HTTP requests.
type Handler struct {
	menuSvc *menuService.Service
}

// NewHandler creates a new Handler.
func NewHandler(menuSvc *menuService.Service) *Handler {
	return &Handler{menuSvc: menuSvc}
}

// Create registers a new menu within a system.
func (h *Handler) Create(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	var req menuService.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	menu, err := h.menuSvc.Create(c.Request.Context(), sid, &req)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, menu)
}

// GetTree returns the full menu tree for a system.
func (h *Handler) GetTree(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	tree, err := h.menuSvc.GetTree(sid)
	if err != nil {
		log.Printf("[DEBUG] GetTree error: systemID=%d, err=%v", sid, err)
		response.RespondError(c, err)
		return
	}

	log.Printf("[DEBUG] GetTree: systemID=%d, count=%d", sid, len(tree))
	response.OKPage(c, int64(len(tree)), tree)
}

// Update modifies an existing menu.
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("mid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrMenuNotFound)
		return
	}

	var req menuService.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	menu, err := h.menuSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, menu)
}

// Delete removes a menu by its ID.
func (h *Handler) Delete(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	id, err := strconv.ParseInt(c.Param("mid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrMenuNotFound)
		return
	}

	if err := h.menuSvc.Delete(c.Request.Context(), sid, id); err != nil {
		response.RespondError(c, err)
		return
	}

	response.OKMessage(c)
}
