package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"go-gin-clean/internal/application/port"
	"go-gin-clean/internal/domain/entity"
	"go-gin-clean/internal/dto"
	"go-gin-clean/pkg/config"
	pkgerrors "go-gin-clean/pkg/errors"
	"go-gin-clean/pkg/logger"
	"go-gin-clean/pkg/message"
	"go-gin-clean/pkg/utils"

	"go.uber.org/zap"
)

type UserUseCase interface {
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	Register(ctx context.Context, req *dto.RegisterRequest) error
	RefreshToken(ctx context.Context, hashedRefreshToken string) (*dto.RefreshTokenResponse, error)
	Logout(ctx context.Context, id string) error
	SendVerifyEmail(ctx context.Context, req *dto.SendVerifyEmailRequest) error
	VerifyEmail(ctx context.Context, token string) error
	SendResetPassword(ctx context.Context, req *dto.SendResetPasswordRequest) error
	ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error
	GetOAuthLoginURL(ctx context.Context, provider, appID, platform string) (*dto.OAuthUrlResponse, error)
	GetOAuthRedirectURL(appID, platform string) string
	HandleOAuthCallback(ctx context.Context, provider string, req *dto.OAuthCallbackRequest) (*dto.LoginResponse, string, string, error)
	GetAllUsers(ctx context.Context, page, pageSize int, search string) (*dto.PaginationResponse[dto.UserInfo], error)
	GetUserByID(ctx context.Context, id string) (*dto.UserInfo, error)
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserInfo, error)
	UpdateUser(ctx context.Context, id string, req *dto.UpdateUserRequest) (*dto.UserInfo, error)
	ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error
	ChangeStatus(ctx context.Context, id string, req *dto.ChangeUserStatusRequest) error
	DeleteUser(ctx context.Context, id string) error
}

type userUseCase struct {
	userRepo         port.UserRepository
	refreshTokenRepo port.RefreshTokenRepository
	outboxUseCase    OutboxUseCase

	jwt       port.TokenMaker
	hasher    port.Hasher
	oauth     port.OAuthProvider
	encryptor port.Encryptor
	storage   port.Storage
	cache     port.Cache

	cfg *config.ServerConfig
}

func NewUserUseCase(
	userRepo port.UserRepository,
	refreshTokenRepo port.RefreshTokenRepository,
	outboxUseCase OutboxUseCase,
	jwt port.TokenMaker,
	hasher port.Hasher,
	oauth port.OAuthProvider,
	encryptor port.Encryptor,
	storage port.Storage,
	cache port.Cache,
	cfg *config.ServerConfig,
) UserUseCase {
	return &userUseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		outboxUseCase:    outboxUseCase,
		jwt:              jwt,
		hasher:           hasher,
		oauth:            oauth,
		encryptor:        encryptor,
		storage:          storage,
		cache:            cache,
		cfg:              cfg,
	}
}

func (u *userUseCase) GetOAuthRedirectURL(appID, platform string) string {
	if platform == "mobile" {
		return u.oauth.GetMobileDeepLinkURL(appID)
	}
	return u.oauth.GetFrontendURL(appID)
}

func (u *userUseCase) GetOAuthLoginURL(ctx context.Context, provider, appID, platform string) (*dto.OAuthUrlResponse, error) {
	switch provider {
	case "google":
		return &dto.OAuthUrlResponse{
			AuthURL: u.oauth.GetGoogleAuthURL(appID, platform),
		}, nil
	default:
		return nil, pkgerrors.NewAppError(pkgerrors.Unprocessable, message.ErrInvalidOAuthProvider)
	}
}

