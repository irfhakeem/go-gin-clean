package security

import (
	"go-gin-clean/internal/application/port"
	"go-gin-clean/internal/domain/auth"
	"go-gin-clean/pkg/config"
	"go-gin-clean/pkg/errors"
	"go-gin-clean/pkg/message"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const (
	accessTokenType  = "access"
	refreshTokenType = "refresh"
)

type JWTMaker struct {
	cfg *config.JWTConfig
}

type accessJWTClaims struct {
	UserID    string `json:"user_id"`
	UserRole  string `json:"user_role"`
	TokenType string `json:"token_type"`

	jwt.RegisteredClaims
}

type refreshJWTClaims struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`

	jwt.RegisteredClaims
}

func NewJWTMaker(cfg *config.JWTConfig) port.TokenMaker {
	return &JWTMaker{cfg: cfg}
}

func (j *JWTMaker) GenerateAccessToken(id uuid.UUID, role string) (string, time.Time, error) {
	now := time.Now()
	expiryAt := now.Add(j.cfg.AccessTokenExpiry)

	claims := jwt.MapClaims{
		"user_id":    id,
		"user_role":  role,
		"token_type": accessTokenType,
		"exp":        expiryAt.Unix(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"iss":        j.cfg.JWTIssuer,
		"sub":        id,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(j.cfg.AccessTokenSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenStr, expiryAt, nil
}

func (j *JWTMaker) GenerateRefreshToken(id uuid.UUID) (string, time.Time, error) {
	now := time.Now()
	expiryAt := now.Add(j.cfg.RefreshTokenExpiry)

	claims := jwt.MapClaims{
		"user_id":    id,
		"token_type": refreshTokenType,
		"exp":        expiryAt.Unix(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"iss":        j.cfg.JWTIssuer,
		"sub":        id,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(j.cfg.RefreshTokenSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenStr, expiryAt, nil
}

func (j *JWTMaker) ValidateAccessToken(tokenStr string) (*auth.AccessTokenClaims, error) {
	var claims accessJWTClaims

	if err := j.parseJWTToken(tokenStr, j.cfg.AccessTokenSecret, &claims); err != nil {
		return nil, err
	}

	if claims.TokenType != accessTokenType {
		return nil, errors.NewAppError(errors.Unauthorized, message.ErrTokenInvalid)
	}

	if claims.UserID == "" ||
		claims.UserRole == "" ||
		claims.Subject == "" ||
		claims.Issuer == "" {
		return nil, errors.NewAppError(errors.Unauthorized, message.ErrInvalidClaims)
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, errors.NewAppError(errors.Unauthorized, message.ErrInvalidClaims)
	}

	if claims.ExpiresAt == nil ||
		claims.IssuedAt == nil ||
		claims.NotBefore == nil {
		return nil, errors.NewAppError(errors.Unauthorized, message.ErrInvalidClaims)
	}

	return &auth.AccessTokenClaims{
		UserID:    userID,
		UserRole:  claims.UserRole,
		TokenType: claims.TokenType,
		ExpiresAt: claims.ExpiresAt.Time,
		IssuedAt:  claims.IssuedAt.Time,
		NotBefore: claims.NotBefore.Time,
		Issuer:    claims.Issuer,
		Subject:   claims.Subject,
	}, nil
}

func (j *JWTMaker) ValidateRefreshToken(tokenStr string) (*auth.RefreshTokenClaims, error) {
	var claims refreshJWTClaims

	if err := j.parseJWTToken(tokenStr, j.cfg.RefreshTokenSecret, &claims); err != nil {
		return nil, err
	}

	if claims.TokenType != refreshTokenType {
		return nil, errors.NewAppError(errors.Unauthorized, message.ErrTokenInvalid)
	}

	if claims.UserID == "" ||
		claims.Subject == "" ||
		claims.Issuer == "" {
		return nil, errors.NewAppError(errors.Unauthorized, message.ErrInvalidClaims)
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, errors.NewAppError(errors.Unauthorized, message.ErrInvalidClaims)
	}

	if claims.ExpiresAt == nil ||
		claims.IssuedAt == nil ||
		claims.NotBefore == nil {
		return nil, errors.NewAppError(errors.Unauthorized, message.ErrInvalidClaims)
	}

	return &auth.RefreshTokenClaims{
		UserID:    userID,
		TokenType: claims.TokenType,
		ExpiresAt: claims.ExpiresAt.Time,
		IssuedAt:  claims.IssuedAt.Time,
		NotBefore: claims.NotBefore.Time,
		Issuer:    claims.Issuer,
		Subject:   claims.Subject,
	}, nil
}

func (j *JWTMaker) parseJWTToken(tokenStr, secret string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.NewAppError(errors.Unauthorized, message.ErrUnexpectedSigningMethod)
			}

			return []byte(secret), nil
		},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
	)

	if err != nil || !token.Valid {
		return errors.NewAppError(errors.Unauthorized, message.ErrInvalidClaims)
	}

	return nil
}
