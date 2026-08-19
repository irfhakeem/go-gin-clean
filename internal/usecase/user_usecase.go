package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-gin-clean/internal/dto"
	"go-gin-clean/internal/dto/event"
	"go-gin-clean/internal/entity"
	"go-gin-clean/internal/gateway/cache"
	"go-gin-clean/internal/gateway/media"
	"go-gin-clean/internal/gateway/messaging"
	"go-gin-clean/internal/gateway/security"
	"go-gin-clean/internal/repository"
	"go-gin-clean/pkg/config"
	pkgerror "go-gin-clean/pkg/error"
	"go-gin-clean/pkg/logger"
	"go-gin-clean/pkg/utils"

	"go.uber.org/zap"
)

type UserUseCaseInterface interface {
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	Register(ctx context.Context, req *dto.RegisterRequest) error
	RefreshToken(ctx context.Context, hashedRefreshToken string) (*dto.RefreshTokenResponse, error)
	Logout(ctx context.Context, id string) error
	SendVerifyEmail(ctx context.Context, req dto.SendVerifyEmailRequest) error
	VerifyEmail(ctx context.Context, token string) error
	SendResetPassword(ctx context.Context, req dto.SendResetPasswordRequest) error
	ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error
	GetOAuthLoginURL(ctx context.Context, provider, appID, platform string) (*dto.OAuthUrlResponse, error)
	GetOAuthRedirectURL(appID, platform string) string
	HandleOAuthCallback(ctx context.Context, req *dto.OAuthCallbackRequest) (*dto.LoginResponse, string, string, error)
	GetAllUsers(ctx context.Context, page, pageSize int, search string) (*dto.PaginationResponse[dto.UserInfo], error)
	GetUserByID(ctx context.Context, id string) (*dto.UserInfo, error)
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserInfo, error)
	UpdateUser(ctx context.Context, id string, req *dto.UpdateUserRequest) (*dto.UserInfo, error)
	ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error
	ChangeStatus(ctx context.Context, id string, req dto.ChangeUserStatusRequest) error
	DeleteUser(ctx context.Context, id string) error
}

type UserUseCase struct {
	userRepo         repository.UserRepositoryInterface
	refreshTokenRepo repository.RefreshTokenRepositoryInterface
	outboxUseCase    OutboxUseCaseInterface

	jwtService    security.JWTServiceInterface
	bcryptService security.HasherServiceInterface
	oauthService  security.OAuthServiceInterface
	aesService    security.EncryptionServiceInterface
	mediaService  media.StorageServiceInterface
	redisService  cache.CacheServiceInterface

	cfg *config.ServerConfig
}

func NewUserUseCase(
	userRepo repository.UserRepositoryInterface,
	refreshTokenRepo repository.RefreshTokenRepositoryInterface,
	outboxUseCase OutboxUseCaseInterface,
	jwtService security.JWTServiceInterface,
	bcryptService security.HasherServiceInterface,
	oauthService security.OAuthServiceInterface,
	aesService security.EncryptionServiceInterface,
	mediaService media.StorageServiceInterface,
	redisService cache.CacheServiceInterface,
	cfg *config.ServerConfig,
) UserUseCaseInterface {
	return &UserUseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		outboxUseCase:    outboxUseCase,
		jwtService:       jwtService,
		bcryptService:    bcryptService,
		oauthService:     oauthService,
		aesService:       aesService,
		mediaService:     mediaService,
		redisService:     redisService,
		cfg:              cfg,
	}
}

func (u *UserUseCase) GetOAuthRedirectURL(appID, platform string) string {
	if platform == "mobile" {
		return u.oauthService.GetMobileDeepLinkURL(appID)
	}
	return u.oauthService.GetFrontendURL(appID)
}

func (u *UserUseCase) GetOAuthLoginURL(ctx context.Context, provider, appID, platform string) (*dto.OAuthUrlResponse, error) {
	switch provider {
	case "google":
		return &dto.OAuthUrlResponse{
			AuthURL: u.oauthService.GetGoogleAuthURL(appID, platform),
		}, nil
	default:
		return nil, pkgerror.InternalServerError(pkgerror.ErrOAuthLoginFailed, pkgerror.ErrInvalidOAuthProvider)
	}
}

