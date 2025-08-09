package security

import (
	"go-gin-clean/internal/core/ports"

	"golang.org/x/crypto/bcrypt"
)

type BcryptService struct{}

func NewBcryptService() ports.BcryptService {
	return &BcryptService{}
}

func (p *BcryptService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (p *BcryptService) ValidatePassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
