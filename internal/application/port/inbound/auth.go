package inbound

import (
	"context"

	"go-gin-clean/internal/dto"
)

type AuthUseCase interface {
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	Register(ctx context.Context, req *dto.RegisterRequest) error
	RefreshToken(ctx context.Context, hashedRefreshToken string) (*dto.RefreshTokenResponse, error)
	Logout(ctx context.Context, id string) error
	SendVerifyEmail(ctx context.Context, req *dto.SendVerifyEmailRequest) error
	VerifyEmail(ctx context.Context, token string) error
	SendResetPassword(ctx context.Context, req *dto.SendResetPasswordRequest) error
	ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error
}
