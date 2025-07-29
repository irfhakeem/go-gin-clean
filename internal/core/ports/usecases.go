package ports

import (
	"context"
	"go-gin-clean/internal/core/dto"
)

// Use case interfaces (primary ports)
type UserUseCase interface {
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	Register(ctx context.Context, req *dto.RegisterRequest) error
	RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshTokenResponse, error)
	Logout(ctx context.Context, userID int64) error
	VerifyEmail(ctx context.Context, token string) error
	SendVerifyEmail(ctx context.Context, email string) error
	SendResetPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error
	GetAllUsers(ctx context.Context, page, pageSize int, search string) (*dto.PaginationResponse[dto.UserInfo], error)
	GetUserByID(ctx context.Context, userID int64) (*dto.UserInfo, error)
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserInfo, error)
	UpdateUser(ctx context.Context, userID int64, req *dto.UpdateUserRequest) (*dto.UserInfo, error)
	ChangePassword(ctx context.Context, userID int64, req *dto.ChangePasswordRequest) error
	DeleteUser(ctx context.Context, userID int64) error
}
