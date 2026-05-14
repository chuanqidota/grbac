package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"grbac/internal/app"
	"grbac/internal/config"
	"grbac/internal/database"
	"grbac/internal/migrate"
	"grbac/internal/model"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Subcommand: ./server migrate
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		db, err := database.NewMySQL(cfg.Database)
		if err != nil {
			log.Fatalf("Failed to connect database: %v", err)
		}
		models := []any{
			&model.User{},
			&model.System{},
			&model.SystemMember{},
			&model.Role{},
			&model.Menu{},
			&model.Permission{},
			&model.UserRole{},
			&model.RoleMenu{},
			&model.RolePermission{},
			&model.Webhook{},
			&model.AuditLog{},
		}
		if err := migrate.Run(db, models); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		return
	}

	go func() {
		if err := app.Run(cfg); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
