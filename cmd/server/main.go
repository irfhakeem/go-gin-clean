package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go-gin-clean/internal/delivery/http/route"
	"go-gin-clean/internal/infrastructure"
	"go-gin-clean/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	db, err := setupDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conn, ch, err := setupRabbitMQ(&cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("Error connecting to RabbitMQ: %v", err)
	}
	defer func() {
		if ch != nil {
			ch.Close()
		}
		if conn != nil {
			conn.Close()
		}
	}()

	container := infrastructure.NewContainer(db, ch, cfg)

	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	route.SetupRoutes(router, &container.UserHandler, &container.OauthHandler, &container.JWTService)

	srv := &http.Server{
		Addr:    cfg.Server.Address(),
		Handler: router,
	}

	go func() {
		log.Printf("Starting server on %s...", cfg.Server.Address())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	<-rootCtx.Done()
	log.Println("Received shutdown signal. Starting graceful shutdown...")

	shutdownCtx, cancelServer := context.WithTimeout(context.Background(), time.Duration(cfg.Server.Timeout)*time.Second)
	defer cancelServer()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown (timeout): %v", err)
	}

	log.Println("Server exiting")
}

func setupDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.DSN()

	var logLevel logger.LogLevel
	if cfg.Host == "localhost" || cfg.Host == "127.0.0.1" {
		logLevel = logger.Info
	} else {
		logLevel = logger.Error
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})

	if err != nil {
		return nil, err
	}

	psqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	psqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	psqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	return db, nil
}

func setupRabbitMQ(cfg *config.RabbitMQConfig) (*amqp091.Connection, *amqp091.Channel, error) {
	dsn := cfg.DSN()
	conn, err := amqp091.Dial(dsn)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	return conn, ch, nil
}
