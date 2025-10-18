package dto

import (
	"go-gin-clean/internal/core/domain/enums"
	"mime/multipart"
)

type (
	UserInfo struct {
		ID       int64        `json:"id"`
		Name     string       `json:"name"`
		Email    string       `json:"email"`
		Avatar   string       `json:"avatar,omitempty"`
		Gender   enums.Gender `json:"gender"`
		IsActive bool         `json:"is_active"`
	}

	LoginRequest struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	LoginResponse struct {
		AccessToken  string   `json:"access_token"`
		RefreshToken string   `json:"refresh_token"`
		User         UserInfo `json:"user"`
	}

	RegisterRequest struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	RefreshTokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	SendResetPasswordRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	ResetPasswordRequest struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	SendVerifyEmailRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	VerifyEmailRequest struct {
		Token string `json:"token" binding:"required"`
	}

	ChangePasswordRequest struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	CreateUserRequest struct {
		Name     string       `json:"name" binding:"required"`
		Email    string       `json:"email" binding:"required,email"`
		Password string       `json:"password" binding:"required,min=8"`
		Gender   enums.Gender `json:"gender,omitempty"`
	}

	UpdateUserRequest struct {
		Name   *string               `form:"name" binding:"omitempty"`
		Gender *enums.Gender         `form:"gender" binding:"omitempty"`
		Avatar *multipart.FileHeader `form:"avatar" binding:"omitempty"`
	}
)
