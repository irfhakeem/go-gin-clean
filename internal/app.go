package internal

import (
	"go-gin-clean/internal/application/port"
	"go-gin-clean/internal/application/usecase"
	"go-gin-clean/internal/delivery/http"
	"go-gin-clean/internal/domain/permission"
	"go-gin-clean/internal/domain/policy"
	"go-gin-clean/internal/infrastructure/cache"
	"go-gin-clean/internal/infrastructure/mailer"
	"go-gin-clean/internal/infrastructure/messaging/rabbitmq"
	"go-gin-clean/internal/infrastructure/repository/postgres"
	"go-gin-clean/internal/infrastructure/security"
	"go-gin-clean/internal/infrastructure/storage"
	"go-gin-clean/internal/worker"
	"go-gin-clean/pkg/config"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type App struct {
	Checker        permission.Checker
	Token          port.TokenMaker
	Publisher      port.Publisher
	UserHandler    *http.UserHandler
	AuthHandler    *http.AuthHandler
	OutboxWorker   *worker.OutboxWorker
	ConsumerWorker *worker.ConsumerWorker
}

func NewApp(db *gorm.DB, rabbitConn *amqp.Connection, redisClient *redis.Client, cfg *config.Config) *App {
	checker := permission.NewChecker()
	userPolicy := policy.NewUserPolicy()

	jwt := security.NewJWTMaker(&cfg.JWT)
	bcrypt := security.NewBcryptHasher()
	aes := security.NewAESEncryptor(&cfg.AES)
	redis := cache.NewRedisCache(redisClient, &cfg.Redis)
	smtp := mailer.NewSMTPMailer(&cfg.Email)
	rabbitMQ := rabbitmq.NewPublisher(rabbitConn, &cfg.RabbitMQ)
	cloudinary := storage.NewCloudinaryStorage(&cfg.Cloudinary)

	txManager := postgres.NewGormTransactionManager(db)
	userRepo := postgres.NewPostgresUserRepository(db)
	refreshTokenRepo := postgres.NewPostgresRefreshTokenRepository(db)
	outboxRepo := postgres.NewPostgresOutboxRepository(db)

	outboxUseCase := usecase.NewOutboxUseCase(outboxRepo, rabbitMQ, &cfg.RabbitMQ)
	authUseCase := usecase.NewAuthUseCase(txManager, userRepo, refreshTokenRepo, outboxUseCase, jwt, bcrypt, aes, &cfg.Server)
	userUseCase := usecase.NewUserUseCase(userPolicy, txManager, userRepo, refreshTokenRepo, outboxUseCase, bcrypt, cloudinary, redis, &cfg.Server)
	emailUseCase := usecase.NewEmailUseCase(smtp, &cfg.Server)

	authHandler := http.NewAuthHandler(authUseCase)
	userHandler := http.NewUserHandler(userUseCase)

	outboxWorker := worker.NewOutboxWorker(outboxUseCase)
	consumerWorker := worker.NewConsumerWorker(rabbitConn, emailUseCase, &cfg.RabbitMQ)

	return &App{
		Checker:        checker,
		Token:          jwt,
		Publisher:      rabbitMQ,
		AuthHandler:    authHandler,
		UserHandler:    userHandler,
		OutboxWorker:   outboxWorker,
		ConsumerWorker: consumerWorker,
	}
}
