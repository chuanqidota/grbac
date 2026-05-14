package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	auditHandler "grbac/internal/handler/audit"
	authHandler "grbac/internal/handler/auth"
	externalHandler "grbac/internal/handler/external"
	menuHandler "grbac/internal/handler/menu"
	permHandler "grbac/internal/handler/permission"
	roleHandler "grbac/internal/handler/role"
	systemHandler "grbac/internal/handler/system"
	userHandler "grbac/internal/handler/user"
	webhookHandler "grbac/internal/handler/webhook"
	"grbac/internal/middleware"
	auditService "grbac/internal/service/audit"
	authService "grbac/internal/service/auth"
	systemService "grbac/internal/service/system"
)

// SetupRouter creates and configures the Gin engine with all routes and
// middleware for the RBAC platform.
func SetupRouter(
	jwtSecret []byte,
	rdb *redis.Client,
	authSvc *authService.Service,
	systemSvc *systemService.Service,
	auditSvc *auditService.Service,
	authH *authHandler.Handler,
	userH *userHandler.Handler,
	systemH *systemHandler.Handler,
	roleH *roleHandler.Handler,
	menuH *menuHandler.Handler,
	permH *permHandler.Handler,
	externalH *externalHandler.Handler,
	webhookH *webhookHandler.Handler,
	auditH *auditHandler.Handler,
) *gin.Engine {
	r := gin.New()

	// Global middleware.
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.AuditMiddleware(auditSvc))

	// Build reusable middleware instances.
	authMW := middleware.AuthMiddleware(jwtSecret, authSvc)
	superAdminMW := middleware.SuperAdminMiddleware()
	systemAdminMW := middleware.SystemAdminMiddleware(systemSvc)
	externalMW := middleware.ExternalAuthMiddleware(systemSvc)
	loginRateLimit := middleware.LoginRateLimit(rdb)
	externalRateLimit := middleware.ExternalRateLimit(rdb)

	api := r.Group("/api")

	// ---- Health check ----
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// ---- Public auth routes (no token required) ----
	publicAuth := api.Group("/auth")
	publicAuth.Use(loginRateLimit)
	{
		publicAuth.POST("/login", authH.Login)
		publicAuth.POST("/refresh", authH.Refresh)
	}

	// ---- Authenticated auth routes ----
	protectedAuth := api.Group("/auth")
	protectedAuth.Use(authMW)
	{
		protectedAuth.POST("/logout", authH.Logout)
		protectedAuth.POST("/change-password", authH.ChangePassword)
	}

	// ---- Authenticated user info route ----
	api.GET("/users/me", authMW, userH.GetMe)

	// ---- Super-admin routes: user management ----
	users := api.Group("/users")
	users.Use(authMW, superAdminMW)
	{
		users.POST("", userH.Create)
		users.GET("", userH.List)
		users.GET("/:id", userH.GetByID)
		users.PUT("/:id", userH.Update)
		users.DELETE("/:id", userH.Delete)
		users.PUT("/:id/status", userH.UpdateStatus)
		users.PUT("/:id/reset-password", userH.ResetPassword)
		users.PUT("/:id/super-admin", userH.UpdateSuperAdmin)
		users.GET("/:id/roles", userH.GetUserRoles)
	}

	// ---- Authenticated routes: system list (filtered by role) ----
	api.GET("/systems", authMW, systemH.List)

	// ---- Super-admin routes: system CRUD ----
	systems := api.Group("/systems")
	systems.Use(authMW, superAdminMW)
	{
		systems.POST("", systemH.Create)
		systems.GET("/:id", systemH.GetByID)
		systems.PUT("/:id", systemH.Update)
		systems.DELETE("/:id", systemH.Delete)
	}

	// ---- System-admin routes: webhooks ----
	webhooks := api.Group("/systems/:id/webhooks")
	webhooks.Use(authMW, systemAdminMW)
	{
		webhooks.POST("", webhookH.Create)
		webhooks.GET("", webhookH.GetBySystemID)
		webhooks.PUT("/:wid", webhookH.Update)
		webhooks.DELETE("/:wid", webhookH.Delete)
	}

	// ---- System-admin routes: member management ----
	members := api.Group("/systems/:id/members")
	members.Use(authMW, systemAdminMW)
	{
		members.POST("", systemH.AddMember)
		members.DELETE("/:uid", systemH.RemoveMember)
		members.GET("", systemH.GetMembers)
		members.GET("/users", systemH.GetMemberUsers)
		members.GET("/:uid/roles", systemH.GetMemberRoles)
		members.GET("/:uid/menus", systemH.GetMemberMenus)
		members.GET("/:uid/permissions", systemH.GetMemberPermissions)
	}

	// ---- System-admin routes: roles ----
	roles := api.Group("/systems/:id/roles")
	roles.Use(authMW, systemAdminMW)
	{
		roles.POST("", roleH.Create)
		roles.GET("", roleH.List)
		roles.PUT("/:rid", roleH.Update)
		roles.DELETE("/:rid", roleH.Delete)
		roles.POST("/:rid/menus", roleH.AssignMenus)
		roles.GET("/:rid/menus", roleH.GetRoleMenus)
		roles.POST("/:rid/permissions", roleH.AssignPermissions)
		roles.GET("/:rid/permissions", roleH.GetRolePermissions)
		roles.POST("/:rid/users", roleH.AssignUsers)
		roles.GET("/:rid/users", roleH.GetRoleUsers)
		roles.DELETE("/:rid/users/:uid", roleH.RemoveUser)
	}

	// ---- System-admin routes: menus ----
	menus := api.Group("/systems/:id/menus")
	menus.Use(authMW, systemAdminMW)
	{
		menus.POST("", menuH.Create)
		menus.GET("", menuH.GetTree)
		menus.PUT("/:mid", menuH.Update)
		menus.DELETE("/:mid", menuH.Delete)
	}

	// ---- System-admin routes: permissions ----
	perms := api.Group("/systems/:id/permissions")
	perms.Use(authMW, systemAdminMW)
	{
		perms.POST("", permH.Create)
		perms.GET("", permH.List)
		perms.PUT("/:pid", permH.Update)
		perms.DELETE("/:pid", permH.Delete)
	}

	// ---- Super-admin routes: audit logs ----
	auditLogs := api.Group("/audit-logs")
	auditLogs.Use(authMW, superAdminMW)
	{
		auditLogs.GET("", auditH.List)
	}

	// ---- External system API (system-credential auth + rate limit) ----
	ext := api.Group("/external")
	ext.Use(externalMW, externalRateLimit)
	{
		ext.GET("/user-roles", externalH.GetUserRoles)
		ext.GET("/menus", externalH.GetMenus)
		ext.GET("/user-apis", externalH.GetUserAPIs)
		ext.GET("/check-permission", externalH.CheckPermission)
	}

	return r
}
