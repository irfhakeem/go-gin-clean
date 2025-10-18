package ports

import (
	"context"
	"go-gin-clean/internal/core/contracts"
)

// Use case interfaces (primary ports)
type UserUseCase interface {
	Login(ctx context.Context, req *contracts.LoginRequest) (*contracts.LoginResponse, error)
	Register(ctx context.Context, req *contracts.RegisterRequest) error
	RefreshToken(ctx context.Context, refreshToken string) (*contracts.RefreshTokenResponse, error)
	Logout(ctx context.Context, userID int64) error
	VerifyEmail(ctx context.Context, token string) error
	SendVerifyEmail(ctx context.Context, email string) error
	SendResetPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, req *contracts.ResetPasswordRequest) error
	GetAllUsers(ctx context.Context, page, pageSize int, search string) (*contracts.PaginationResponse[contracts.UserInfo], error)
	GetUserByID(ctx context.Context, userID int64) (*contracts.UserInfo, error)
	CreateUser(ctx context.Context, req *contracts.CreateUserRequest) (*contracts.UserInfo, error)
	UpdateUser(ctx context.Context, userID int64, req *contracts.UpdateUserRequest) (*contracts.UserInfo, error)
	ChangePassword(ctx context.Context, userID int64, req *contracts.ChangePasswordRequest) error
	DeleteUser(ctx context.Context, userID int64) error
}

type EmailUseCase interface {
	SendVerifyEmail(toEmail, toName, verifyToken string) error
	SendResetPasswordEmail(toEmail, toName, resetToken string) error
}
