package migrate

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gorm.io/gorm"
)

type migrationRecord struct {
	Filename  string    `gorm:"column:filename;primaryKey"`
	AppliedAt time.Time `gorm:"column:applied_at"`
}

func (migrationRecord) TableName() string { return "schema_migrations" }

// Run synchronises the database schema via GORM AutoMigrate, then executes any
// pending SQL migration files from dir.
//   - models: GORM model structs whose tags define columns, indexes, and constraints.
//   - dir: path to the directory containing numbered .sql migration files.
func Run(db *gorm.DB, dir string, models []any) error {
	// Step 1: AutoMigrate — creates/alters tables and indexes based on model tags.
	log.Println("migrate: running auto-migrate on models...")
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}
	log.Println("migrate: auto-migrate done")

	// Step 2: Ensure tracking table exists.
	if err := db.AutoMigrate(&migrationRecord{}); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// Step 3: Collect .sql files.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	if len(files) == 0 {
		log.Println("migrate: no .sql files found")
		return nil
	}

	// Step 4: Load already-applied filenames.
	var applied []migrationRecord
	if err := db.Find(&applied).Error; err != nil {
		return fmt.Errorf("query applied migrations: %w", err)
	}
	appliedSet := make(map[string]bool, len(applied))
	for _, r := range applied {
		appliedSet[r.Filename] = true
	}

	// Step 5: Execute pending SQL migrations.
	for _, name := range files {
		if appliedSet[name] {
			log.Printf("migrate: %s — already applied, skipping", name)
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(string(content)).Error; err != nil {
				return err
			}
			return tx.Create(&migrationRecord{
				Filename:  name,
				AppliedAt: time.Now(),
			}).Error
		})
		if err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}

		log.Printf("migrate: %s — applied", name)
	}

	log.Println("migrate: all done")
	return nil
}