func (u *userUseCase) HandleOAuthCallback(ctx context.Context, provider string, req *dto.OAuthCallbackRequest) (*dto.LoginResponse, string, string, error) {
	var (
		user     *entity.User
		appID    string
		platform string
		err      error
	)

	switch provider {
	case "google":
		user, appID, platform, err = u.oauth.HandleGoogleCallback(ctx, req.State, req.Code)
		if err != nil {
			return nil, appID, platform, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrOAuthCallback, err)
		}
	default:
		return nil, appID, platform, pkgerrors.NewAppError(pkgerrors.Unprocessable, message.ErrInvalidOAuthProvider)
	}

	existingUser, err := u.userRepo.FindByOAuthID(ctx, provider, user.OAuthID)
	if err == nil {
		user = existingUser
	} else {
		existingUserByEmail, err := u.userRepo.FindByEmail(ctx, user.Email)
		if err == nil {
			if err = u.userRepo.UpdateOAuthInfo(ctx, existingUserByEmail.ID.String(), provider, user.OAuthID); err != nil {
				return nil, appID, platform, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrLinkOAuth, err)
			}
			user = existingUserByEmail
		} else {
			user, err = u.userRepo.Create(ctx, user)
			if err != nil {
				return nil, appID, platform, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrOAuthSignUp, err)
			}
		}
	}

	accessToken, _, err := u.jwt.GenerateAccessToken(user.ID, user.Role.String())
	if err != nil {
		return nil, appID, platform, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrAccessToken, err)
	}

	refreshToken, expiryAt, err := u.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, appID, platform, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRefreshToken, err)
	}

	hashedRefreshToken, err := u.encryptor.EncryptInternal(refreshToken)
	if err != nil {
		return nil, appID, platform, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrProcessToken, err)
	}

	tokenData := entity.NewRefreshToken(user.ID, hashedRefreshToken, expiryAt, false, user)
	if err := u.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return nil, appID, platform, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRefreshToken, err)
	}

	return dto.FormatLoginResponse(accessToken, hashedRefreshToken), appID, platform, nil
}

func (u *userUseCase) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Unauthorized, message.ErrLoginFailed, err)
	}

	if user.IsOAuthUser() {
		return nil, pkgerrors.WrapAppError(pkgerrors.Unauthorized, message.ErrLoginFailed, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrOAuthUserUseOAuthLogin))
	}

	if !user.IsActive {
		return nil, pkgerrors.WrapAppError(pkgerrors.Unauthorized, message.ErrLoginFailed, pkgerrors.NewAppError(pkgerrors.NotFound, message.ErrUserNotFound))
	}

	if !user.IsVerified {
		return nil, pkgerrors.WrapAppError(pkgerrors.Unauthorized, message.ErrLoginFailed, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrEmailNotVerified))
	}

	if err := u.hasher.ValidatePassword(req.Password, user.Password); err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Unauthorized, message.ErrLoginFailed, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrPasswordNotMatch))
	}

	accessToken, _, err := u.jwt.GenerateAccessToken(user.ID, user.Role.String())
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrLoginFailed, err)
	}

	refreshToken, expiryAt, err := u.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrLoginFailed, err)
	}

	hashedRefreshToken, err := u.encryptor.EncryptInternal(refreshToken)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrLoginFailed, err)
	}

	tokenData := entity.NewRefreshToken(user.ID, hashedRefreshToken, expiryAt, false, user)
	if err := u.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrLoginFailed, err)
	}

	return dto.FormatLoginResponse(accessToken, hashedRefreshToken), nil
}

func (u *userUseCase) Register(ctx context.Context, req *dto.RegisterRequest) error {
	if exist := u.userRepo.ExistByEmail(ctx, req.Email); exist {
		return pkgerrors.WrapAppError(pkgerrors.Conflict, message.ErrRegisterFailed, pkgerrors.NewAppError(pkgerrors.Conflict, message.ErrEmailAlreadyExists))
	}

	hashedPassword, err := u.hasher.HashPassword(req.Password)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRegisterFailed, err)
	}

	userData, err := entity.NewUser(req.Name, req.Email, hashedPassword)
	if err != nil || userData == nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrRegisterFailed, err)
	}

	savedUser, err := u.userRepo.Create(ctx, userData)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRegisterFailed, err)
	}

	plainText := fmt.Sprintf("%s_%s", savedUser.ID.String(), time.Now().Add(24*time.Hour).Format(time.RFC3339))
	token, err := u.encryptor.EncryptURLSafe(plainText)
	if err != nil {
		logger.Error("failed to prepare verification email", zap.Error(err))
		return nil
	}

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", u.cfg.AppUrl, token)
	message := entity.UserRegisterEvent{
		UserEvent:       entity.UserEvent{UserID: savedUser.ID, Name: savedUser.Name},
		Email:           savedUser.Email,
		VerificationURL: verificationURL,
	}

	if err := u.outboxUseCase.SaveOutboxMessage(ctx, "user", savedUser.ID.String(), entity.EventUserRegistered, message); err != nil {
		logger.Error("failed to save outbox message for register event", zap.Error(err))
	}

	return nil
}