func (u *UserUseCase) HandleOAuthCallback(ctx context.Context, req *dto.OAuthCallbackRequest) (*dto.LoginResponse, string, string, error) {
	var (
		user     *entity.User
		appID    string
		platform string
		err      error
	)

	switch req.Provider {
	case "google":
		user, appID, platform, err = u.oauthService.HandleGoogleCallback(ctx, req.State, req.Code)
		if err != nil {
			return nil, appID, platform, pkgerror.Unauthorized(pkgerror.ErrOAuthCallback, err)
		}
	default:
		return nil, appID, platform, pkgerror.Unauthorized(pkgerror.ErrInvalidOAuthProvider, nil)
	}

	existingUser, err := u.userRepo.FindByOAuthID(ctx, req.Provider, user.OAuthID)
	if err == nil {
		user = existingUser
	} else {
		existingUserByEmail, err := u.userRepo.FindByEmail(ctx, user.Email)
		if err == nil {
			if err = u.userRepo.UpdateOAuthInfo(ctx, existingUserByEmail.ID.String(), req.Provider, user.OAuthID); err != nil {
				return nil, appID, platform, pkgerror.Unauthorized(pkgerror.ErrLinkOAuth, err)
			}
			user = existingUserByEmail
		} else {
			user, err = u.userRepo.Create(ctx, user)
			if err != nil {
				return nil, appID, platform, pkgerror.Unauthorized(pkgerror.ErrOAuthSignUp, err)
			}
		}
	}

	accessToken, _, err := u.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, appID, platform, pkgerror.Unauthorized(pkgerror.ErrAccessToken, err)
	}

	refreshToken, expiryAt, err := u.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, appID, platform, pkgerror.Unauthorized(pkgerror.ErrRefreshToken, err)
	}

	hashedRefreshToken, err := u.aesService.EncryptInternal(refreshToken)
	if err != nil {
		return nil, appID, platform, pkgerror.Unauthorized(pkgerror.ErrProcessToken, err)
	}

	tokenData := entity.NewRefreshToken(user.ID, hashedRefreshToken, expiryAt, false, *user)
	if err := u.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return nil, appID, platform, pkgerror.Unauthorized(pkgerror.ErrRefreshToken, err)
	}

	return dto.FormatLoginResponse(accessToken, hashedRefreshToken), appID, platform, nil
}

func (u *UserUseCase) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, pkgerror.Unauthorized(pkgerror.ErrLoginFailed, err)
	}

	if user.IsOAuthUser() {
		return nil, pkgerror.Unauthorized(pkgerror.ErrLoginFailed, pkgerror.ErrOAuthUserUseOAuthLogin)
	}

	if !user.IsActive {
		return nil, pkgerror.Unauthorized(pkgerror.ErrLoginFailed, pkgerror.ErrUserNotFound)
	}

	if !user.IsVerified {
		return nil, pkgerror.Unauthorized(pkgerror.ErrLoginFailed, pkgerror.ErrEmailNotVerified)
	}

	if err := u.bcryptService.ValidatePassword(req.Password, user.Password); err != nil {
		return nil, pkgerror.Unauthorized(pkgerror.ErrLoginFailed, pkgerror.ErrPasswordNotMatch)
	}

	accessToken, _, err := u.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, pkgerror.Unauthorized(pkgerror.ErrLoginFailed, err)
	}

	refreshToken, expiryAt, err := u.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, pkgerror.Unauthorized(pkgerror.ErrLoginFailed, err)
	}

	hashedRefreshToken, err := u.aesService.EncryptInternal(refreshToken)
	if err != nil {
		return nil, pkgerror.Unauthorized(pkgerror.ErrLoginFailed, err)
	}

	tokenData := entity.NewRefreshToken(user.ID, hashedRefreshToken, expiryAt, false, *user)
	if err := u.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return nil, pkgerror.Unauthorized(pkgerror.ErrLoginFailed, err)
	}

	return dto.FormatLoginResponse(accessToken, hashedRefreshToken), nil
}

