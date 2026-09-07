package dto

import (
	"mime/multipart"

	"go-gin-clean/internal/domain/entity"
	"go-gin-clean/internal/domain/vo"

	"github.com/google/uuid"
)

type (
	UserInfo struct {
		ID       uuid.UUID `json:"id"`
		Name     string    `json:"name"`
		Email    string    `json:"email"`
		Avatar   string    `json:"avatar,omitempty"`
		Gender   vo.Gender `json:"gender"`
		IsActive bool      `json:"is_active"`
	}

	GetAllUserQuery struct {
		Role     string `form:"role" binding:"omitempty"`
		IsActive *bool  `form:"is_active" binding:"omitempty"`
		PaginationRequest
	}

	ChangePasswordRequest struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,password"`
	}

	CreateUserRequest struct {
		Name     string    `json:"name"     binding:"required,min=2,max=100"`
		Email    string    `json:"email"    binding:"required,email,max=254"`
		Password string    `json:"password" binding:"required,password"`
		Gender   vo.Gender `json:"gender"   binding:"omitempty,gender"`
	}

	UpdateUserRequest struct {
		Name   *string               `form:"name"   binding:"omitempty,min=2,max=100"`
		Gender *vo.Gender            `form:"gender" binding:"omitempty,gender"`
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
