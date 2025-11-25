package usecase

import (
	"context"
	"fmt"
	"log"

	"go-gin-clean/internal/entity"
	"go-gin-clean/internal/gateway/cache"
	"go-gin-clean/internal/gateway/media"
	"go-gin-clean/internal/gateway/messaging"
	"go-gin-clean/internal/gateway/security"
	"go-gin-clean/internal/model"
	"go-gin-clean/internal/repository"
	"go-gin-clean/pkg/config"
	"go-gin-clean/pkg/errors"
	"strings"
	"time"
)

type UserUseCase struct {
	userRepo         *repository.UserRepository
	refreshTokenRepo *repository.RefreshTokenRepository

	jwtService          *security.JWTService
	bcryptService       *security.BcryptService
	oauthService        *security.OAuthService
	aesService          *security.AESService
	cloudinaryService   *media.CloudinaryService
	localStorageService *media.LocalStorageService
	redisService        *cache.RedisService

	UserPublisher *messaging.UserPublisher
}

func NewUserUseCase(
	userRepo *repository.UserRepository,
	refreshTokenRepo *repository.RefreshTokenRepository,
	jwtService *security.JWTService,
	bcryptService *security.BcryptService,
	oauthService *security.OAuthService,
	aesService *security.AESService,
	cloudinaryService *media.CloudinaryService,
	localStorageService *media.LocalStorageService,
	redisService *cache.RedisService,

	UserPublisher *messaging.UserPublisher,
) *UserUseCase {
	return &UserUseCase{
		userRepo:          userRepo,
		refreshTokenRepo:  refreshTokenRepo,
		jwtService:        jwtService,
		bcryptService:     bcryptService,
		oauthService:      oauthService,
		aesService:        aesService,
		cloudinaryService: cloudinaryService,
		redisService:      redisService,
		UserPublisher:     UserPublisher,
	}
}

func formatUserInfo(user *entity.User) *model.UserInfo {
	return &model.UserInfo{
		PKID:     user.PKID,
		Name:     user.Name,
		Code:     user.Code,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Gender:   user.Gender,
		IsActive: user.IsActive,
	}
}

func (u *UserUseCase) GetOAuthLoginURL(ctx context.Context, provider string, appID string) (*model.OAuthUrlResponse, error) {
	var authURL string

	switch provider {
	case "google":
		authURL = u.oauthService.GetGoogleAuthURL(appID)
	default:
		return nil, errors.ErrInvalidOAuthProvider
	}

	return &model.OAuthUrlResponse{
		AuthURL: authURL,
	}, nil
}

func (u *UserUseCase) HandleOAuthCallback(ctx context.Context, req *model.OAuthCallbackRequest) (*model.LoginResponse, string, error) {
	var user *entity.User
	var err error
	var appID string

	switch req.Provider {
	case "google":
		user, appID, err = u.oauthService.HandleGoogleCallback(ctx, req.State, req.Code)
	default:
		return nil, appID, errors.ErrInvalidOAuthProvider
	}

	if err != nil {
		return nil, appID, fmt.Errorf("OAuth callback error: %w", err)
	}

	existingUser, err := u.userRepo.FindByOAuthID(ctx, req.Provider, user.OAuthID)
	if err == nil {
		user = existingUser
	} else {
		existingUserByEmail, err := u.userRepo.FindByEmail(ctx, user.Email)
		if err == nil {
			err = u.userRepo.UpdateOAuthInfo(
				ctx,
				existingUserByEmail.PKID,
				req.Provider,
				user.OAuthID,
			)
			if err != nil {
				return nil, "", fmt.Errorf("failed to link OAuth to existing account: %w", err)
			}

			user = existingUserByEmail
		} else {
			user, err = u.userRepo.Create(ctx, user)
			if err != nil {
				return nil, "", fmt.Errorf("failed to create user: %w", err)
			}
		}
	}

	accessToken, _, err := u.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, "", err
	}

	refreshToken, expiryAt, err := u.jwtService.GenerateRefreshToken(user.PKID)
	if err != nil {
		return nil, "", err
	}

	hashedRefreshToken, err := u.aesService.EncryptInternal(refreshToken)
	if err != nil {
		return nil, "", err
	}

	tokenData := entity.NewRefreshToken(user.PKID, hashedRefreshToken, expiryAt, false, *user)

	if err := u.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return nil, "", err
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: hashedRefreshToken,
	}, appID, nil
}

func (u *UserUseCase) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if user.OAuthProvider != "" {
		return nil, errors.ErrOAuthUserUseOAuthLogin
	}

	if !user.IsActive {
		return nil, errors.ErrUserNotFound
	}

	if !user.IsVerified {
		return nil, errors.ErrEmailNotVerified
	}

	if err := u.bcryptService.ValidatePassword(req.Password, user.Password); err != nil {
		return nil, errors.ErrPasswordNotMatch
	}

	accessToken, _, err := u.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, expiryAt, err := u.jwtService.GenerateRefreshToken(user.PKID)
	if err != nil {
		return nil, err
	}

	hashedRefreshToken, err := u.aesService.EncryptInternal(refreshToken)
	if err != nil {
		return nil, err
	}

	tokenData := entity.NewRefreshToken(user.PKID, hashedRefreshToken, expiryAt, false, *user)

	if err := u.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: hashedRefreshToken,
	}, nil
}

