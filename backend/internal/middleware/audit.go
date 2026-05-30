package middleware

import (
	"bytes"
	"io"
	"log"
	"strconv"
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

		// Extract resource from URL route pattern.
		resource, resourceID := extractResource(c)

		// Build detail from request body.
		detail := string(bodyBytes)
		if len(detail) > 2000 {
			detail = detail[:2000]
		}

		// Resolve system ID if present.
		var systemID *int64
		if sid := c.Param("id"); sid != "" {
			if id, err := strconv.ParseInt(sid, 10, 64); err == nil {
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
		go func() {
			if err := auditSvc.Record(auditLog); err != nil {
				log.Printf("[AUDIT] failed to record: user=%d action=%s resource=%s err=%v",
					auditLog.UserID, auditLog.Action, auditLog.Resource, err)
			}
		}()

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

func extractResource(c *gin.Context) (string, *int64) {
	pattern := c.FullPath()
	segments := strings.Split(pattern, "/")

	skipPrefixes := map[string]bool{"api": true, "systems": true, "auth": true, "external": true}

	// Find the last named resource segment (non-parameter, non-prefix).
	resource := ""
	for i := len(segments) - 1; i >= 0; i-- {
		seg := segments[i]
		if strings.HasPrefix(seg, ":") || seg == "" {
			continue
		}
		if skipPrefixes[seg] {
			continue
		}
		resource = seg
		break
	}

	// Extract the last numeric ID from URL parameters.
	var resourceID *int64
	for _, param := range c.Params {
		if id, err := strconv.ParseInt(param.Value, 10, 64); err == nil {
			resourceID = &id
		}
	}

	return resource, resourceID
}
