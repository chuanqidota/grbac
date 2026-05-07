package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jinang/grbac/internal/config"
	"github.com/jinang/grbac/internal/database"
	"github.com/jinang/grbac/internal/handler"
	"github.com/jinang/grbac/internal/repository"
	"github.com/jinang/grbac/internal/router"
	"github.com/jinang/grbac/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接数据库
	db, err := database.NewMySQL(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	redis, err := database.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}

	// 初始化Repository
	systemRepo := repository.NewSystemRepo(db)
	userRepo := repository.NewUserRepo(db)
	roleRepo := repository.NewRoleRepo(db)
	menuRepo := repository.NewMenuRepo(db)
	permRepo := repository.NewPermissionRepo(db)

	// 初始化Service
	authService := service.NewAuthService(userRepo, redis, cfg.JWT.Secret, cfg.Password)
	userService := service.NewUserService(userRepo)
	systemService := service.NewSystemService(systemRepo, userRepo)
	roleService := service.NewRoleService(roleRepo, userRepo)
	menuService := service.NewMenuService(menuRepo)
	permService := service.NewPermissionService(permRepo)
	externalService := service.NewExternalService(
		userRepo, systemRepo, roleRepo, menuRepo, permRepo,
		cfg.JWT.Secret, authService,
	)

	// 初始化Handler
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	systemHandler := handler.NewSystemHandler(systemService)
	roleHandler := handler.NewRoleHandler(roleService)
	menuHandler := handler.NewMenuHandler(menuService)
	permHandler := handler.NewPermissionHandler(permService)
	externalHandler := handler.NewExternalHandler(externalService)

	// 设置路由
	r := router.SetupRouter(
		[]byte(cfg.JWT.Secret),
		authService, systemService,
		authHandler, userHandler, systemHandler,
		roleHandler, menuHandler, permHandler, externalHandler,
	)

	// 启动服务
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		log.Printf("Starting RBAC server on %s", addr)
		if err := r.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