func (u *UserUseCase) Register(ctx context.Context, req *dto.RegisterRequest) error {
	if exist := u.userRepo.ExistByEmail(ctx, req.Email); exist {
		return pkgerror.BadRequest(pkgerror.ErrRegisterFailed, pkgerror.ErrEmailAlreadyExists)
	}

	hashedPassword, err := u.bcryptService.HashPassword(req.Password)
	if err != nil {
		return pkgerror.BadRequest(pkgerror.ErrRegisterFailed, err)
	}

	userData, err := entity.NewUser(req.Name, req.Email, hashedPassword, "", entity.Other)
	if err != nil || userData == nil {
		return pkgerror.BadRequest(pkgerror.ErrRegisterFailed, err)
	}

	savedUser, err := u.userRepo.Create(ctx, userData)
	if err != nil {
		return pkgerror.BadRequest(pkgerror.ErrRegisterFailed, err)
	}

	plainText := fmt.Sprintf("%s_%s", savedUser.ID.String(), time.Now().Add(24*time.Hour).Format(time.RFC3339))
	token, err := u.aesService.EncryptURLSafe(plainText)
	if err != nil {
		logger.Error("failed to prepare verification email", zap.Error(err))
		return nil
	}

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", u.cfg.AppUrl, token)
	message := event.RegisterEvent{
		UserEvent:       event.UserEvent{UserID: savedUser.ID, Name: savedUser.Name},
		Email:           savedUser.Email,
		VerificationURL: verificationURL,
	}

	if err := u.outboxUseCase.SaveOutboxMessage(ctx, "user", savedUser.ID.String(), messaging.UserRegisteredEvent, message); err != nil {
		logger.Error("failed to save outbox message for register event", zap.Error(err))
	}

	return nil
}

func (u *UserUseCase) RefreshToken(ctx context.Context, hashedRefreshToken string) (*dto.RefreshTokenResponse, error) {
	refreshToken, err := u.aesService.DecryptInternal(hashedRefreshToken)
	if err != nil {
		return nil, pkgerror.Unauthorized(pkgerror.ErrRefreshToken, err)
	}

	claims, err := u.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, pkgerror.Unauthorized(pkgerror.ErrRefreshToken, err)
	}

	if !u.refreshTokenRepo.IsTokenValid(ctx, refreshToken) {
		return nil, pkgerror.Unauthorized(pkgerror.ErrRefreshToken, pkgerror.ErrTokenInvalid)
	}

	user, err := u.userRepo.FindByID(ctx, claims.UserID.String())
	if err != nil {
		return nil, pkgerror.Unauthorized(pkgerror.ErrRefreshToken, err)
	}

	newAccessToken, _, err := u.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, pkgerror.Unauthorized(pkgerror.ErrRefreshToken, err)
	}

	newRefreshToken, expiryAt, err := u.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, pkgerror.Unauthorized(pkgerror.ErrRefreshToken, err)
	}

	if err := u.refreshTokenRepo.RevokeByToken(ctx, refreshToken); err != nil {
		return nil, pkgerror.Unauthorized(pkgerror.ErrRefreshToken, err)
	}

	newHashedRefreshToken, err := u.aesService.EncryptInternal(newRefreshToken)
	if err != nil {
		return nil, pkgerror.Unauthorized(pkgerror.ErrRefreshToken, err)
	}

	tokenData := entity.NewRefreshToken(user.ID, newHashedRefreshToken, expiryAt, false, *user)
	if err := u.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return nil, pkgerror.Unauthorized(pkgerror.ErrRefreshToken, err)
	}

	return &dto.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (u *UserUseCase) Logout(ctx context.Context, id string) error {
	if err := u.refreshTokenRepo.RevokeAllByUserID(ctx, id); err != nil {
		return pkgerror.InternalServerError(pkgerror.ErrLogoutFailed, err)
	}
	return nil
}

