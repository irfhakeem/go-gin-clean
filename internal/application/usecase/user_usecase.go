package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"gorm.io/gorm"

	"go-gin-clean/internal/application/port"
	"go-gin-clean/internal/domain/entity"
	"go-gin-clean/internal/domain/policy"
	"go-gin-clean/internal/dto"
	"go-gin-clean/pkg/config"
	pkgerrors "go-gin-clean/pkg/errors"
	"go-gin-clean/pkg/logger"
	"go-gin-clean/pkg/message"
	"go-gin-clean/pkg/utils"

	"go.uber.org/zap"

	"github.com/google/uuid"
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

type userUseCase struct {
	policy policy.UserPolicy

	userRepo         port.UserRepository
	refreshTokenRepo port.RefreshTokenRepository
	outboxUseCase    OutboxUseCase

	hasher  port.Hasher
	storage port.Storage
	cache   port.Cache

	cfg *config.ServerConfig
}

func NewUserUseCase(
	policy policy.UserPolicy,
	userRepo port.UserRepository,
	refreshTokenRepo port.RefreshTokenRepository,
	outboxUseCase OutboxUseCase,
	hasher port.Hasher,
	storage port.Storage,
	cache port.Cache,
	cfg *config.ServerConfig,
) UserUseCase {
	return &userUseCase{
		policy:           policy,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		outboxUseCase:    outboxUseCase,
		hasher:           hasher,
		storage:          storage,
		cache:            cache,
		cfg:              cfg,
	}
}

func (u *userUseCase) checkAccess(actor policy.Actor, action policy.Action, targetID string) error {
	scope := u.policy.Scope(actor, action)
	if scope.Type == policy.ScopeNone {
		return pkgerrors.NewAppError(pkgerrors.Forbidden, message.ErrForbidden)
	}

	if scope.Type == policy.ScopeAll {
		return nil
	}

	targetUUID, err := uuid.Parse(targetID)
	if err != nil {
		return pkgerrors.NewAppError(pkgerrors.NotFound, message.ErrUserNotFound)
	}

	if !slices.Contains(scope.UserIDs, targetUUID) {
		return pkgerrors.NewAppError(pkgerrors.Forbidden, message.ErrForbidden)
	}

	return nil
}

func (u *userUseCase) GetAllUsers(ctx context.Context, actor policy.Actor, query *dto.GetAllUserQuery, offset int) (*dto.PaginationResponse[dto.UserInfo], error) {
	limit := query.PerPage
	page := query.Page

	scope := u.policy.Scope(actor, policy.ActionRead)
	if scope.Type == policy.ScopeNone {
		return nil, pkgerrors.NewAppError(pkgerrors.Forbidden, message.ErrForbidden)
	}

	cacheKey := fmt.Sprintf("users:all:%s:%s:page:%d:size:%d:search:%s:role:%s", actor.Role, actor.ID, page, limit, query.Search, query.Role)

	var cachedResult dto.PaginationResponse[dto.UserInfo]
	if err := u.cache.Get(ctx, cacheKey, &cachedResult); err == nil {
		return &cachedResult, nil
	} else if !errors.Is(err, pkgerrors.ErrCacheMiss) {
		logger.Error("cache error", zap.String("key", cacheKey), zap.Error(err))
	}

	if scope.Type == policy.ScopeFiltered && len(scope.UserIDs) == 0 {
		return dto.NewPaginationResponse([]dto.UserInfo{}, page, limit, 0), nil
	}

	params := port.FindAllUsersParams{
		Limit:    query.PerPage,
		Offset:   offset,
		Search:   query.Search,
		SortBy:   query.SortBy,
		Sort:     query.Sort,
		Role:     query.Role,
		IsActive: query.IsActive,
	}

	if scope.Type == policy.ScopeFiltered {
		params.IDs = scope.UserIDs
	}

	users, total, err := u.userRepo.FindAll(ctx, params)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrGetAllUsers, err)
	}

	userInfos := make([]dto.UserInfo, len(users))
	for i, user := range users {
		userInfos[i] = *dto.FormatUserInfo(user)
	}

	result := dto.NewPaginationResponse(userInfos, page, limit, int(total))

	if err := u.cache.SetWithExpiration(ctx, cacheKey, result, 5*time.Minute); err != nil {
		logger.Error("failed to cache result", zap.Error(err))
	}

	return result, nil
}

func (u *userUseCase) GetUserByID(ctx context.Context, actor policy.Actor, id string) (*dto.UserInfo, error) {
	if err := u.checkAccess(actor, policy.ActionRead, id); err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("user:id:%s", id)

	var cachedUser dto.UserInfo
	if err := u.cache.Get(ctx, cacheKey, &cachedUser); err == nil {
		return &cachedUser, nil
	} else if !errors.Is(err, pkgerrors.ErrCacheMiss) {
		logger.Error("cache error", zap.String("key", cacheKey), zap.Error(err))
	}

	user, err := u.userRepo.FindByID(ctx, id)
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, pkgerrors.NewAppError(pkgerrors.NotFound, message.ErrUserNotFound)
	case err != nil:
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrGetUserInformation, err)
	}

	if user == nil || !user.IsActive || !user.IsVerified {
		return nil, pkgerrors.NewAppError(pkgerrors.NotFound, message.ErrUserNotFound)
	}

	userInfo := dto.FormatUserInfo(user)

	if err := u.cache.SetWithExpiration(ctx, cacheKey, userInfo, 5*time.Minute); err != nil {
		logger.Error("failed to cache result", zap.Error(err))
	}

	return userInfo, nil
}

