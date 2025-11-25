package model

import "time"

type AccessTokenClaims struct {
	UserPKID  int64     `json:"user_pkid"`
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
	UserPKID  int64     `json:"user_pkid"`
	TokenType string    `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
	NotBefore time.Time `json:"not_before"`
	Issuer    string    `json:"issuer"`
	Subject   string    `json:"subject"`
}
