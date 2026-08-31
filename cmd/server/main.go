package main

import (
	"context"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go-gin-clean/internal"
	"go-gin-clean/internal/delivery/http/route"
	"go-gin-clean/internal/dto/validator"
	"go-gin-clean/pkg/config"
	"go-gin-clean/pkg/connection"
	"go-gin-clean/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load(".env")

	cfg, err := config.Load()
	if err != nil {
		panic("failed to load configuration: " + err.Error())
	}

	logger.Init(cfg.Server.Environment)
	defer logger.Sync()

	postgresConn, err := connection.ConnectPostgres(cfg.Server.Environment, &cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	redisClient, err := connection.ConnectRedis(&cfg.Redis)
	if err != nil {
		logger.Fatal("failed to connect to redis", zap.Error(err))
	}

	rabbitConn, err := connection.ConnectRabbitMQ(&cfg.RabbitMQ)
	if err != nil {
		logger.Fatal("failed to connect to RabbitMQ", zap.Error(err))
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	container := internal.NewApp(postgresConn, rabbitConn, redisClient, cfg)

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
		connection.WatchRabbitMQ(rootCtx, rabbitConn, &cfg.RabbitMQ, func(newConn *amqp.Connection) {
			rabbitConn = newConn
			if err := container.Publisher.UpdateConnection(newConn); err != nil {
				logger.Error("rabbitmq: failed to update publisher connection", zap.Error(err))
			}
			container.ConsumerWorker.Reconnect(newConn)
		})
	}()

	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	if err := validator.RegisterCustomValidations(); err != nil {
		logger.Fatal("failed to register custom validations", zap.Error(err))
	}

	route.SetupRoutes(router, container.Token, container.Checker, container.UserHandler, container.OauthHandler)

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

	if err := connection.ClosePostgres(postgresConn); err != nil {
		logger.Warn("failed to close database connection", zap.Error(err))
	}

	if err := connection.CloseRedis(redisClient); err != nil {
		logger.Warn("failed to close redis connection", zap.Error(err))
	}

	if err := connection.CloseRabbitMQ(rabbitConn); err != nil {
		logger.Warn("failed to close rabbitmq connection", zap.Error(err))
	}

	logger.Info("server exited")
}
