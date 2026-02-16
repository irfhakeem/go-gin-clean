package model

import (
	"go-gin-clean/internal/entity"
	"mime/multipart"

	"github.com/google/uuid"
)

type (
	UserInfo struct {
		ID       uuid.UUID     `json:"id"`
		Code     string        `json:"code"`
		Name     string        `json:"name"`
		Email    string        `json:"email"`
		Avatar   string        `json:"avatar,omitempty"`
		Gender   entity.Gender `json:"gender"`
		IsActive bool          `json:"is_active"`
	}

	LoginRequest struct {
		Email    string `json:"email" binding:"required,email" validate:"required,email"`
		Password string `json:"password" binding:"required" validate:"required"`
	}

	LoginResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	RegisterRequest struct {
		Name     string `json:"name" binding:"required" validate:"required"`
		Email    string `json:"email" binding:"required,email" validate:"required,email"`
		Password string `json:"password" binding:"required" validate:"required,password"`
	}

	RefreshTokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	ResetPasswordRequest struct {
		Token       string `json:"token" binding:"required" validate:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8" validate:"required,password"`
	}

	VerifyEmailRequest struct {
		Token string `json:"token" binding:"required" validate:"required"`
	}

	SendResetPasswordRequest struct {
		Email string `json:"email" binding:"required,email" validate:"required,email"`
	}

	SendVerifyEmailRequest struct {
		Email string `json:"email" binding:"required,email" validate:"required,email"`
	}

	ChangePasswordRequest struct {
		OldPassword string `json:"old_password" binding:"required" validate:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8" validate:"required,password"`
	}

	CreateUserRequest struct {
		Name     string        `json:"name" binding:"required" validate:"required"`
		Email    string        `json:"email" binding:"required,email" validate:"required,email"`
		Password string        `json:"password" binding:"required,min=8" validate:"required,password"`
		Gender   entity.Gender `json:"gender,omitempty" validate:"omitempty,gender"`
	}

	UpdateUserRequest struct {
		Name   *string               `form:"name"`
		Gender *entity.Gender        `form:"gender" validate:"omitempty,gender"`
		Avatar *multipart.FileHeader `form:"avatar"`
	}

	ChangeUserStatusRequest struct {
		IsActive bool `json:"is_active" binding:"required" validate:"required"`
	}
)
