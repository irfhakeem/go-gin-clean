package port

import (
	"go-gin-clean/internal/domain/auth"
	"time"

	"github.com/google/uuid"
)

type TokenMaker interface {
	GenerateAccessToken(userID uuid.UUID, userRole string) (string, time.Time, error)
	GenerateRefreshToken(userID uuid.UUID) (string, time.Time, error)
	ValidateAccessToken(tokenString string) (*auth.AccessTokenClaims, error)
	ValidateRefreshToken(tokenString string) (*auth.RefreshTokenClaims, error)
}

type Hasher interface {
	HashPassword(password string) (string, error)
	ValidatePassword(password, hashedPassword string) error
}

type Encryptor interface {
	EncryptInternal(plainText string) (string, error)
	DecryptInternal(cipherText string) (string, error)
	EncryptURLSafe(plainText string) (string, error)
	DecryptURLSafe(cipherText string) (string, error)
}
