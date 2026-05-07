package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/response"
	"github.com/jinang/grbac/internal/service"
)

// SystemAdminMiddleware returns a middleware that verifies the authenticated
// user is an administrator of the system identified by the "sid" path parameter.
// Must be placed after AuthMiddleware so that CtxUserID is already set.
func SystemAdminMiddleware(systemService *service.SystemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Parse system ID from path parameter.
		sid, err := strconv.ParseInt(c.Param("sid"), 10, 64)
		if err != nil {
			response.Fail(c, errors.ErrSystemNotFound)
			c.Abort()
			return
		}

		// 2. Get the authenticated user ID from context.
		userID, exists := c.Get(CtxUserID)
		if !exists {
			response.Unauthorized(c, errors.ErrTokenInvalid)
			c.Abort()
			return
		}

		// 3. Check admin membership.
		isAdmin, err := systemService.IsAdmin(sid, userID.(int64))
		if err != nil || !isAdmin {
			response.Forbidden(c, errors.ErrNoPermission)
			c.Abort()
			return
		}

		c.Next()
	}
}
