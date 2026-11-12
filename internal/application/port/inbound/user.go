package inbound

import (
	"context"

	"go-gin-clean/internal/domain/policy"
	"go-gin-clean/internal/dto"
)

type UserUseCase interface {
	GetAllUsers(ctx context.Context, actor policy.Actor, query *dto.GetAllUserQuery, offset int) (*dto.PaginationResponse[dto.UserInfo], error)
	GetUserByID(ctx context.Context, actor policy.Actor, id string) (*dto.UserInfo, error)
	CreateUser(ctx context.Context, actor policy.Actor, req *dto.CreateUserRequest) (*dto.UserInfo, error)
	UpdateUser(ctx context.Context, actor policy.Actor, id string, req *dto.UpdateUserRequest) (*dto.UserInfo, error)
	ChangePassword(ctx context.Context, actor policy.Actor, userID string, req *dto.ChangePasswordRequest) error
	ChangeStatus(ctx context.Context, actor policy.Actor, id string, req *dto.ChangeUserStatusRequest) error
	DeleteUser(ctx context.Context, actor policy.Actor, id string) error
}
