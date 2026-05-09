package middleware

import (
	"github.com/gin-gonic/gin"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/response"
	systemService "grbac/internal/service/system"
)

// Context keys for external (system-credential) authentication.
const (
	CtxSystemID   = "system_id"
	CtxSystemCode = "system_code"
)

// ExternalAuthMiddleware returns a middleware that validates the system
// credentials carried in the X-System-Code / X-System-Secret request headers.
// On success the system_id and system_code are stored in the request context.
func ExternalAuthMiddleware(systemSvc *systemService.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.GetHeader("X-System-Code")
		secret := c.GetHeader("X-System-Secret")
		if code == "" || secret == "" {
			response.Unauthorized(c, errors.ErrSystemCredential)
			c.Abort()
			return
		}

		system, err := systemSvc.ValidateSecret(code, secret)
		if err != nil {
			response.Unauthorized(c, errors.ErrSystemCredential)
			c.Abort()
			return
		}

		if system.Status != 1 {
			response.Unauthorized(c, errors.ErrSystemCredential)
			c.Abort()
			return
		}

		c.Set(CtxSystemID, system.ID)
		c.Set(CtxSystemCode, system.Code)

		c.Next()
	}
}
