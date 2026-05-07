package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/response"
	"github.com/jinang/grbac/internal/service"
)

// ---------- MenuHandler ----------

// MenuHandler handles menu management HTTP requests.
type MenuHandler struct {
	menuService *service.MenuService
}

// NewMenuHandler creates a new MenuHandler.
func NewMenuHandler(menuService *service.MenuService) *MenuHandler {
	return &MenuHandler{menuService: menuService}
}

// Create registers a new menu within a system.
func (h *MenuHandler) Create(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	var req service.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	menu, err := h.menuService.Create(sid, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, menu)
}

// GetTree returns the full menu tree for a system.
func (h *MenuHandler) GetTree(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	tree, err := h.menuService.GetTree(sid)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, tree)
}

// Update modifies an existing menu.
func (h *MenuHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrMenuNotFound)
		return
	}

	var req service.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal)
		return
	}

	menu, err := h.menuService.Update(id, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, menu)
}

// Delete removes a menu by its ID.
func (h *MenuHandler) Delete(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrSystemNotFound)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrMenuNotFound)
		return
	}

	if err := h.menuService.Delete(sid, id); err != nil {
		respondError(c, err)
		return
	}

	response.OKMessage(c)
}
