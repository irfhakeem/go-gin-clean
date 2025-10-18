package ports

import (
	"go-gin-clean/internal/core/contracts"
	"go-gin-clean/internal/core/domain/entities"
	"io"
	"time"
)

// External service interfaces (secondary ports)
type JWTService interface {
	GenerateAccessToken(user *entities.User) (string, time.Time, error)
	GenerateRefreshToken(userID int64) (string, time.Time, error)
	ValidateAccessToken(token string) (*contracts.AccessTokenClaims, error)
	ValidateRefreshToken(token string) (*contracts.RefreshTokenClaims, error)
}

type BcryptService interface {
	HashPassword(password string) (string, error)
	ValidatePassword(password, hashedPassword string) error
}

type EncryptionService interface {
	EncryptInternal(plaintext string) (string, error)
	DecryptInternal(ciphertext string) (string, error)
	EncryptURLSafe(plaintext string) (string, error)
	DecryptURLSafe(ciphertext string) (string, error)
}

type MailerService interface {
	SendEmail(to string, subject string, body string) error
	LoadTemplate(templateName string, data any) (string, error)
}

type MediaService interface {
	UploadFile(filename string, size int64, content io.Reader, filePath string) (*string, error)
	DeleteFile(filePath string) error
}
