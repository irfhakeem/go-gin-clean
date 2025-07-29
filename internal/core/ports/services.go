package ports

import (
	"go-gin-clean/internal/core/domain/entities"
	"go-gin-clean/internal/core/dto"
	"time"
)

// External service interfaces (secondary ports)
type JWTService interface {
	GenerateAccessToken(user *entities.User) (string, time.Time, error)
	GenerateRefreshToken(userID int64) (string, time.Time, error)
	ValidateAccessToken(token string) (*dto.AccessTokenClaims, error)
	ValidateRefreshToken(token string) (*dto.RefreshTokenClaims, error)
}

type BcryptService interface {
	HashPassword(password string) (string, error)
	ValidatePassword(password, hashedPassword string) error
}

type EncryptionService interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

type MailerService interface {
	SendEmail(to string, subject string, body string) error
	LoadTemplate(templateName string, data any) (string, error)
}

type EmailService interface {
	SendVerifyEmail(to string, name string, token string) error
	SendResetPasswordEmail(to string, name string, token string) error
}