func (u *UserUseCase) SendVerifyEmail(ctx context.Context, req dto.SendVerifyEmailRequest) error {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return pkgerror.InternalServerError(pkgerror.ErrPrepareVerificationEmail, err)
	}

	plainText := fmt.Sprintf("%s_%s", user.ID.String(), time.Now().Add(24*time.Hour).Format(time.RFC3339))
	token, err := u.aesService.EncryptURLSafe(plainText)
	if err != nil {
		return pkgerror.InternalServerError(pkgerror.ErrPrepareVerificationEmail, err)
	}

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", u.cfg.AppUrl, token)
	message := event.RegisterEvent{
		UserEvent:       event.UserEvent{UserID: user.ID, Name: user.Name},
		Email:           user.Email,
		VerificationURL: verificationURL,
	}

	if err := u.outboxUseCase.SaveOutboxMessage(ctx, "user", user.ID.String(), messaging.UserRegisteredEvent, message); err != nil {
		logger.Error("failed to save outbox message for verify email event", zap.Error(err))
	}

	return nil
}

func (u *UserUseCase) VerifyEmail(ctx context.Context, token string) error {
	decrypted, err := u.aesService.DecryptURLSafe(token)
	if err != nil {
		return pkgerror.BadRequest(pkgerror.ErrVerifyEmailFailed, err)
	}

	payload := strings.Split(decrypted, "_")
	if len(payload) != 2 {
		return pkgerror.BadRequest(pkgerror.ErrVerifyEmailFailed, pkgerror.ErrTokenInvalid)
	}

	expiryTime, err := time.Parse(time.RFC3339, payload[1])
	if err != nil {
		return pkgerror.BadRequest(pkgerror.ErrVerifyEmailFailed, err)
	}

	if time.Now().After(expiryTime) {
		return pkgerror.BadRequest(pkgerror.ErrVerifyEmailFailed, pkgerror.ErrTokenExpired)
	}

	user, err := u.userRepo.FindByID(ctx, payload[0])
	if err != nil {
		return pkgerror.BadRequest(pkgerror.ErrVerifyEmailFailed, err)
	}

	user.VerifyEmail()
	if _, err = u.userRepo.Update(ctx, user); err != nil {
		return pkgerror.BadRequest(pkgerror.ErrVerifyEmailFailed, err)
	}

	return nil
}

func (u *UserUseCase) SendResetPassword(ctx context.Context, req dto.SendResetPasswordRequest) error {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return pkgerror.InternalServerError(pkgerror.ErrPrepareForgotPasswordEmail, err)
	}

	plainText := fmt.Sprintf("%s_%s", user.Email, time.Now().Add(1*time.Hour).Format(time.RFC3339))
	token, err := u.aesService.EncryptURLSafe(plainText)
	if err != nil {
		return pkgerror.InternalServerError(pkgerror.ErrPrepareForgotPasswordEmail, err)
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", u.cfg.AppUrl, token)
	message := event.ResetPasswordEvent{
		UserEvent: event.UserEvent{UserID: user.ID, Name: user.Name},
		Email:     user.Email,
		ResetURL:  resetURL,
	}

	if err := u.outboxUseCase.SaveOutboxMessage(ctx, "user", user.ID.String(), messaging.UserResetPasswordEvent, message); err != nil {
		logger.Error("failed to save outbox message for reset password event", zap.Error(err))
	}

	return nil
}

func (u *UserUseCase) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	decrypted, err := u.aesService.DecryptURLSafe(req.Token)
	if err != nil {
		return pkgerror.BadRequest(pkgerror.ErrResetPasswordFailed, err)
	}

	payload := strings.Split(decrypted, "_")
	if len(payload) != 2 {
		return pkgerror.BadRequest(pkgerror.ErrResetPasswordFailed, pkgerror.ErrTokenInvalid)
	}

	expiryTime, err := time.Parse(time.RFC3339, payload[1])
	if err != nil {
		return pkgerror.BadRequest(pkgerror.ErrResetPasswordFailed, err)
	}

	if time.Now().After(expiryTime) {
		return pkgerror.BadRequest(pkgerror.ErrResetPasswordFailed, pkgerror.ErrTokenExpired)
	}

	user, err := u.userRepo.FindByEmail(ctx, payload[0])
	if err != nil {
		return pkgerror.BadRequest(pkgerror.ErrResetPasswordFailed, err)
	}

	hashedPassword, err := u.bcryptService.HashPassword(req.NewPassword)
	if err != nil {
		return pkgerror.BadRequest(pkgerror.ErrResetPasswordFailed, err)
	}

	user.SetPassword(hashedPassword)
	if _, err = u.userRepo.Update(ctx, user); err != nil {
		return pkgerror.BadRequest(pkgerror.ErrResetPasswordFailed, err)
	}

	return nil
}

