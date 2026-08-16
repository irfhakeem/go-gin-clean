package connection

import (
	"go-gin-clean/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func ConnectPostgres(env string, cfg *config.DatabaseConfig) (*gorm.DB, error) {
	logLevel := gormlogger.Error
	if env != "production" {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.Url), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.MaxIdleTime)

	return db, nil
}

func ClosePostgres(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