func (u *UserUseCase) Register(ctx context.Context, req *model.RegisterRequest) error {
	if exist := u.userRepo.ExistByEmail(ctx, req.Email); exist {
		return errors.ErrEmailAlreadyExists
	}

	hashedPassword, err := u.bcryptService.HashPassword(req.Password)
	if err != nil {
		return err
	}

	userData, err := entity.NewUser(req.Name, req.Email, hashedPassword, "", entity.Other)
	if err != nil {
		return err
	}

	if userData == nil {
		return errors.ErrInvalidInput
	}

	savedUser, err := u.userRepo.Create(ctx, userData)
	if err != nil {
		return err
	}

	go func() {
		plainText := fmt.Sprintf("%s_%s", savedUser.Code, time.Now().Add(24*time.Hour).Format(time.RFC3339))

		token, err := u.aesService.EncryptURLSafe(plainText)
		if err != nil {
			log.Println("Failed to prepare email", err)
		}

		verificationURL := fmt.Sprintf("%s/verify-email?token=%s", config.GetAppURL(), token)

		message := model.RegisterEvent{
			UserEvent: model.UserEvent{
				UserPKID: savedUser.PKID,
				Name:     savedUser.Name,
			},
			Email:           savedUser.Email,
			VerificationURL: verificationURL,
		}

		if err := u.UserPublisher.RegisterEventPublish(message); err != nil {
			log.Println("Failed to publish user register event:", err)
		}
	}()

	return nil
}

func (u *UserUseCase) RefreshToken(ctx context.Context, hashedRefreshToken string) (*model.RefreshTokenResponse, error) {
	refreshToken, err := u.aesService.DecryptInternal(hashedRefreshToken)
	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	claims, err := u.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	if !u.refreshTokenRepo.IsTokenValid(ctx, refreshToken) {
		return nil, errors.ErrTokenInvalid
	}

	user, err := u.userRepo.FindByID(ctx, claims.UserPKID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	newAccessToken, _, err := u.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, expiryAt, err := u.jwtService.GenerateRefreshToken(user.PKID)
	if err != nil {
		return nil, err
	}

	if err := u.refreshTokenRepo.RevokeByToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	newHashedRefreshToken, err := u.aesService.EncryptInternal(newRefreshToken)
	if err != nil {
		return nil, err
	}

	tokenData := entity.NewRefreshToken(user.PKID, newHashedRefreshToken, expiryAt, false, *user)

	if err := u.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return nil, err
	}

	return &model.RefreshTokenResponse{
		AccessToken: newAccessToken,
	}, nil
}

func (u *UserUseCase) Logout(ctx context.Context, pkid int64) error {
	return u.refreshTokenRepo.RevokeAllByUserID(ctx, pkid)
}

func (u *UserUseCase) SendVerifyEmail(ctx context.Context, req model.SendVerifyEmailRequest) error {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return errors.ErrUserNotFound
	}

	plainText := fmt.Sprintf("%s_%s", user.Code, time.Now().Add(24*time.Hour).Format(time.RFC3339))

	token, err := u.aesService.EncryptURLSafe(plainText)
	if err != nil {
		return err
	}

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", config.GetAppURL(), token)

	message := model.RegisterEvent{
		UserEvent: model.UserEvent{
			UserPKID: user.PKID,
			Name:     user.Name,
		},
		Email:           user.Email,
		VerificationURL: verificationURL,
	}

	go func() {
		if err := u.UserPublisher.RegisterEventPublish(message); err != nil {
			log.Println("Failed to publish user register event:", err)
		}
	}()

	return nil
}

func (u *UserUseCase) VerifyEmail(ctx context.Context, token string) error {
	token, err := u.aesService.DecryptURLSafe(token)
	if err != nil {
		return errors.ErrTokenInvalid
	}

	var code string
	var expiry string

	payload := strings.Split(token, "_")
	if len(payload) != 2 {
		return errors.ErrTokenInvalid
	}

	code = payload[0]
	expiry = payload[1]

	expiryTime, err := time.Parse(time.RFC3339, expiry)
	if err != nil {
		return errors.ErrTokenInvalid
	}

	if time.Now().After(expiryTime) {
		return errors.ErrTokenExpired
	}

	user, err := u.userRepo.FindByCode(ctx, code)
	if err != nil {
		return errors.ErrUserNotFound
	}

	user.VerifyEmail()

	_, err = u.userRepo.Update(ctx, user, user.Code)
	return err
}

