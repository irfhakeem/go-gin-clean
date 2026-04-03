package infrastructure

import (
	"go-gin-clean/internal/delivery/http"
	"go-gin-clean/internal/gateway/cache"
	"go-gin-clean/internal/gateway/media"
	"go-gin-clean/internal/gateway/messaging"
	"go-gin-clean/internal/gateway/security"
	"go-gin-clean/internal/infrastructure/worker"
	"go-gin-clean/internal/repository"
	"go-gin-clean/internal/usecase"
	"go-gin-clean/pkg/config"

	"github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type Container struct {
	UserHandler  http.UserHandler
	OauthHandler http.OAuthHandler
	JWTService   security.JWTService
	OAuthService security.OAuthService
	OutboxWorker *worker.OutboxWorker
}

func NewContainer(db *gorm.DB, ch *amqp091.Channel, cfg *config.Config) *Container {
	// Init repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	outboxRepo := repository.NewOutboxRepository(db)

	// Init services
	jwtService := security.NewJWTService(&cfg.JWT)
	passwordService := security.NewBcryptService()
	oauthService := security.NewOAuthService(&cfg.OAuth)
	aesService := security.NewAESService(&cfg.AES)
	cloudinaryService := media.NewCloudinaryService(&cfg.Cloudinary)
	localStorageService := media.NewLocalStorageService("")
	redisService := cache.NewRedisService(&cfg.Redis)
	publisherService := messaging.NewPublisherService(ch, cfg.RabbitMQ.Exchange)

	// Init use cases
	outboxUseCase := usecase.NewOutboxUseCase(outboxRepo, publisherService, cfg.RabbitMQ.Exchange)
	userUseCase := usecase.NewUserUseCase(userRepo, refreshTokenRepo, outboxUseCase, jwtService, passwordService, oauthService, aesService, cloudinaryService, localStorageService, redisService)

	// Init handlers
	userHandler := http.NewUserHandler(userUseCase)
	oauthHandler := http.NewOAuthHandler(userUseCase)

	// Init workers
	outboxWorker := worker.NewOutboxWorker(outboxUseCase)

	return &Container{
		UserHandler:  *userHandler,
		OauthHandler: *oauthHandler,
		JWTService:   *jwtService,
		OAuthService: *oauthService,
		OutboxWorker: outboxWorker,
	}
}