func (u *UserUseCase) GetAllUsers(ctx context.Context, page, pageSize int, search string) (*dto.PaginationResponse[dto.UserInfo], error) {
	cacheKey := fmt.Sprintf("users:all:page:%d:size:%d:search:%s", page, pageSize, search)

	var cachedResult dto.PaginationResponse[dto.UserInfo]
	if err := u.redisService.Get(ctx, cacheKey, &cachedResult); err == nil {
		return &cachedResult, nil
	} else if !errors.Is(err, pkgerror.ErrCacheMiss) {
		logger.Error("cache error", zap.String("key", cacheKey), zap.Error(err))
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	users, total, err := u.userRepo.FindAll(ctx, limit, offset, search)
	if err != nil {
		return nil, pkgerror.InternalServerError(pkgerror.ErrGetAllUsers, err)
	}

	userInfos := make([]dto.UserInfo, len(users))
	for i, user := range users {
		userInfos[i] = *dto.FormatUserInfo(user)
	}

	result := dto.NewPaginationResponse(userInfos, page, pageSize, int(total))

	if err := u.redisService.SetWithExpiration(ctx, cacheKey, result, 5*time.Minute); err != nil {
		logger.Error("failed to cache result", zap.Error(err))
	}

	return result, nil
}

func (u *UserUseCase) GetUserByID(ctx context.Context, id string) (*dto.UserInfo, error) {
	cacheKey := fmt.Sprintf("user:id:%s", id)

	var cachedUser dto.UserInfo
	if err := u.redisService.Get(ctx, cacheKey, &cachedUser); err == nil {
		return &cachedUser, nil
	} else if !errors.Is(err, pkgerror.ErrCacheMiss) {
		logger.Error("cache error", zap.String("key", cacheKey), zap.Error(err))
	}

	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil || user == nil {
		return nil, pkgerror.InternalServerError(pkgerror.ErrGetUserInformation, err)
	}
	if !user.IsActive || !user.IsVerified {
		return nil, pkgerror.NotFound(pkgerror.ErrUserNotFound, nil)
	}

	userInfo := dto.FormatUserInfo(user)

	if err := u.redisService.SetWithExpiration(ctx, cacheKey, userInfo, 5*time.Minute); err != nil {
		logger.Error("failed to cache result", zap.Error(err))
	}

	return userInfo, nil
}

func (u *UserUseCase) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserInfo, error) {
	if u.userRepo.ExistByEmail(ctx, req.Email) {
		return nil, pkgerror.BadRequest(pkgerror.ErrCreateUser, pkgerror.ErrEmailAlreadyExists)
	}

	hashedPassword, err := u.bcryptService.HashPassword(req.Password)
	if err != nil {
		return nil, pkgerror.BadRequest(pkgerror.ErrCreateUser, err)
	}

	userData, err := entity.NewUser(req.Name, req.Email, hashedPassword, "", entity.Other)
	if err != nil {
		return nil, pkgerror.BadRequest(pkgerror.ErrCreateUser, err)
	}

	savedUser, err := u.userRepo.Create(ctx, userData)
	if err != nil {
		return nil, pkgerror.BadRequest(pkgerror.ErrCreateUser, err)
	}

	if err := u.redisService.DeletePattern(ctx, "users:all:*"); err != nil {
		logger.Error("failed to invalidate users cache", zap.Error(err))
	}

	return dto.FormatUserInfo(savedUser), nil
}

