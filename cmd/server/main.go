package main

import (
	"context"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go-gin-clean/internal/delivery/http/route"
	"go-gin-clean/internal/infrastructure"
	"go-gin-clean/internal/model/validator"
	"go-gin-clean/pkg/config"
	"go-gin-clean/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	_ = godotenv.Load(".env")

	cfg, err := config.Load()
	if err != nil {
		panic("failed to load configuration: " + err.Error())
	}

	logger.Init(cfg.Server.Environment)
	defer logger.Sync()

	db, err := setupDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conn, err := setupRabbitMQ(&cfg.RabbitMQ)
	if err != nil {
		logger.Fatal("failed to connect to RabbitMQ", zap.Error(err))
	}

	container := infrastructure.NewContainer(db, conn, cfg)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		container.OutboxWorker.Run(rootCtx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		container.ConsumerWorker.Run(rootCtx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		watchRabbitMQConnection(rootCtx, conn, &cfg.RabbitMQ, container)
	}()

	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	if err := validator.RegisterCustomValidations(); err != nil {
		logger.Fatal("failed to register custom validations", zap.Error(err))
	}

	route.SetupRoutes(router, &container.UserHandler, &container.OauthHandler, container.JWTService)

	srv := &http.Server{
		Addr:    cfg.Server.Address(),
		Handler: router,
	}

	go func() {
		logger.Info("server starting", zap.String("address", cfg.Server.Address()))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server listen failed", zap.Error(err))
		}
	}()

	<-rootCtx.Done()
	logger.Info("shutdown signal received, starting graceful shutdown")

	shutdownCtx, cancelServer := context.WithTimeout(context.Background(), time.Duration(cfg.Server.Timeout)*time.Second)
	defer cancelServer()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Warn("server forced to shutdown", zap.Error(err))
	}

	logger.Info("waiting for background tasks to finish...")
	wg.Wait()

	if db != nil {
		if dbSQL, err := db.DB(); err == nil {
			dbSQL.Close()
		}
	}

	if conn != nil {
		conn.Close()
	}

	logger.Info("server exited")
}

func setupDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var logLevel gormlogger.LogLevel
	if cfg.Host == "localhost" || cfg.Host == "127.0.0.1" {
		logLevel = gormlogger.Info
	} else {
		logLevel = gormlogger.Error
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
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

func setupRabbitMQ(cfg *config.RabbitMQConfig) (*amqp.Connection, error) {
	return amqp.Dial(cfg.DSN())
}

func watchRabbitMQConnection(ctx context.Context, conn *amqp.Connection, cfg *config.RabbitMQConfig, container *infrastructure.Container) {
	const (
		initDelay = 2 * time.Second
		maxDelay  = 1 * time.Minute
	)

	connClose := conn.NotifyClose(make(chan *amqp.Error, 1))

	for {
		select {
		case <-ctx.Done():
			return
		case amqpErr, ok := <-connClose:
			if !ok {
				return
			}
			logger.Warn("rabbitmq: connection lost, reconnecting", zap.Error(amqpErr))
		}

		delay := initDelay
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}

			newConn, err := setupRabbitMQ(cfg)
			if err != nil {
				logger.Warn("rabbitmq: reconnect failed, retrying", zap.Duration("delay", delay), zap.Error(err))
				delay *= 2
				if delay > maxDelay {
					delay = maxDelay
				}
				continue
			}

			logger.Info("rabbitmq: connection restored, reinitializing workers")
			if err := container.PublisherService.UpdateConnection(newConn); err != nil {
				logger.Error("rabbitmq: failed to update publisher connection", zap.Error(err))
			}
			container.ConsumerWorker.Reconnect(newConn)

			connClose = newConn.NotifyClose(make(chan *amqp.Error, 1))
			break
		}
	}
}