func (u *userUseCase) CreateUser(ctx context.Context, actor policy.Actor, req *dto.CreateUserRequest) (*dto.UserInfo, error) {
	scope := u.policy.Scope(actor, policy.ActionCreate)
	if scope.Type == policy.ScopeNone {
		return nil, pkgerrors.NewAppError(pkgerrors.Forbidden, message.ErrForbidden)
	}

	if u.userRepo.ExistByEmail(ctx, req.Email) {
		return nil, pkgerrors.WrapAppError(pkgerrors.Conflict, message.ErrCreateUser, pkgerrors.NewAppError(pkgerrors.Conflict, message.ErrEmailAlreadyExists))
	}

	hashedPassword, err := u.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrCreateUser, err)
	}

	userData, err := entity.NewUser(req.Name, req.Email, hashedPassword)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrCreateUser, err)
	}

	savedUser, err := u.userRepo.Create(ctx, userData)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrCreateUser, err)
	}

	if err := u.cache.DeletePattern(ctx, "users:all:*"); err != nil {
		logger.Error("failed to invalidate users cache", zap.Error(err))
	}

	return dto.FormatUserInfo(savedUser), nil
}

func (u *userUseCase) UpdateUser(ctx context.Context, actor policy.Actor, id string, req *dto.UpdateUserRequest) (*dto.UserInfo, error) {
	if err := u.checkAccess(actor, policy.ActionUpdate, id); err != nil {
		return nil, err
	}

	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.NewAppError(pkgerrors.NotFound, message.ErrUserNotFound)
		}
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrUpdateUser, err)
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Avatar != nil {
		allowedExtensions := []string{".jpg", ".jpeg", ".png"}
		if !utils.IsValidExtension(req.Avatar.Filename, allowedExtensions) {
			return nil, pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrUpdateUser, pkgerrors.NewAppError(pkgerrors.BadRequest, message.ErrUnsupportedImageType))
		}

		path, err := u.storage.UploadFile(
			ctx,
			fmt.Sprintf("avatar_%s_%d.jpg", user.ID.String(), time.Now().Unix()),
			req.Avatar.Size,
			*req.Avatar,
			"users/"+user.ID.String()+"/avatar/",
		)
		if err != nil || path == nil {
			return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrUpdateUser, err)
		}

		user.Avatar = *path
	}

	if req.Gender != nil {
		user.Gender = *req.Gender
	}

	updatedUser, err := u.userRepo.Update(ctx, user)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrUpdateUser, err)
	}

	cacheKey := fmt.Sprintf("user:id:%s", id)
	if err := u.cache.Delete(ctx, cacheKey); err != nil {
		logger.Error("failed to invalidate user cache", zap.Error(err))
	}
	if err := u.cache.DeletePattern(ctx, "users:all:*"); err != nil {
		logger.Error("failed to invalidate users cache", zap.Error(err))
	}

	return dto.FormatUserInfo(updatedUser), nil
}

func (u *userUseCase) ChangePassword(ctx context.Context, actor policy.Actor, userID string, req *dto.ChangePasswordRequest) error {
	if err := u.checkAccess(actor, policy.ActionUpdate, userID); err != nil {
		return err
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkgerrors.NewAppError(pkgerrors.NotFound, message.ErrUserNotFound)
		}
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrUpdateUserPassword, err)
	}

	if err := u.hasher.ValidatePassword(req.OldPassword, user.Password); err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrUpdateUserPassword, pkgerrors.NewAppError(pkgerrors.BadRequest, message.ErrPasswordNotMatch))
	}

	hashedPassword, err := u.hasher.HashPassword(req.NewPassword)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrUpdateUserPassword, err)
	}

	user.SetPassword(hashedPassword)
	if _, err = u.userRepo.Update(ctx, user); err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrUpdateUserPassword, err)
	}

	return nil
}

func (u *userUseCase) ChangeStatus(ctx context.Context, actor policy.Actor, id string, req *dto.ChangeUserStatusRequest) error {
	if err := u.checkAccess(actor, policy.ActionUpdate, id); err != nil {
		return err
	}

	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkgerrors.NewAppError(pkgerrors.NotFound, message.ErrUserNotFound)
		}
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrUpdateUserStatus, err)
	}

	user.IsActive = *req.IsActive
	if _, err = u.userRepo.Update(ctx, user); err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrUpdateUserStatus, err)
	}

	cacheKey := fmt.Sprintf("user:id:%s", id)
	if err := u.cache.Delete(ctx, cacheKey); err != nil {
		logger.Error("failed to invalidate user cache", zap.Error(err))
	}
	if err := u.cache.DeletePattern(ctx, "users:all:*"); err != nil {
		logger.Error("failed to invalidate users cache", zap.Error(err))
	}

	return nil
}

func (u *userUseCase) DeleteUser(ctx context.Context, actor policy.Actor, id string) error {
	if err := u.checkAccess(actor, policy.ActionDelete, id); err != nil {
		return err
	}

	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkgerrors.NewAppError(pkgerrors.NotFound, message.ErrUserNotFound)
		}
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrDeleteUser, err)
	}

	if err = u.refreshTokenRepo.RevokeAllByUserID(ctx, user.ID.String()); err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrDeleteUser, err)
	}

	if err = u.userRepo.Delete(ctx, user.ID.String()); err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrDeleteUser, err)
	}

	cacheKey := fmt.Sprintf("user:id:%s", id)
	if err := u.cache.Delete(ctx, cacheKey); err != nil {
		logger.Error("failed to invalidate user cache", zap.Error(err))
	}
	if err := u.cache.DeletePattern(ctx, "users:all:*"); err != nil {
		logger.Error("failed to invalidate users cache", zap.Error(err))
	}

	return nil
}
