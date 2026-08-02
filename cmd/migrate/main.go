package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"go-gin-clean/pkg/config"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	_ = godotenv.Load(".env")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	if len(os.Args) < 2 {
		log.Fatal("Usage: migrate [up|down|force|version|create]")
	}

	command := os.Args[1]

	switch command {
	case "up":
		runMigrateUp(cfg)
	case "down":
		runMigrateDown(cfg)
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("Usage: migrate force <version>")
		}
		version, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("Invalid version: %v", err)
		}
		runMigrateForce(cfg, version)
	case "version":
		runMigrateVersion(cfg)
	case "create":
		if len(os.Args) < 3 {
			log.Fatal("Usage: migrate create <migration_name>")
		}
		createMigration(os.Args[2])
	default:
		log.Fatal("Unknown command. Available commands: up, down, force, version, create")
	}
}

func setupDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.DSN()

	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	return db, nil
}

func getMigrate(cfg *config.Config) (*migrate.Migrate, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	driver, err := migratepg.WithInstance(db, &migratepg.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return m, nil
}

func runMigrateUp(cfg *config.Config) {
	log.Println("Running migrations up...")

	m, err := getMigrate(cfg)
	if err != nil {
		log.Fatalf("Error initializing migrate: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No new migrations to apply")
			return
		}
		log.Fatalf("Migration failed: %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		log.Printf("Could not get version: %v", err)
	} else {
		log.Printf("Migration completed successfully. Current version: %d, Dirty: %v", version, dirty)
	}
}

func runMigrateDown(cfg *config.Config) {
	log.Println("Running migrations down...")

	m, err := getMigrate(cfg)
	if err != nil {
		log.Fatalf("Error initializing migrate: %v", err)
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No migrations to rollback")
			return
		}
		log.Fatalf("Migration rollback failed: %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		log.Printf("Could not get version: %v", err)
	} else {
		log.Printf("Migration rollback completed. Current version: %d, Dirty: %v", version, dirty)
	}
}

func runMigrateForce(cfg *config.Config, version int) {
	log.Printf("Forcing migration to version %d...\n", version)

	m, err := getMigrate(cfg)
	if err != nil {
		log.Fatalf("Error initializing migrate: %v", err)
	}
	defer m.Close()

	if err := m.Force(version); err != nil {
		log.Fatalf("Force migration failed: %v", err)
	}

	log.Printf("Successfully forced migration to version %d", version)
}

func runMigrateVersion(cfg *config.Config) {
	m, err := getMigrate(cfg)
	if err != nil {
		log.Fatalf("Error initializing migrate: %v", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			log.Println("No migrations have been applied yet")
			return
		}
		log.Fatalf("Could not get version: %v", err)
	}

	log.Printf("Current migration version: %d, Dirty: %v", version, dirty)
}

func createMigration(name string) {
	if name == "" {
		log.Fatal("Migration name cannot be empty")
	}

	files, err := os.ReadDir("migrations")
	if err != nil {
		log.Fatalf("Error reading migrations directory: %v", err)
	}

	nextNum := 1
	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()
			if len(filename) >= 6 {
				if num, err := strconv.Atoi(filename[:6]); err == nil {
					if num >= nextNum {
						nextNum = num + 1
					}
				}
			}
		}
	}

	migrationPrefix := fmt.Sprintf("%06d_%s", nextNum, name)
	upFile := fmt.Sprintf("migrations/%s.up.sql", migrationPrefix)
	downFile := fmt.Sprintf("migrations/%s.down.sql", migrationPrefix)

	// Create up migration file
	upContent := fmt.Sprintf("-- Migration: %s\n-- Created at: %s\n\n-- Add your UP migration SQL here\n", name, os.Args[0])
	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		log.Fatalf("Error creating up migration file: %v", err)
	}

	downContent := fmt.Sprintf("-- Migration: %s\n-- Created at: %s\n\n-- Add your DOWN migration SQL here\n", name, os.Args[0])
	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		log.Fatalf("Error creating down migration file: %v", err)
	}

	log.Printf("Created migration files:\n  - %s\n  - %s\n", upFile, downFile)
}
