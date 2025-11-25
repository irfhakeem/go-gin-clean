package infrastructure

import (
	"go-gin-clean/internal/delivery/http"
	"go-gin-clean/internal/gateway/cache"
	"go-gin-clean/internal/gateway/media"
	"go-gin-clean/internal/gateway/messaging"
	"go-gin-clean/internal/gateway/security"
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
}

func NewContainer(db *gorm.DB, ch *amqp091.Channel, cfg *config.Config) *Container {
	// Init repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Init services
	jwtService := security.NewJWTService(&cfg.JWT)
	passwordService := security.NewBcryptService()
	oauthService := security.NewOAuthService(&cfg.OAuth)
	aesService := security.NewAESService(&cfg.AES)
	cloudinaryService := media.NewCloudinaryService(&cfg.Cloudinary)
	localStorageService := media.NewLocalStorageService("")
	redisService := cache.NewRedisService(&cfg.Redis)

	// init message publisher
	userPublisher := messaging.NewUserPublisher(ch)

	// Init use cases
	userUseCase := usecase.NewUserUseCase(userRepo, refreshTokenRepo, jwtService, passwordService, oauthService, aesService, cloudinaryService, localStorageService, redisService, userPublisher)

	// Init handlers
	userHandler := http.NewUserHandler(userUseCase)
	oauthHandler := http.NewOAuthHandler(userUseCase)

	return &Container{
		UserHandler:  *userHandler,
		OauthHandler: *oauthHandler,
		JWTService:   *jwtService,
		OAuthService: *oauthService,
	}
}
