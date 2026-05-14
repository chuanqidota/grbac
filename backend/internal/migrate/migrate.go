package migrate

import (
	"fmt"
	"log"

	"grbac/internal/model"
	"grbac/internal/pkg/crypto"

	"gorm.io/gorm"
)

// Run synchronises the database schema via GORM AutoMigrate, then seeds the
// initial admin user if it does not yet exist.
func Run(db *gorm.DB, models []any) error {
	log.Println("migrate: running auto-migrate...")
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}
	log.Println("migrate: auto-migrate done")

	// Seed admin user.
	var count int64
	if err := db.Model(&model.User{}).Where("username = ?", "admin").Count(&count).Error; err != nil {
		return fmt.Errorf("check admin user: %w", err)
	}
	if count > 0 {
		log.Println("migrate: admin user already exists, skipping seed")
		return nil
	}

	hash, err := crypto.HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	admin := &model.User{
		Username:     "admin",
		ChineseName:  "管理员",
		PasswordHash: hash,
		IsSuperAdmin: 1,
		Status:       1,
	}
	if err := db.Create(admin).Error; err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}

	log.Println("migrate: seeded admin user (admin / admin123)")
	return nil
}
