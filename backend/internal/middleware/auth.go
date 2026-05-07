package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/pkg/errors"
	"github.com/jinang/grbac/internal/pkg/jwt"
	"github.com/jinang/grbac/internal/pkg/response"
	"github.com/jinang/grbac/internal/service"
)

// Context keys set by the auth middleware.
// Keep in sync with handler.CtxUserID / handler.CtxAccessToken.
const (
	CtxUserID       = "user_id"
	CtxUsername     = "username"
	CtxIsSuperAdmin = "is_super_admin"
	CtxAccessToken  = "access_token"
)

// AuthMiddleware returns a Gin middleware that validates the JWT Bearer token
// in the Authorization header. On success it populates the request context
// with user_id, username, is_super_admin, and access_token.
func AuthMiddleware(jwtSecret []byte, authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract the Bearer token.
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, errors.ErrTokenInvalid)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, errors.ErrTokenInvalid)
			c.Abort()
			return
		}
		accessToken := parts[1]

		// 2. Check the token blacklist (logout / forced invalidation).
		if authService.IsTokenBlacklisted(c.Request.Context(), accessToken) {
			response.Unauthorized(c, errors.ErrTokenExpired)
			c.Abort()
			return
		}

		// 3. Parse and validate the JWT.
		claims, err := jwt.ParseToken(accessToken, jwtSecret)
		if err != nil {
			response.Unauthorized(c, errors.ErrTokenInvalid)
			c.Abort()
			return
		}

		// 4. Populate context for downstream handlers.
		c.Set(CtxUserID, int64(claims.UserID))
		c.Set(CtxUsername, claims.Username)
		c.Set(CtxIsSuperAdmin, claims.IsSuperAdmin)
		c.Set(CtxAccessToken, accessToken)

		c.Next()
	}
}

// SuperAdminMiddleware returns a middleware that rejects requests when the
// authenticated user is NOT a super administrator. Must be placed after
// AuthMiddleware so that CtxIsSuperAdmin is already set.
func SuperAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isSuper, exists := c.Get(CtxIsSuperAdmin)
		if !exists || !isSuper.(bool) {
			response.Forbidden(c, errors.ErrNoPermission)
			c.Abort()
			return
		}
		c.Next()
	}
}
