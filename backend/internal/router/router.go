package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinang/grbac/internal/handler"
	"github.com/jinang/grbac/internal/middleware"
	"github.com/jinang/grbac/internal/service"
)

// SetupRouter creates and configures the Gin engine with all routes and
// middleware for the RBAC platform.
//
// Route hierarchy:
//
//	/api/health                              - public
//	/api/auth/login, refresh                 - public
//	/api/auth/* (other)                      - authenticated
//	/api/users/*, /api/systems/* (CRUD)      - authenticated + super-admin
//	/api/systems/:sid/roles/* etc.           - authenticated + system-admin
//	/api/external/*                          - system-credential auth
func SetupRouter(
	jwtSecret []byte,
	authService *service.AuthService,
	systemService *service.SystemService,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	systemHandler *handler.SystemHandler,
	roleHandler *handler.RoleHandler,
	menuHandler *handler.MenuHandler,
	permHandler *handler.PermissionHandler,
	externalHandler *handler.ExternalHandler,
) *gin.Engine {
	r := gin.New()

	// Global middleware.
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// Build reusable middleware instances.
	authMW := middleware.AuthMiddleware(jwtSecret, authService)
	superAdminMW := middleware.SuperAdminMiddleware()
	systemAdminMW := middleware.SystemAdminMiddleware(systemService)
	externalMW := middleware.ExternalAuthMiddleware(systemService)

	api := r.Group("/api")

	// ---- Health check ----
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// ---- Public auth routes (no token required) ----
	publicAuth := api.Group("/auth")
	{
		publicAuth.POST("/login", authHandler.Login)
		publicAuth.POST("/refresh", authHandler.Refresh)
	}

	// ---- Authenticated auth routes ----
	protectedAuth := api.Group("/auth")
	protectedAuth.Use(authMW)
	{
		protectedAuth.POST("/logout", authHandler.Logout)
		protectedAuth.POST("/change-password", authHandler.ChangePassword)
	}

	// ---- Super-admin routes: user management ----
	users := api.Group("/users")
	users.Use(authMW, superAdminMW)
	{
		users.POST("", userHandler.Create)
		users.GET("", userHandler.List)
		users.GET("/:id", userHandler.GetByID)
		users.PUT("/:id", userHandler.Update)
		users.DELETE("/:id", userHandler.Delete)
		users.PATCH("/:id/status", userHandler.UpdateStatus)
	}

	// ---- Super-admin routes: system management ----
	systems := api.Group("/systems")
	systems.Use(authMW, superAdminMW)
	{
		systems.POST("", systemHandler.Create)
		systems.GET("", systemHandler.List)
		systems.GET("/:id", systemHandler.GetByID)
		systems.PUT("/:id", systemHandler.Update)
		systems.DELETE("/:id", systemHandler.Delete)
		systems.POST("/:id/members", systemHandler.AddMember)
		systems.DELETE("/:id/members/:uid", systemHandler.RemoveMember)
		systems.GET("/:id/members", systemHandler.GetMembers)
	}

	// ---- System-admin routes: roles ----
	roles := api.Group("/systems/:sid/roles")
	roles.Use(authMW, systemAdminMW)
	{
		roles.POST("", roleHandler.Create)
		roles.GET("", roleHandler.List)
		roles.PUT("/:id", roleHandler.Update)
		roles.DELETE("/:id", roleHandler.Delete)
		roles.POST("/:id/menus", roleHandler.AssignMenus)
		roles.GET("/:id/menus", roleHandler.GetRoleMenus)
		roles.POST("/:id/permissions", roleHandler.AssignPermissions)
		roles.GET("/:id/permissions", roleHandler.GetRolePermissions)
		roles.POST("/:id/users", roleHandler.AssignUsers)
		roles.DELETE("/:id/users/:uid", roleHandler.RemoveUser)
	}

	// ---- System-admin routes: menus ----
	menus := api.Group("/systems/:sid/menus")
	menus.Use(authMW, systemAdminMW)
	{
		menus.POST("", menuHandler.Create)
		menus.GET("", menuHandler.GetTree)
		menus.PUT("/:id", menuHandler.Update)
		menus.DELETE("/:id", menuHandler.Delete)
	}

	// ---- System-admin routes: permissions ----
	perms := api.Group("/systems/:sid/permissions")
	perms.Use(authMW, systemAdminMW)
	{
		perms.POST("", permHandler.Create)
		perms.GET("", permHandler.List)
		perms.PUT("/:id", permHandler.Update)
		perms.DELETE("/:id", permHandler.Delete)
	}

	// ---- External system API (system-credential auth) ----
	ext := api.Group("/external")
	ext.Use(externalMW)
	{
		ext.POST("/verify", externalHandler.Verify)
		ext.POST("/user-info", externalHandler.GetUserInfo)
		ext.POST("/menus", externalHandler.GetMenus)
		ext.POST("/permissions", externalHandler.GetPermissions)
		ext.POST("/validate-permission", externalHandler.ValidatePermission)
	}

	return r
}
