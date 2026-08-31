package auth

import (
	"time"

	"github.com/google/uuid"
)

type AccessTokenClaims struct {
	UserID    uuid.UUID
	UserRole  string
	TokenType string
	ExpiresAt time.Time
	IssuedAt  time.Time
	NotBefore time.Time
	Issuer    string
	Subject   string
}

type RefreshTokenClaims struct {
	UserID    uuid.UUID
	TokenType string
	ExpiresAt time.Time
	IssuedAt  time.Time
	NotBefore time.Time
	Issuer    string
	Subject   string
}
