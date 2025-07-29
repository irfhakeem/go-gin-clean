package infrastructure

import (
	"go-gin-clean/internal/adapters/secondary/database"
	"go-gin-clean/internal/adapters/secondary/security"
	"go-gin-clean/internal/core/ports"
	"go-gin-clean/internal/core/usecases"
	"go-gin-clean/pkg/config"

	"gorm.io/gorm"
)

type Container struct {
	UserUseCase ports.UserUseCase
	JWTService  ports.JWTService
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	// Init repositories
	userRepo := database.NewUserRepository(db)
	refreshTokenRepo := database.NewRefreshTokenRepository(db)

	// Init services
	jwtService := security.NewJWTService(&cfg.JWT)
	passwordService := security.NewBcryptService()

	// Init use cases
	userUseCase := usecases.NewUserUseCase(
		userRepo,
		refreshTokenRepo,
		jwtService,
		passwordService,
	)

	return &Container{
		UserUseCase: userUseCase,
		JWTService:  jwtService,
	}
}
