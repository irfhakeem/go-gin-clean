package valueobjects

import (
	"go-gin-clean/internal/core/domain/errors"
	"regexp"
)

type Password struct {
	value string `gorm:"column:password;not null"`
}

func NewPassword(password string) (*Password, error) {
	// For hashed passwords, skip validation
	if len(password) > 50 {
		return &Password{value: password}, nil
	}

	if err := isValidPassword(password); err != nil {
		return nil, err
	}

	return &Password{value: password}, nil
}

func (p *Password) String() string {
	return p.value
}

func (p *Password) Value() string {
	return p.value
}

func (p *Password) Equals(other *Password) bool {
	if other == nil {
		return false
	}
	return p.value == other.value
}

func isValidPassword(password string) error {
	if len(password) < 8 {
		return errors.ErrInvalidPasswordLength
	}

	specialCharRegex := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`)
	if !specialCharRegex.MatchString(password) {
		return errors.ErrPasswordWeak
	}

	return nil
}
