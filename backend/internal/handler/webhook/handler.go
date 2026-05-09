package webhook

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/response"
	webhookService "grbac/internal/service/webhook"
)

// Handler handles webhook management requests.
type Handler struct {
	webhookSvc *webhookService.Service
}

// NewHandler creates a new Handler.
func NewHandler(webhookSvc *webhookService.Service) *Handler {
	return &Handler{webhookSvc: webhookSvc}
}

// Create registers a new webhook for a system.
func (h *Handler) Create(c *gin.Context) {
	systemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrInternal.Wrap("invalid system id"))
		return
	}

	var req webhookService.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal.Wrap(err.Error()))
		return
	}

	if !webhookService.ValidateEvents(req.Events) {
		response.Fail(c, errors.ErrInternal.Wrap("无效的事件类型"))
		return
	}

	webhook, err := h.webhookSvc.Create(systemID, &req)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, webhook)
}

// GetBySystemID returns all webhooks for a system.
func (h *Handler) GetBySystemID(c *gin.Context) {
	systemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrInternal.Wrap("invalid system id"))
		return
	}

	webhooks, err := h.webhookSvc.GetBySystemID(systemID)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, webhooks)
}

// Update modifies a webhook.
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("wid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrInternal.Wrap("invalid webhook id"))
		return
	}

	var req webhookService.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrInternal.Wrap(err.Error()))
		return
	}

	if req.Events != "" && !webhookService.ValidateEvents(req.Events) {
		response.Fail(c, errors.ErrInternal.Wrap("无效的事件类型"))
		return
	}

	webhook, err := h.webhookSvc.Update(id, &req)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, webhook)
}

// Delete removes a webhook.
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("wid"), 10, 64)
	if err != nil {
		response.Fail(c, errors.ErrInternal.Wrap("invalid webhook id"))
		return
	}

	if err := h.webhookSvc.Delete(id); err != nil {
		response.RespondError(c, err)
		return
	}

	response.OK(c, nil)
}