func (u *userUseCase) RefreshToken(ctx context.Context, hashedRefreshToken string) (*dto.RefreshTokenResponse, error) {
	refreshToken, err := u.encryptor.DecryptInternal(hashedRefreshToken)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Unauthorized, message.ErrRefreshToken, err)
	}

	claims, err := u.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, pkgerrors.AsAppError(pkgerrors.Unauthorized, message.ErrRefreshToken, err)
	}

	if !u.refreshTokenRepo.IsTokenValid(ctx, refreshToken) {
		return nil, pkgerrors.WrapAppError(pkgerrors.Unauthorized, message.ErrRefreshToken, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrTokenInvalid))
	}

	user, err := u.userRepo.FindByID(ctx, claims.UserID.String())
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Unauthorized, message.ErrRefreshToken, err)
	}

	newAccessToken, _, err := u.jwt.GenerateAccessToken(user.ID, user.Role.String())
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRefreshToken, err)
	}

	newRefreshToken, expiryAt, err := u.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRefreshToken, err)
	}

	if err := u.refreshTokenRepo.RevokeByToken(ctx, refreshToken); err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRefreshToken, err)
	}

	newHashedRefreshToken, err := u.encryptor.EncryptInternal(newRefreshToken)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRefreshToken, err)
	}

	tokenData := entity.NewRefreshToken(user.ID, newHashedRefreshToken, expiryAt, false, user)
	if err := u.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRefreshToken, err)
	}

	return &dto.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (u *userUseCase) Logout(ctx context.Context, id string) error {
	if err := u.refreshTokenRepo.RevokeAllByUserID(ctx, id); err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrLogoutFailed, err)
	}
	return nil
}

func (u *userUseCase) SendVerifyEmail(ctx context.Context, req *dto.SendVerifyEmailRequest) error {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrPrepareVerificationEmail, err)
	}

	plainText := fmt.Sprintf("%s_%s", user.ID.String(), time.Now().Add(24*time.Hour).Format(time.RFC3339))
	token, err := u.encryptor.EncryptURLSafe(plainText)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrPrepareVerificationEmail, err)
	}

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", u.cfg.AppUrl, token)
	message := entity.UserRegisterEvent{
		UserEvent:       entity.UserEvent{UserID: user.ID, Name: user.Name},
		Email:           user.Email,
		VerificationURL: verificationURL,
	}

	if err := u.outboxUseCase.SaveOutboxMessage(ctx, "user", user.ID.String(), entity.EventUserRegistered, message); err != nil {
		logger.Error("failed to save outbox message for verify email event", zap.Error(err))
	}

	return nil
}

func (u *userUseCase) VerifyEmail(ctx context.Context, token string) error {
	decrypted, err := u.encryptor.DecryptURLSafe(token)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrVerifyEmailFailed, err)
	}

	payload := strings.Split(decrypted, "_")
	if len(payload) != 2 {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrVerifyEmailFailed, pkgerrors.NewAppError(pkgerrors.BadRequest, message.ErrTokenInvalid))
	}

	expiryTime, err := time.Parse(time.RFC3339, payload[1])
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrVerifyEmailFailed, err)
	}

	if time.Now().After(expiryTime) {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrVerifyEmailFailed, pkgerrors.NewAppError(pkgerrors.BadRequest, message.ErrTokenExpired))
	}

	user, err := u.userRepo.FindByID(ctx, payload[0])
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrVerifyEmailFailed, err)
	}

	user.VerifyEmail()
	if _, err = u.userRepo.Update(ctx, user); err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrVerifyEmailFailed, err)
	}

	return nil
}

