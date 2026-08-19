package dto

import (
	"mime/multipart"

	"go-gin-clean/internal/entity"

	"github.com/google/uuid"
)

type (
	UserInfo struct {
		ID       uuid.UUID     `json:"id"`
		Name     string        `json:"name"`
		Email    string        `json:"email"`
		Avatar   string        `json:"avatar,omitempty"`
		Gender   entity.Gender `json:"gender"`
		IsActive bool          `json:"is_active"`
	}

	LoginRequest struct {
		Email    string `json:"email"    binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	LoginResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	RegisterRequest struct {
		Name     string `json:"name"     binding:"required,min=2,max=100"`
		Email    string `json:"email"    binding:"required,email,max=254"`
		Password string `json:"password" binding:"required,password"`
	}

	RefreshTokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	VerifyEmailRequest struct {
		Token string `json:"token" binding:"required"`
	}

	SendVerifyEmailRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	SendResetPasswordRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	ResetPasswordRequest struct {
		Token       string `json:"token"        binding:"required"`
		NewPassword string `json:"new_password" binding:"required,password"`
	}

	ChangePasswordRequest struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,password"`
	}

	CreateUserRequest struct {
		Name     string        `json:"name"     binding:"required,min=2,max=100"`
		Email    string        `json:"email"    binding:"required,email,max=254"`
		Password string        `json:"password" binding:"required,password"`
		Gender   entity.Gender `json:"gender"   binding:"omitempty,gender"`
	}

	UpdateUserRequest struct {
		Name   *string               `form:"name"   binding:"omitempty,min=2,max=100"`
		Gender *entity.Gender        `form:"gender" binding:"omitempty,gender"`
		Avatar *multipart.FileHeader `form:"avatar"`
	}

	ChangeUserStatusRequest struct {
		IsActive *bool `json:"is_active" binding:"required"`
	}
)

func FormatUserInfo(user *entity.User) *UserInfo {
	return &UserInfo{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Gender:   user.Gender,
		IsActive: user.IsActive,
	}
}

func FormatLoginResponse(accessToken, refreshToken string) *LoginResponse {
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}
