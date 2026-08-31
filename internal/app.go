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
	"go-gin-clean/internal/infrastructure/oauth"
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
	UserPolicy     policy.UserPolicy
	Token          port.TokenMaker
	Publisher      port.Publisher
	UserHandler    *http.UserHandler
	OauthHandler   *http.OAuthHandler
	OutboxWorker   *worker.OutboxWorker
	ConsumerWorker *worker.ConsumerWorker
}

func NewApp(db *gorm.DB, rabbitConn *amqp.Connection, redisClient *redis.Client, cfg *config.Config) *App {
	checker := permission.NewChecker()
	userPolicy := policy.NewUserPolicy()

	jwt := security.NewJWTMaker(&cfg.JWT)
	bcrypt := security.NewBcryptHasher()
	aes := security.NewAESEncryptor(&cfg.AES)
	oauthClient := oauth.NewGoogleOAuth(&cfg.OAuth)
	redis := cache.NewRedisCache(redisClient, &cfg.Redis)
	smtp := mailer.NewSMTPMailer(&cfg.Email)
	rabbitMQ := rabbitmq.NewPublisher(rabbitConn, &cfg.RabbitMQ)
	cloudinary := storage.NewCloudinaryStorage(&cfg.Cloudinary)

	userRepo := postgres.NewPostgresUserRepository(db)
	refreshTokenRepo := postgres.NewPostgresRefreshTokenRepository(db)
	outboxRepo := postgres.NewPostgresOutboxRepository(db)

	outboxUseCase := usecase.NewOutboxUseCase(outboxRepo, rabbitMQ, &cfg.RabbitMQ)
	userUseCase := usecase.NewUserUseCase(userRepo, refreshTokenRepo, outboxUseCase, jwt, bcrypt, oauthClient, aes, cloudinary, redis, &cfg.Server)
	emailUseCase := usecase.NewEmailUseCase(smtp, &cfg.Server)

	userHandler := http.NewUserHandler(userUseCase)
	oauthHandler := http.NewOAuthHandler(userUseCase)

	outboxWorker := worker.NewOutboxWorker(outboxUseCase)
	consumerWorker := worker.NewConsumerWorker(rabbitConn, emailUseCase, &cfg.RabbitMQ)

	return &App{
		Checker:        checker,
		UserPolicy:     userPolicy,
		Token:          jwt,
		Publisher:      rabbitMQ,
		UserHandler:    userHandler,
		OauthHandler:   oauthHandler,
		OutboxWorker:   outboxWorker,
		ConsumerWorker: consumerWorker,
	}
}
