package security

import (
	"time"

	"go-gin-clean/internal/entity"
	"go-gin-clean/internal/model"
	"go-gin-clean/pkg/config"

	pkgerror "go-gin-clean/pkg/error"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type JWTServiceInterface interface {
	GenerateAccessToken(user *entity.User) (string, time.Time, error)
	GenerateRefreshToken(userID string) (string, time.Time, error)
	ValidateAccessToken(tokenString string) (*model.AccessTokenClaims, error)
	ValidateRefreshToken(tokenString string) (*model.RefreshTokenClaims, error)
}

type JWTService struct {
	cfg *config.JWTConfig
}

func NewJWTService(cfg *config.JWTConfig) *JWTService {
	return &JWTService{cfg: cfg}
}

func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func (j *JWTService) GenerateAccessToken(user *entity.User) (string, time.Time, error) {
	now := time.Now()
	expiryAt := now.Add(j.cfg.AccessTokenExpiry)

	claims := jwt.MapClaims{
		"user_id":    user.ID.String(),
		"user_role":  user.Role,
		"token_type": "access",
		"exp":        expiryAt.Unix(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"iss":        j.cfg.JWTIssuer,
		"sub":        user.ID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.cfg.AccessTokenSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiryAt, nil
}

func (j *JWTService) GenerateRefreshToken(userID string) (string, time.Time, error) {
	now := time.Now()
	expiryAt := now.Add(j.cfg.RefreshTokenExpiry)

	claims := jwt.MapClaims{
		"user_id":    userID,
		"token_type": "refresh",
		"exp":        expiryAt.Unix(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"iss":        j.cfg.JWTIssuer,
		"sub":        userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.cfg.RefreshTokenSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiryAt, nil
}

func (j *JWTService) ValidateAccessToken(tokenString string) (*model.AccessTokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, pkgerror.ErrUnexpectedSigningMethod
		}
		return []byte(j.cfg.AccessTokenSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, pkgerror.ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, pkgerror.ErrInvalidClaims
	}

	tokenType, ok := claims["token_type"].(string)
	if !ok || tokenType != "access" {
		return nil, pkgerror.ErrTokenInvalid
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, pkgerror.ErrInvalidClaims
	}

	userRole, ok := claims["user_role"].(string)
	if !ok {
		return nil, pkgerror.ErrInvalidClaims
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, pkgerror.ErrInvalidClaims
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, pkgerror.ErrInvalidClaims
	}

	nbf, ok := claims["nbf"].(float64)
	if !ok {
		return nil, pkgerror.ErrInvalidClaims
	}

	iss, ok := claims["iss"].(string)
	if !ok {
		return nil, pkgerror.ErrInvalidClaims
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, pkgerror.ErrInvalidClaims
	}

	parsedUUID, err := ParseUUID(userID)
	if err != nil {
		return nil, pkgerror.ErrInvalidClaims
	}

	return &model.AccessTokenClaims{
		UserID:    parsedUUID,
		UserRole:  userRole,
		TokenType: tokenType,
		ExpiresAt: time.Unix(int64(exp), 0),
		IssuedAt:  time.Unix(int64(iat), 0),
		NotBefore: time.Unix(int64(nbf), 0),
		Issuer:    iss,
		Subject:   sub,
	}, nil
}

func (j *JWTService) ValidateRefreshToken(tokenString string) (*model.RefreshTokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, pkgerror.ErrUnexpectedSigningMethod
		}
		return []byte(j.cfg.RefreshTokenSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, pkgerror.ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, pkgerror.ErrInvalidClaims
	}

	tokenType, ok := claims["token_type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, pkgerror.ErrTokenInvalid
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, pkgerror.ErrInvalidClaims
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, pkgerror.ErrInvalidClaims
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, pkgerror.ErrInvalidClaims
	}

	nbf, ok := claims["nbf"].(float64)
	if !ok {
		return nil, pkgerror.ErrInvalidClaims
	}

	iss, ok := claims["iss"].(string)
	if !ok {
		return nil, pkgerror.ErrInvalidClaims
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, pkgerror.ErrInvalidClaims
	}

	parsedUUID, err := ParseUUID(userID)
	if err != nil {
		return nil, pkgerror.ErrInvalidClaims
	}

	return &model.RefreshTokenClaims{
		UserID:    parsedUUID,
		TokenType: tokenType,
		ExpiresAt: time.Unix(int64(exp), 0),
		IssuedAt:  time.Unix(int64(iat), 0),
		NotBefore: time.Unix(int64(nbf), 0),
		Issuer:    iss,
		Subject:   sub,
	}, nil
}