func (u *userUseCase) SendResetPassword(ctx context.Context, req *dto.SendResetPasswordRequest) error {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrPrepareForgotPasswordEmail, err)
	}

	plainText := fmt.Sprintf("%s_%s", user.Email, time.Now().Add(1*time.Hour).Format(time.RFC3339))
	token, err := u.encryptor.EncryptURLSafe(plainText)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrPrepareForgotPasswordEmail, err)
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", u.cfg.AppUrl, token)
	message := entity.UserResetPasswordEvent{
		UserEvent: entity.UserEvent{UserID: user.ID, Name: user.Name},
		Email:     user.Email,
		ResetURL:  resetURL,
	}

	if err := u.outboxUseCase.SaveOutboxMessage(ctx, "user", user.ID.String(), entity.EventUserResetPassword, message); err != nil {
		logger.Error("failed to save outbox message for reset password event", zap.Error(err))
	}

	return nil
}

func (u *userUseCase) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	decrypted, err := u.encryptor.DecryptURLSafe(req.Token)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrResetPasswordFailed, err)
	}

	payload := strings.Split(decrypted, "_")
	if len(payload) != 2 {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrResetPasswordFailed, pkgerrors.NewAppError(pkgerrors.BadRequest, message.ErrTokenInvalid))
	}

	expiryTime, err := time.Parse(time.RFC3339, payload[1])
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrResetPasswordFailed, err)
	}

	if time.Now().After(expiryTime) {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrResetPasswordFailed, pkgerrors.NewAppError(pkgerrors.BadRequest, message.ErrTokenExpired))
	}

	user, err := u.userRepo.FindByEmail(ctx, payload[0])
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrResetPasswordFailed, err)
	}

	hashedPassword, err := u.hasher.HashPassword(req.NewPassword)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrResetPasswordFailed, err)
	}

	user.SetPassword(hashedPassword)
	if _, err = u.userRepo.Update(ctx, user); err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrResetPasswordFailed, err)
	}

	return nil
}

func (u *userUseCase) GetAllUsers(ctx context.Context, page, pageSize int, search string) (*dto.PaginationResponse[dto.UserInfo], error) {
	cacheKey := fmt.Sprintf("users:all:page:%d:size:%d:search:%s", page, pageSize, search)

	var cachedResult dto.PaginationResponse[dto.UserInfo]
	if err := u.cache.Get(ctx, cacheKey, &cachedResult); err == nil {
		return &cachedResult, nil
	} else if !errors.Is(err, pkgerrors.ErrCacheMiss) {
		logger.Error("cache error", zap.String("key", cacheKey), zap.Error(err))
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	users, total, err := u.userRepo.FindAll(ctx, limit, offset, search)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrGetAllUsers, err)
	}

	userInfos := make([]dto.UserInfo, len(users))
	for i, user := range users {
		userInfos[i] = *dto.FormatUserInfo(user)
	}

	result := dto.NewPaginationResponse(userInfos, page, pageSize, int(total))

	if err := u.cache.SetWithExpiration(ctx, cacheKey, result, 5*time.Minute); err != nil {
		logger.Error("failed to cache result", zap.Error(err))
	}

	return result, nil
}

func (u *userUseCase) GetUserByID(ctx context.Context, id string) (*dto.UserInfo, error) {
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

func (u *userUseCase) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserInfo, error) {
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

func (u *userUseCase) UpdateUser(ctx context.Context, id string, req *dto.UpdateUserRequest) (*dto.UserInfo, error) {
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

func (u *userUseCase) ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error {
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

func (u *userUseCase) ChangeStatus(ctx context.Context, id string, req *dto.ChangeUserStatusRequest) error {
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

func (u *userUseCase) DeleteUser(ctx context.Context, id string) error {
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
