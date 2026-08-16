package security

import (
	"golang.org/x/crypto/bcrypt"
)

type HasherServiceInterface interface {
	HashPassword(password string) (string, error)
	ValidatePassword(password, hashedPassword string) error
}

type BcryptService struct{}

func NewBcryptService() HasherServiceInterface {
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
