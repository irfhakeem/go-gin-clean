package internal

import (
	"go-gin-clean/internal/delivery/http"
	"go-gin-clean/internal/gateway/cache"
	"go-gin-clean/internal/gateway/mailer"
	"go-gin-clean/internal/gateway/media"
	"go-gin-clean/internal/gateway/messaging"
	"go-gin-clean/internal/gateway/security"
	"go-gin-clean/internal/repository"
	"go-gin-clean/internal/usecase"
	"go-gin-clean/internal/worker"
	"go-gin-clean/pkg/config"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type App struct {
	UserHandler      http.UserHandler
	OauthHandler     http.OAuthHandler
	JWTService       security.JWTServiceInterface
	OAuthService     security.OAuthServiceInterface
	PublisherService messaging.RabbitMQPublisherServiceInterface
	OutboxWorker     *worker.OutboxWorker
	ConsumerWorker   *worker.ConsumerWorker
}

func NewApp(db *gorm.DB, rabbitConn *amqp.Connection, redisClient *redis.Client, cfg *config.Config) *App {
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	outboxRepo := repository.NewOutboxRepository(db)

	jwtService := security.NewJWTService(&cfg.JWT)
	bcryptService := security.NewBcryptService()
	oauthService := security.NewOAuthService(&cfg.OAuth)
	aesService := security.NewAESService(&cfg.AES)
	redisService := cache.NewRedisService(redisClient, &cfg.Redis)
	smtpService := mailer.NewSMTPService(&cfg.Email)
	publisherService := messaging.NewRabbitMQPublisherService(rabbitConn, &cfg.RabbitMQ)
	storageService := media.NewCloudinaryService(&cfg.Cloudinary)

	outboxUseCase := usecase.NewOutboxUseCase(outboxRepo, publisherService, &cfg.RabbitMQ)
	userUseCase := usecase.NewUserUseCase(userRepo, refreshTokenRepo, outboxUseCase, jwtService, bcryptService, oauthService, aesService, storageService, redisService, &cfg.Server)
	emailUseCase := usecase.NewEmailUseCase(smtpService, &cfg.Server)

	userHandler := http.NewUserHandler(userUseCase)
	oauthHandler := http.NewOAuthHandler(userUseCase)

	outboxWorker := worker.NewOutboxWorker(outboxUseCase)
	consumerWorker := worker.NewConsumerWorker(rabbitConn, emailUseCase, &cfg.RabbitMQ)

	return &App{
		UserHandler:      *userHandler,
		OauthHandler:     *oauthHandler,
		JWTService:       jwtService,
		OAuthService:     oauthService,
		PublisherService: publisherService,
		OutboxWorker:     outboxWorker,
		ConsumerWorker:   consumerWorker,
	}
}
