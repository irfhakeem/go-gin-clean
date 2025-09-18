package ports

import (
	"go-gin-clean/internal/core/domain/entities"
	"go-gin-clean/internal/core/dto"
	"mime/multipart"
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
	UploadFile(fileHeader multipart.FileHeader, filePath string) (*string, error)
	DeleteFile(filePath string) error
}
