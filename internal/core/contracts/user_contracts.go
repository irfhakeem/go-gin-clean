package contracts

import (
	"go-gin-clean/internal/core/domain/enums"
	"io"
	"time"
)

type (
	UserInfo struct {
		ID       int64
		Name     string
		Email    string
		Avatar   string
		Gender   enums.Gender
		IsActive bool
	}

	LoginRequest struct {
		Email    string
		Password string
	}

	LoginResponse struct {
		AccessToken  string
		RefreshToken string
		User         UserInfo
	}

	RegisterRequest struct {
		Name     string
		Email    string
		Password string
	}

	RefreshTokenResponse struct {
		AccessToken  string
		RefreshToken string
	}

	ResetPasswordRequest struct {
		Token       string
		NewPassword string
	}

	ChangePasswordRequest struct {
		OldPassword string
		NewPassword string
	}

	CreateUserRequest struct {
		Name     string
		Email    string
		Password string
		Gender   enums.Gender
	}

	UpdateUserRequest struct {
		Name   *string
		Gender *enums.Gender
		Avatar *FileUpload
	}

	FileUpload struct {
		Filename string
		Size     int64
		Content  io.Reader
	}

	AccessTokenClaims struct {
		UserID    int64
		Email     string
		TokenType string
		ExpiresAt time.Time
		IssuedAt  time.Time
		NotBefore time.Time
		Issuer    string
		Subject   string
	}

	RefreshTokenClaims struct {
		UserID    int64
		TokenType string
		ExpiresAt time.Time
		IssuedAt  time.Time
		NotBefore time.Time
		Issuer    string
		Subject   string
	}
)
