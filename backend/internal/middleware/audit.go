package middleware

import (
	"bytes"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"grbac/internal/model"
	auditService "grbac/internal/service/audit"
)

// AuditMiddleware records write operations (POST, PUT, DELETE, PATCH) to the audit log.
func AuditMiddleware(auditSvc *auditService.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		// Read and restore the request body for downstream handlers.
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Extract user info from context (set by auth middleware).
		userID, _ := c.Get(CtxUserID)
		username, _ := c.Get(CtxUsername)

		var uid int64
		var uname string
		if v, ok := userID.(int64); ok {
			uid = v
		}
		if v, ok := username.(string); ok {
			uname = v
		}

		// Determine action from HTTP method.
		action := methodToAction(method)

		// Extract resource from URL path.
		resource, resourceID := extractResource(c.FullPath())

		// Build detail from request body.
		detail := string(bodyBytes)
		if len(detail) > 2000 {
			detail = detail[:2000]
		}

		// Resolve system ID if present.
		var systemID *int64
		if sid := c.Param("id"); sid != "" {
			if id, err := parseID(sid); err == nil {
				systemID = &id
			}
		}

		auditLog := &model.AuditLog{
			UserID:     uid,
			Username:   uname,
			SystemID:   systemID,
			Action:     action,
			Resource:   resource,
			ResourceID: resourceID,
			Detail:     detail,
			IP:         c.ClientIP(),
		}

		// Record asynchronously to avoid blocking the response.
		go auditSvc.Record(auditLog)

		c.Next()
	}
}

func methodToAction(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return method
	}
}

func extractResource(path string) (string, *int64) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// Typical patterns:
	// /api/users/:id -> resource=users
	// /api/systems/:id/roles/:rid -> resource=roles
	// /api/systems/:id/menus/:mid -> resource=menus

	resource := ""
	var resourceID *int64

	for i, part := range parts {
		if part == "api" || part == "systems" || part == "auth" || part == "external" {
			continue
		}
		if strings.HasPrefix(part, ":") {
			continue
		}
		// This looks like a resource name.
		if resource == "" {
			resource = part
		}
		// Check if next part is an ID.
		if i+1 < len(parts) && strings.HasPrefix(parts[i+1], ":") {
			if id, err := parseID(strings.TrimPrefix(parts[i+1], ":")); err == nil {
				resourceID = &id
			}
		}
	}

	return resource, resourceID
}

func parseID(s string) (int64, error) {
	var id int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		id = id*10 + int64(c-'0')
	}
	return id, nil
}
