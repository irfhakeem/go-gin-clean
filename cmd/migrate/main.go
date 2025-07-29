package main

import (
	"log"
	"os"

	"go-gin-clean/pkg/config"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database schema models for migration
type User struct {
	ID            int64  `gorm:"primaryKey;autoIncrement"`
	Name          string `gorm:"not null"`
	EmailEmail    string `gorm:"column:email_email;uniqueIndex;not null"`
	PasswordValue string `gorm:"column:password_password;not null"`
	Avatar        string `gorm:"default:''"`
	Gender        string `gorm:"type:gender;default:'Unknown';not null"`
	IsActive      bool   `gorm:"default:true;not null"`
	CreatedAt     int64  `gorm:"column:created_at;type:bigint;not null"`
	UpdatedAt     int64  `gorm:"column:updated_at;type:bigint;not null"`
	DeletedAt     *int64 `gorm:"column:deleted_at;type:bigint"`
	IsDeleted     bool   `gorm:"column:is_deleted;type:boolean;default:false"`
}

type RefreshToken struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	UserID    int64  `gorm:"not null"`
	Token     string `gorm:"not null;unique"`
	ExpiryAt  int64  `gorm:"not null;type:bigint"`
	IsRevoked bool   `gorm:"default:false;not null"`
	CreatedAt int64  `gorm:"column:created_at;type:bigint;not null"`
	UpdatedAt int64  `gorm:"column:updated_at;type:bigint;not null"`
	DeletedAt *int64 `gorm:"column:deleted_at;type:bigint"`
	IsDeleted bool   `gorm:"column:is_deleted;type:boolean;default:false"`
}

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

	// Setup database connection
	db, err := setupDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Check command line arguments
	if len(os.Args) < 2 {
		log.Fatal("Usage: migrate [migrate|rollback|fresh]")
	}

	command := os.Args[1]

	switch command {
	case "migrate":
		runMigrations(db)
	case "rollback":
		runRollback(db)
	case "fresh":
		runFreshMigrations(db)
	default:
		log.Fatal("Unknown command. Available commands: migrate, rollback, fresh")
	}
}

func setupDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.DSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *gorm.DB) {
	log.Println("Running database migrations...")

	// Create enums
	enums := map[string][]string{
		"gender": {"Male", "Female", "Unknown"},
	}

	for name, values := range enums {
		quotedValues := make([]string, len(values))
		for i, value := range values {
			quotedValues[i] = "'" + value + "'"
		}
		sql := "CREATE TYPE IF NOT EXISTS " + name + " AS ENUM (" +
			append([]string{}, quotedValues...)[0]
		for _, v := range quotedValues[1:] {
			sql += ", " + v
		}
		sql += ")"

		if err := db.Exec(sql).Error; err != nil {
			log.Printf("Error creating enum %s: %v", name, err)
		}
	}

	// Run auto migrations
	err := db.AutoMigrate(
		&User{},
		&RefreshToken{},
	)

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Database migrations completed successfully")
}

func runRollback(db *gorm.DB) {
	log.Println("Running database rollback...")

	models := []interface{}{
		&User{},
		&RefreshToken{},
	}

	for _, model := range models {
		if err := db.Migrator().DropTable(model); err != nil {
			log.Printf("Error dropping table %T: %v", model, err)
		}
	}

	// Drop enums
	enums := []string{"gender"}
	for _, name := range enums {
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
