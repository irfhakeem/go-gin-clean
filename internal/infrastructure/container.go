package infrastructure

import (
	"go-gin-clean/internal/adapters/secondary/database"
	"go-gin-clean/internal/adapters/secondary/mailer"
	"go-gin-clean/internal/adapters/secondary/media"
	"go-gin-clean/internal/adapters/secondary/security"
	"go-gin-clean/internal/core/ports"
	"go-gin-clean/internal/core/usecases"
	"go-gin-clean/pkg/config"

	"gorm.io/gorm"
)

type Container struct {
	UserUseCase   ports.UserUseCase
	EmailUseCase  ports.EmailUseCase
	JWTService    ports.JWTService
	MailerService ports.MailerService
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	// Init repositories
	userRepo := database.NewUserRepository(db)
	refreshTokenRepo := database.NewRefreshTokenRepository(db)

	// Init services
	jwtService := security.NewJWTService(&cfg.JWT)
	bcryptService := security.NewBcryptService()
	aesService := security.NewAESService(&cfg.AES)
	smtpService := mailer.NewSMTPService(&cfg.Mailer)
	localStorageService := media.NewLocalStorageService()

	// Init use cases
	emailUseCase := usecases.NewEmailUseCase(smtpService)
	userUseCase := usecases.NewUserUseCase(userRepo, emailUseCase, refreshTokenRepo, jwtService, bcryptService, aesService, localStorageService)

	return &Container{
		UserUseCase:  userUseCase,
		EmailUseCase: emailUseCase,
		JWTService:   jwtService,
	}
}
