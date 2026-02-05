package model

import (
	"time"

	"github.com/google/uuid"
)

type AccessTokenClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	UserCode  string    `json:"user_code"`
	UserRole  string    `json:"user_role"`
	TokenType string    `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
	NotBefore time.Time `json:"not_before"`
	Issuer    string    `json:"issuer"`
	Subject   string    `json:"subject"`
}

type RefreshTokenClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenType string    `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
	NotBefore time.Time `json:"not_before"`
	Issuer    string    `json:"issuer"`
	Subject   string    `json:"subject"`
}
