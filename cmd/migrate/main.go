package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"go-gin-clean/internal/entity"
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

var (
	models = []any{
		&entity.User{},
		&entity.RefreshToken{},
	}

	enums = map[string][]string{
		"gender":      {"Male", "Female", "Other"},
		"join_status": {"Pending", "Accepted", "Rejected"},
		"role":        {"Admin", "User"},
	}
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Check command line arguments
	if len(os.Args) < 2 {
		log.Fatal("Usage: migrate [up|down|force|version|create|migrate|rollback|fresh]")
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
	case "migrate":
		// Legacy GORM migration
		db, err := setupDatabase(&cfg.Database)
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		runMigrations(db)
	case "rollback":
		db, err := setupDatabase(&cfg.Database)
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		runRollback(db)
	case "fresh":
		db, err := setupDatabase(&cfg.Database)
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		runFreshMigrations(db)
	default:
		log.Fatal("Unknown command. Available commands: up, down, force, version, create, migrate, rollback, fresh")
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

func runMigrations(db *gorm.DB) {
	log.Println("Running database migrations...")

	for name, values := range enums {
		quotedValues := make([]string, len(values))
		for i, value := range values {
			quotedValues[i] = "'" + value + "'"
		}
		sql := "DO $$ BEGIN " +
			"IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = '" + name + "') THEN " +
			"CREATE TYPE " + name + " AS ENUM (" + quotedValues[0]
		for _, v := range quotedValues[1:] {
			sql += ", " + v
		}
		sql += "); END IF; END $$;"

		if err := db.Exec(sql).Error; err != nil {
			log.Printf("Error creating enum %s: %v", name, err)
		}
	}

	// Run auto migrations
	err := db.AutoMigrate(
		models...,
	)

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Database migrations completed successfully")
}

func runRollback(db *gorm.DB) {
	log.Println("Running database rollback...")

	for _, model := range models {
		if err := db.Migrator().DropTable(model); err != nil {
			log.Printf("Error dropping table %T: %v", model, err)
		}
	}

	for name := range enums {
		if err := db.Exec("DROP TYPE IF EXISTS " + name).Error; err != nil {
			log.Printf("Error dropping enum %s: %v", name, err)
		}
	}

	log.Println("Database rollback completed successfully")
}

func runFreshMigrations(db *gorm.DB) {
	log.Println("Running fresh migrations...")
	runRollback(db)
	runMigrations(db)
}

// Golang-migrate functions for production

func getMigrate(cfg *config.Config) (*migrate.Migrate, error) {
	// Create database connection string for lib/pq
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create driver instance
	driver, err := migratepg.WithInstance(db, &migratepg.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	// Create migrate instance
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

	// Get the next migration number
	files, err := os.ReadDir("migrations")
	if err != nil {
		log.Fatalf("Error reading migrations directory: %v", err)
	}

	nextNum := 1
	for _, file := range files {
		if !file.IsDir() {
			// Migration files are named like: 000001_name.up.sql
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

	// Create down migration file
	downContent := fmt.Sprintf("-- Migration: %s\n-- Created at: %s\n\n-- Add your DOWN migration SQL here\n", name, os.Args[0])
	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		log.Fatalf("Error creating down migration file: %v", err)
	}

	log.Printf("Created migration files:\n  - %s\n  - %s\n", upFile, downFile)
}