func (u *UserUseCase) UpdateUser(ctx context.Context, id string, req *dto.UpdateUserRequest) (*dto.UserInfo, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, pkgerror.BadRequest(pkgerror.ErrUpdateUser, err)
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Avatar != nil {
		allowedExtensions := []string{".jpg", ".jpeg", ".png"}
		if !utils.IsValidExtension(req.Avatar.Filename, allowedExtensions) {
			return nil, pkgerror.BadRequest(pkgerror.ErrUpdateUser, pkgerror.ErrUnsupportedImageType)
		}

		path, err := u.mediaService.UploadFile(
			ctx,
			fmt.Sprintf("avatar_%s_%d.jpg", user.ID.String(), time.Now().Unix()),
			req.Avatar.Size,
			*req.Avatar,
			"users/"+user.ID.String()+"/avatar/",
		)
		if err != nil || path == nil {
			return nil, pkgerror.BadRequest(pkgerror.ErrUpdateUser, err)
		}

		user.Avatar = *path
	}

	if req.Gender != nil {
		user.Gender = *req.Gender
	}

	updatedUser, err := u.userRepo.Update(ctx, user)
	if err != nil {
		return nil, pkgerror.BadRequest(pkgerror.ErrUpdateUser, err)
	}

	cacheKey := fmt.Sprintf("user:id:%s", id)
	if err := u.redisService.Delete(ctx, cacheKey); err != nil {
		logger.Error("failed to invalidate user cache", zap.Error(err))
	}
	if err := u.redisService.DeletePattern(ctx, "users:all:*"); err != nil {
		logger.Error("failed to invalidate users cache", zap.Error(err))
	}

	return dto.FormatUserInfo(updatedUser), nil
}

func (u *UserUseCase) ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return pkgerror.BadRequest(pkgerror.ErrUpdateUserPassword, err)
	}

	if err := u.bcryptService.ValidatePassword(req.OldPassword, user.Password); err != nil {
		return pkgerror.BadRequest(pkgerror.ErrUpdateUserPassword, pkgerror.ErrPasswordNotMatch)
	}

	hashedPassword, err := u.bcryptService.HashPassword(req.NewPassword)
	if err != nil {
		return pkgerror.BadRequest(pkgerror.ErrUpdateUserPassword, err)
	}

	user.SetPassword(hashedPassword)
	if _, err = u.userRepo.Update(ctx, user); err != nil {
		return pkgerror.BadRequest(pkgerror.ErrUpdateUserPassword, err)
	}

	return nil
}

func (u *UserUseCase) ChangeStatus(ctx context.Context, id string, req dto.ChangeUserStatusRequest) error {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return pkgerror.InternalServerError(pkgerror.ErrUpdateUserStatus, err)
	}

	user.IsActive = *req.IsActive
	if _, err = u.userRepo.Update(ctx, user); err != nil {
		return pkgerror.InternalServerError(pkgerror.ErrUpdateUserStatus, err)
	}

	cacheKey := fmt.Sprintf("user:id:%s", id)
	if err := u.redisService.Delete(ctx, cacheKey); err != nil {
		logger.Error("failed to invalidate user cache", zap.Error(err))
	}
	if err := u.redisService.DeletePattern(ctx, "users:all:*"); err != nil {
		logger.Error("failed to invalidate users cache", zap.Error(err))
	}

	return nil
}

func (u *UserUseCase) DeleteUser(ctx context.Context, id string) error {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return pkgerror.InternalServerError(pkgerror.ErrDeleteUser, err)
	}

	if err = u.refreshTokenRepo.RevokeAllByUserID(ctx, user.ID.String()); err != nil {
		return pkgerror.InternalServerError(pkgerror.ErrDeleteUser, err)
	}

	if err = u.userRepo.Delete(ctx, user.ID.String()); err != nil {
		return pkgerror.InternalServerError(pkgerror.ErrDeleteUser, err)
	}

	cacheKey := fmt.Sprintf("user:id:%s", id)
	if err := u.redisService.Delete(ctx, cacheKey); err != nil {
		logger.Error("failed to invalidate user cache", zap.Error(err))
	}
	if err := u.redisService.DeletePattern(ctx, "users:all:*"); err != nil {
		logger.Error("failed to invalidate users cache", zap.Error(err))
	}

	return nil
}
