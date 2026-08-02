package infrastructure

import (
	"go-gin-clean/internal/delivery/http"
	"go-gin-clean/internal/gateway/cache"
	"go-gin-clean/internal/gateway/mailer"
	"go-gin-clean/internal/gateway/media"
	"go-gin-clean/internal/gateway/messaging"
	"go-gin-clean/internal/gateway/security"
	"go-gin-clean/internal/infrastructure/worker"
	"go-gin-clean/internal/repository"
	"go-gin-clean/internal/usecase"
	"go-gin-clean/pkg/config"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type Container struct {
	UserHandler      http.UserHandler
	OauthHandler     http.OAuthHandler
	JWTService       security.JWTServiceInterface
	OAuthService     security.OAuthServiceInterface
	PublisherService messaging.PublisherServiceInterface
	OutboxWorker     *worker.OutboxWorker
	ConsumerWorker   *worker.ConsumerWorker
}

func NewContainer(db *gorm.DB, conn *amqp.Connection, cfg *config.Config) *Container {
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	outboxRepo := repository.NewOutboxRepository(db)

	jwtService := security.NewJWTService(&cfg.JWT)
	bcryptService := security.NewBcryptService()
	oauthService := security.NewOAuthService(&cfg.OAuth)
	aesService := security.NewAESService(&cfg.AES)
	redisService := cache.NewRedisService(&cfg.Redis)
	smtpService := mailer.NewSMTPService(&cfg.Email)
	publisherService := messaging.NewPublisherService(conn, cfg.RabbitMQ.Exchange)

	var storageService media.StorageServiceInterface
	if cfg.Cloudinary.CloudinaryURL != "" && cfg.Cloudinary.CloudinaryURL != "cloudinary://API_KEY:API_SECRET@CLOUD_NAME" {
		storageService = media.NewCloudinaryService(&cfg.Cloudinary)
	} else {
		storageService = media.NewLocalStorageService("")
	}

	outboxUseCase := usecase.NewOutboxUseCase(outboxRepo, publisherService, cfg.RabbitMQ.Exchange)
	userUseCase := usecase.NewUserUseCase(userRepo, refreshTokenRepo, outboxUseCase, jwtService, bcryptService, oauthService, aesService, storageService, redisService)
	emailUseCase := usecase.NewEmailUseCase(smtpService, cfg.Server.AppName)

	userHandler := http.NewUserHandler(userUseCase)
	oauthHandler := http.NewOAuthHandler(userUseCase)

	outboxWorker := worker.NewOutboxWorker(outboxUseCase)
	consumerWorker := worker.NewConsumerWorker(conn, cfg.RabbitMQ.Exchange, emailUseCase)

	return &Container{
		UserHandler:      *userHandler,
		OauthHandler:     *oauthHandler,
		JWTService:       jwtService,
		OAuthService:     oauthService,
		PublisherService: publisherService,
		OutboxWorker:     outboxWorker,
		ConsumerWorker:   consumerWorker,
	}
}
