package audit

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/response"
	auditService "grbac/internal/service/audit"
)

// Handler handles audit log query requests.
type Handler struct {
	auditSvc *auditService.Service
}

// NewHandler creates a new Handler.
func NewHandler(auditSvc *auditService.Service) *Handler {
	return &Handler{auditSvc: auditSvc}
}

// List returns a paginated list of audit logs.
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var systemID *int64
	if sid := c.Query("system_id"); sid != "" {
		id, err := strconv.ParseInt(sid, 10, 64)
		if err == nil {
			systemID = &id
		}
	}

	logs, total, err := h.auditSvc.List(page, pageSize, systemID)
	if err != nil {
		response.Fail(c, errors.ErrInternal.Wrap(err.Error()))
		return
	}

	response.OKPage(c, total, logs)
}
