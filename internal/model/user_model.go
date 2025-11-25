package model

import (
	"go-gin-clean/internal/entity"
	"mime/multipart"
)

type (
	UserInfo struct {
		PKID     int64         `json:"pkid"`
		Code     string        `json:"code"`
		Name     string        `json:"name"`
		Email    string        `json:"email"`
		Avatar   string        `json:"avatar,omitempty"`
		Gender   entity.Gender `json:"gender"`
		IsActive bool          `json:"is_active"`
	}

	LoginRequest struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	LoginResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	RegisterRequest struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	RefreshTokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	ResetPasswordRequest struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	VerifyEmailRequest struct {
		Token string `json:"token" binding:"required"`
	}

	SendResetPasswordRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	SendVerifyEmailRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	ChangePasswordRequest struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	CreateUserRequest struct {
		Name     string        `json:"name" binding:"required"`
		Email    string        `json:"email" binding:"required,email"`
		Password string        `json:"password" binding:"required,min=8"`
		Gender   entity.Gender `json:"gender,omitempty"`
	}

	UpdateUserRequest struct {
		Name   *string               `form:"name"`
		Gender *entity.Gender        `form:"gender"`
		Avatar *multipart.FileHeader `form:"avatar"`
	}

	ChangeUserStatusRequest struct {
		IsActive bool `json:"is_active" binding:"required"`
	}
)
