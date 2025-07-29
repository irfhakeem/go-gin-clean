package valueobjects

import (
	"go-gin-clean/internal/core/domain/errors"
	"regexp"
	"strings"
)

type EmailAddress struct {
	value string `gorm:"column:email;uniqueIndex;not null"`
}

func NewEmailAddress(email string) (*EmailAddress, error) {
	if !isValidEmail(email) {
		return nil, errors.ErrInvalidEmail
	}
	return &EmailAddress{value: email}, nil
}

func (e *EmailAddress) String() string {
	return e.value
}

func (e *EmailAddress) Value() string {
	return e.value
}

func (e *EmailAddress) Equals(other *EmailAddress) bool {
	if other == nil {
		return false
	}
	return e.value == other.value
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	return emailRegex.MatchString(strings.ToLower(email))
}
