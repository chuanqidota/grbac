package app

import (
	"fmt"
	"log"

	"grbac/internal/config"
	"grbac/internal/database"
	auditH "grbac/internal/handler/audit"
	authH "grbac/internal/handler/auth"
	extH "grbac/internal/handler/external"
	menuH "grbac/internal/handler/menu"
	permH "grbac/internal/handler/permission"
	roleH "grbac/internal/handler/role"
	systemH "grbac/internal/handler/system"
	userH "grbac/internal/handler/user"
	webhookH "grbac/internal/handler/webhook"
	auditR "grbac/internal/repository/audit"
	menuR "grbac/internal/repository/menu"
	permR "grbac/internal/repository/permission"
	roleR "grbac/internal/repository/role"
	systemR "grbac/internal/repository/system"
	userR "grbac/internal/repository/user"
	webhookR "grbac/internal/repository/webhook"
	"grbac/internal/router"
	auditSvc "grbac/internal/service/audit"
	authSvc "grbac/internal/service/auth"
	extSvc "grbac/internal/service/external"
	menuSvc "grbac/internal/service/menu"
	permSvc "grbac/internal/service/permission"
	roleSvc "grbac/internal/service/role"
	systemSvc "grbac/internal/service/system"
	userSvc "grbac/internal/service/user"
	webhookSvc "grbac/internal/service/webhook"
)

// Run initializes all dependencies and starts the HTTP server.
func Run(cfg *config.Config) error {
	// Database connections
	db, err := database.NewMySQL(cfg.Database)
	if err != nil {
		return fmt.Errorf("connect mysql: %w", err)
	}

	redis, err := database.NewRedis(cfg.Redis)
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}

	// Repositories
	userRepo := userR.NewRepo(db)
	sysRepo := systemR.NewRepo(db)
	roleRepo := roleR.NewRepo(db)
	menuRepo := menuR.NewRepo(db)
	permRepo := permR.NewRepo(db)
	webhookRepo := webhookR.NewRepo(db)
	auditRepo := auditR.NewRepo(db)

	// Services
	authService := authSvc.NewService(userRepo, redis, cfg.JWT.Secret, cfg.Password)
	userService := userSvc.NewService(userRepo, roleRepo, sysRepo)
	systemService := systemSvc.NewService(sysRepo, userRepo, roleRepo, menuRepo, permRepo, cfg.EncryptionKey)
	roleService := roleSvc.NewService(db, roleRepo, userRepo)
	menuService := menuSvc.NewService(menuRepo)
	permService := permSvc.NewService(permRepo)
	webhookService := webhookSvc.NewService(webhookRepo)
	auditService := auditSvc.NewService(auditRepo)
	extService := extSvc.NewService(
		userRepo, sysRepo, roleRepo, menuRepo, permRepo,
		cfg.JWT.Secret, authService, redis,
	)

	// Handlers
	authHandler := authH.NewHandler(authService)
	userHandler := userH.NewHandler(userService)
	systemHandler := systemH.NewHandler(systemService)
	roleHandler := roleH.NewHandler(roleService)
	menuHandler := menuH.NewHandler(menuService)
	permHandler := permH.NewHandler(permService)
	extHandler := extH.NewHandler(extService)
	webhookHandler := webhookH.NewHandler(webhookService)
	auditHandler := auditH.NewHandler(auditService)

	// Router
	r := router.SetupRouter(
		[]byte(cfg.JWT.Secret),
		redis,
		authService, systemService, auditService,
		authHandler, userHandler, systemHandler,
		roleHandler, menuHandler, permHandler, extHandler, webhookHandler, auditHandler,
	)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting RBAC server on %s", addr)
	return r.Run(addr)
}
