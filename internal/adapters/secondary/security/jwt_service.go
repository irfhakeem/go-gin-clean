package security

import (
	"go-gin-clean/internal/core/domain/entities"
	"go-gin-clean/internal/core/domain/errors"
	"go-gin-clean/internal/core/dto"
	"go-gin-clean/internal/core/ports"
	"go-gin-clean/pkg/config"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTService struct {
	cfg *config.JWTConfig
}

func NewJWTService(cfg *config.JWTConfig) ports.JWTService {
	return &JWTService{cfg: cfg}
}

func (j *JWTService) GenerateAccessToken(user *entities.User) (string, time.Time, error) {
	now := time.Now()
	expiryAt := now.Add(j.cfg.AccessTokenExpiry)

	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"email":      user.Email.Value(),
		"token_type": "access",
		"exp":        expiryAt.Unix(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"iss":        "go-gin-clean",
		"sub":        strconv.FormatInt(user.ID, 10),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.cfg.AccessTokenSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiryAt, nil
}

func (j *JWTService) GenerateRefreshToken(userID int64) (string, time.Time, error) {
	now := time.Now()
	expiryAt := now.Add(j.cfg.RefreshTokenExpiry)

	claims := jwt.MapClaims{
		"user_id":    userID,
		"token_type": "refresh",
		"exp":        expiryAt.Unix(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"iss":        "go-gin-clean",
		"sub":        strconv.FormatInt(userID, 10),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.cfg.RefreshTokenSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiryAt, nil
}

func (j *JWTService) ValidateAccessToken(tokenString string) (*dto.AccessTokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.ErrUnexpectedSigningMethod
		}
		return []byte(j.cfg.AccessTokenSecret), nil
	})

	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	if !token.Valid {
		return nil, errors.ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.ErrInvalidClaims
	}

	tokenType, ok := claims["token_type"].(string)
	if !ok || tokenType != "access" {
		return nil, errors.ErrTokenInvalid
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.ErrInvalidClaims
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.ErrInvalidClaims
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.ErrInvalidClaims
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, errors.ErrInvalidClaims
	}

	nbf, ok := claims["nbf"].(float64)
	if !ok {
		return nil, errors.ErrInvalidClaims
	}

	iss, ok := claims["iss"].(string)
	if !ok {
		return nil, errors.ErrInvalidClaims
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.ErrInvalidClaims
	}

	return &dto.AccessTokenClaims{
		UserID:    int64(userID),
		Email:     email,
		TokenType: tokenType,
		ExpiresAt: time.Unix(int64(exp), 0),
		IssuedAt:  time.Unix(int64(iat), 0),
		NotBefore: time.Unix(int64(nbf), 0),
		Issuer:    iss,
		Subject:   sub,
	}, nil
}

func (j *JWTService) ValidateRefreshToken(tokenString string) (*dto.RefreshTokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.ErrUnexpectedSigningMethod
		}
		return []byte(j.cfg.RefreshTokenSecret), nil
	})

	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	if !token.Valid {
		return nil, errors.ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.ErrInvalidClaims
	}

	tokenType, ok := claims["token_type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, errors.ErrTokenInvalid
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.ErrInvalidClaims
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.ErrInvalidClaims
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, errors.ErrInvalidClaims
	}

	nbf, ok := claims["nbf"].(float64)
	if !ok {
		return nil, errors.ErrInvalidClaims
	}

	iss, ok := claims["iss"].(string)
	if !ok {
		return nil, errors.ErrInvalidClaims
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.ErrInvalidClaims
	}

	return &dto.RefreshTokenClaims{
		UserID:    int64(userID),
		TokenType: tokenType,
		ExpiresAt: time.Unix(int64(exp), 0),
		IssuedAt:  time.Unix(int64(iat), 0),
		NotBefore: time.Unix(int64(nbf), 0),
		Issuer:    iss,
		Subject:   sub,
	}, nil
}