func (u *UserUseCase) SendResetPassword(ctx context.Context, req model.SendResetPasswordRequest) error {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return errors.ErrUserNotFound
	}

	plainText := fmt.Sprintf("%s_%s", user.Email, time.Now().Add(1*time.Hour).Format(time.RFC3339))

	token, err := u.aesService.EncryptURLSafe(plainText)
	if err != nil {
		return err
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", config.GetAppURL(), token)

	message := model.ResetPasswordEvent{
		UserEvent: model.UserEvent{
			UserPKID: user.PKID,
			Name:     user.Name,
		},
		Email:    user.Email,
		ResetURL: resetURL,
	}

	go func() {
		if err := u.UserPublisher.ResetPasswordEventPublish(message); err != nil {
			log.Println("Failed to publish user reset password event:", err)
		}
	}()

	return nil
}

func (u *UserUseCase) ResetPassword(ctx context.Context, req *model.ResetPasswordRequest) error {
	token, err := u.aesService.DecryptURLSafe(req.Token)
	if err != nil {
		return errors.ErrTokenInvalid
	}

	var email string
	var expiry string
	payload := strings.Split(token, "_")

	if len(payload) != 2 {
		return errors.ErrTokenInvalid
	}

	email = payload[0]
	expiry = payload[1]

	expiryTime, err := time.Parse(time.RFC3339, expiry)
	if err != nil {
		return errors.ErrTokenInvalid
	}

	if time.Now().After(expiryTime) {
		return errors.ErrTokenExpired
	}

	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return errors.ErrUserNotFound
	}

	hashedPassword, err := u.bcryptService.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.SetPassword(hashedPassword)

	_, err = u.userRepo.Update(ctx, user, user.Code)
	return err
}

func (u *UserUseCase) GetAllUsers(ctx context.Context, page, pageSize int, search string) (*model.PaginationResponse[model.UserInfo], error) {
	offset := model.Offset(page, pageSize)
	users, total, err := u.userRepo.FindAll(ctx, pageSize, offset, search)
	if err != nil {
		return nil, err
	}

	userInfos := make([]model.UserInfo, len(users))
	for i, user := range users {
		userInfos[i] = *formatUserInfo(user)
	}

	return model.NewPaginationResponse(userInfos, page, pageSize, int(total)), nil
}

func (u *UserUseCase) GetUserByCode(ctx context.Context, code string) (*model.UserInfo, error) {
	user, err := u.userRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	return formatUserInfo(user), nil
}

func (u *UserUseCase) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.UserInfo, error) {
	if u.userRepo.ExistByEmail(ctx, req.Email) {
		return nil, errors.ErrEmailAlreadyExists
	}

	hashedPassword, err := u.bcryptService.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	userData, err := entity.NewUser(req.Name, req.Email, hashedPassword, "", entity.Other)
	if err != nil {
		return nil, err
	}

	savedUser, err := u.userRepo.Create(ctx, userData)
	if err != nil {
		return nil, err
	}

	return formatUserInfo(savedUser), nil
}

func (u *UserUseCase) UpdateUser(ctx context.Context, code string, req *model.UpdateUserRequest) (*model.UserInfo, error) {
	user, err := u.userRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Avatar != nil {
		// Choose one of the storage services to upload the avatar
		// path, err := u.cloudinaryService.UploadFile(ctx, req.Avatar.Filename, req.Avatar.Size, *req.Avatar, "users/"+user.Code+"/avatar/")
		// if err != nil || path == nil {
		// 	return nil, errors.ErrUploadImage
		// }

		path, err := u.localStorageService.UploadFile(
			ctx,
			fmt.Sprintf("avatar_%d_%d.jpg", user.PKID, time.Now().Unix()),
			req.Avatar.Size,
			*req.Avatar,
			"users/"+user.Code+"/avatar/",
		)
		if err != nil || path == nil {
			return nil, errors.ErrUploadImage
		}

		user.Avatar = *path
	}

	if req.Gender != nil {
		user.Gender = *req.Gender
	}

	updatedUser, err := u.userRepo.Update(ctx, user, user.Code)
	if err != nil {
		return nil, err
	}

	return formatUserInfo(updatedUser), nil
}

func (u *UserUseCase) ChangePassword(ctx context.Context, userPKID int64, req *model.ChangePasswordRequest) error {
	user, err := u.userRepo.FindByID(ctx, userPKID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	if err := u.bcryptService.ValidatePassword(req.OldPassword, user.Password); err != nil {
		return errors.ErrPasswordNotMatch
	}

	hashedPassword, err := u.bcryptService.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.SetPassword(hashedPassword)

	_, err = u.userRepo.Update(ctx, user, user.Code)
	return err
}

func (u *UserUseCase) ChangeStatus(ctx context.Context, code string, req model.ChangeUserStatusRequest) error {
	user, err := u.userRepo.FindByCode(ctx, code)
	if err != nil {
		return errors.ErrUserNotFound
	}

	user.IsActive = req.IsActive

	_, err = u.userRepo.Update(ctx, user, user.Code)
	return err
}

func (u *UserUseCase) DeleteUser(ctx context.Context, code string) error {
	user, err := u.userRepo.FindByCode(ctx, code)
	if err != nil {
		return errors.ErrUserNotFound
	}

	err = u.refreshTokenRepo.RevokeAllByUserID(ctx, user.PKID)
	if err != nil {
		return err
	}

	return u.userRepo.Delete(ctx, user.Code)
}
