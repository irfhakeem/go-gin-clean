package dto

import "time"

type AccessTokenClaims struct {
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	TokenType string    `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
	NotBefore time.Time `json:"not_before"`
	Issuer    string    `json:"issuer"`
	Subject   string    `json:"subject"`
}

type RefreshTokenClaims struct {
	UserID    int64     `json:"user_id"`
	TokenType string    `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
	NotBefore time.Time `json:"not_before"`
	Issuer    string    `json:"issuer"`
	Subject   string    `json:"subject"`
}
