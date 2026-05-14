package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"grbac/internal/pkg/errors"
	"grbac/internal/pkg/response"
	systemService "grbac/internal/service/system"
)

// SystemAdminMiddleware returns a middleware that verifies the authenticated
// user is an administrator of the system identified by the "sid" path parameter.
// Super-admins (is_super_admin=true) bypass the system membership check.
// Must be placed after AuthMiddleware so that CtxUserID and CtxIsSuperAdmin are already set.
func SystemAdminMiddleware(systemService *systemService.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Super-admins pass through without system membership check.
		if isSuper, exists := c.Get(CtxIsSuperAdmin); exists && isSuper.(bool) {
			c.Next()
			return
		}

		// 2. Parse system ID from path parameter.
		sid, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Fail(c, errors.ErrSystemNotFound)
			c.Abort()
			return
		}

		// 3. Get the authenticated user ID from context.
		userID, exists := c.Get(CtxUserID)
		if !exists {
			response.Unauthorized(c, errors.ErrTokenInvalid)
			c.Abort()
			return
		}

		// 4. Check admin membership.
		isAdmin, err := systemService.IsAdmin(sid, userID.(int64))
		if err != nil || !isAdmin {
			response.Forbidden(c, errors.ErrNoPermission)
			c.Abort()
			return
		}

		c.Next()
	}
}
