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
	"go-gin-clean/internal/model/validator"
	"go-gin-clean/internal/repository"
	"go-gin-clean/pkg/config"
	"go-gin-clean/pkg/errors"
	"go-gin-clean/pkg/utils"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
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
	userValidator *validator.UserValidator
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
	userValidator *validator.UserValidator,
) *UserUseCase {
	return &UserUseCase{
		userRepo:            userRepo,
		refreshTokenRepo:    refreshTokenRepo,
		jwtService:          jwtService,
		bcryptService:       bcryptService,
		oauthService:        oauthService,
		aesService:          aesService,
		cloudinaryService:   cloudinaryService,
		localStorageService: localStorageService,
		redisService:        redisService,
		UserPublisher:       UserPublisher,
		userValidator:       userValidator,
	}
}

func formatUserInfo(user *entity.User) *model.UserInfo {
	return &model.UserInfo{
		ID:       user.ID,
		Name:     user.Name,
		Code:     user.Code,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Gender:   user.Gender,
		IsActive: user.IsActive,
	}
}

// isValidImageExtension checks if the file extension is allowed for images
func isValidImageExtension(filename string) bool {
	allowedExtensions := []string{".jpg", ".jpeg", ".png"}
	filename = strings.ToLower(filename)
	for _, ext := range allowedExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
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
		if err != nil {
			return nil, appID, errors.ErrOAuthCallback
		}
	default:
		return nil, appID, errors.ErrInvalidOAuthProvider
	}

	existingUser, err := u.userRepo.FindByOAuthID(ctx, req.Provider, user.OAuthID)
	if err == nil {
		user = existingUser
	} else {
		existingUserByEmail, err := u.userRepo.FindByEmail(ctx, user.Email)
		if err == nil {
			err = u.userRepo.UpdateOAuthInfo(
				ctx,
				existingUserByEmail.ID.String(),
				req.Provider,
				user.OAuthID,
			)
			if err != nil {
				return nil, appID, errors.ErrLinkOAuth
			}

			user = existingUserByEmail
		} else {
			user, err = u.userRepo.Create(ctx, user)
			if err != nil {
				return nil, appID, errors.ErrOAuthSignUp
			}
		}
	}

	accessToken, _, err := u.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, appID, errors.ErrAccessToken
	}

	refreshToken, expiryAt, err := u.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, appID, errors.ErrRefreshToken
	}

	hashedRefreshToken, err := u.aesService.EncryptInternal(refreshToken)
	if err != nil {
		return nil, appID, errors.ErrProcessToken
	}

	tokenData := entity.NewRefreshToken(user.ID, hashedRefreshToken, expiryAt, false, *user)

	if err := u.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return nil, appID, errors.ErrRefreshToken
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

	if user.IsOAuthUser() {
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
		return nil, errors.ErrAccessToken
	}

	refreshToken, expiryAt, err := u.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, errors.ErrRefreshToken
	}

	hashedRefreshToken, err := u.aesService.EncryptInternal(refreshToken)
	if err != nil {
		return nil, errors.ErrProcessToken
	}

	tokenData := entity.NewRefreshToken(user.ID, hashedRefreshToken, expiryAt, false, *user)

	if err := u.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return nil, errors.ErrRefreshToken
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *UserUseCase) Register(ctx context.Context, req *model.RegisterRequest) error {
	if err := u.userValidator.Validate(req); err != nil {
		return err
	}

	if exist := u.userRepo.ExistByEmail(ctx, req.Email); exist {
		return errors.ErrEmailAlreadyExists
	}

	hashedPassword, err := u.bcryptService.HashPassword(req.Password)
	if err != nil {
		return errors.ErrProcessUserPassword
	}

	userData, err := entity.NewUser(req.Name, req.Email, hashedPassword, "", entity.Other)
	if err != nil {
		return errors.ErrRegisterFailed
	}

	if userData == nil {
		return errors.ErrInvalidInput
	}

	savedUser, err := u.userRepo.Create(ctx, userData)
	if err != nil {
		return errors.ErrCreateUser
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
				UserID: savedUser.ID,
				Name:   savedUser.Name,
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
		return nil, errors.ErrProcessToken
	}

	claims, err := u.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	if !u.refreshTokenRepo.IsTokenValid(ctx, refreshToken) {
		return nil, errors.ErrTokenInvalid
	}

	user, err := u.userRepo.FindByID(ctx, claims.UserID.String())
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	newAccessToken, _, err := u.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, errors.ErrAccessToken
	}

	newRefreshToken, expiryAt, err := u.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, errors.ErrRefreshToken
	}

	if err := u.refreshTokenRepo.RevokeByToken(ctx, refreshToken); err != nil {
		return nil, errors.ErrRefreshToken
	}

	newHashedRefreshToken, err := u.aesService.EncryptInternal(newRefreshToken)
	if err != nil {
		return nil, errors.ErrProcessToken
	}

	tokenData := entity.NewRefreshToken(user.ID, newHashedRefreshToken, expiryAt, false, *user)

	if err := u.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return nil, errors.ErrRefreshToken
	}

	return &model.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (u *UserUseCase) Logout(ctx context.Context, id string) error {
	if err := u.refreshTokenRepo.RevokeAllByUserID(ctx, id); err != nil {
		return errors.ErrTerminateAllSessions
	}
	return nil
}

func (u *UserUseCase) SendVerifyEmail(ctx context.Context, req model.SendVerifyEmailRequest) error {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return errors.ErrUserNotFound
	}

	plainText := fmt.Sprintf("%s_%s", user.Code, time.Now().Add(24*time.Hour).Format(time.RFC3339))

	token, err := u.aesService.EncryptURLSafe(plainText)
	if err != nil {
		return errors.ErrPrepareVerificationEmail
	}

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", config.GetAppURL(), token)

	message := model.RegisterEvent{
		UserEvent: model.UserEvent{
			UserID: user.ID,
			Name:   user.Name,
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
		return errors.ErrProcessToken
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
	if err != nil {
		return errors.ErrActivateUser
	}

	return nil
}

func (u *UserUseCase) SendResetPassword(ctx context.Context, req model.SendResetPasswordRequest) error {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return errors.ErrUserNotFound
	}

	plainText := fmt.Sprintf("%s_%s", user.Email, time.Now().Add(1*time.Hour).Format(time.RFC3339))

	token, err := u.aesService.EncryptURLSafe(plainText)
	if err != nil {
		return errors.ErrPrepareForgotPasswordEmail
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", config.GetAppURL(), token)

	message := model.ResetPasswordEvent{
		UserEvent: model.UserEvent{
			UserID: user.ID,
			Name:   user.Name,
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
	if err := u.userValidator.Validate(req); err != nil {
		return err
	}

	token, err := u.aesService.DecryptURLSafe(req.Token)
	if err != nil {
		return errors.ErrProcessToken
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
		return errors.ErrProcessUserPassword
	}

	user.SetPassword(hashedPassword)

	_, err = u.userRepo.Update(ctx, user, user.Code)
	if err != nil {
		return errors.ErrUpdateUserPassword
	}

	return nil
}

func (u *UserUseCase) GetAllUsers(ctx context.Context, page, pageSize int, search string) (*model.PaginationResponse[model.UserInfo], error) {
	// Create cache key
	cacheKey := fmt.Sprintf("users:all:page:%d:size:%d:search:%s", page, pageSize, search)

	var cachedResult model.PaginationResponse[model.UserInfo]
	err := u.redisService.Get(ctx, cacheKey, &cachedResult)
	if err == nil {
		log.Printf("Cache HIT for key: %s", cacheKey)
		return &cachedResult, nil
	}

	if err != redis.Nil {
		log.Printf("Cache error: %v", err)
	} else {
		log.Printf("Cache MISS for key: %s", cacheKey)
	}

	users, total, err := u.userRepo.FindAll(ctx, page, pageSize, search)
	if err != nil {
		return nil, errors.ErrGetAllUsers
	}

	userInfos := make([]model.UserInfo, len(users))
	for i, user := range users {
		userInfos[i] = *formatUserInfo(user)
	}

	result := model.NewPaginationResponse(userInfos, page, pageSize, int(total))

	// Store in cache (5 minutes expiration)
	if err := u.redisService.SetWithExpiration(ctx, cacheKey, result, 5*time.Minute); err != nil {
		log.Printf("Failed to cache result: %v", err)
	}

	return result, nil
}

func (u *UserUseCase) GetUserByCode(ctx context.Context, code string) (*model.UserInfo, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("user:code:%s", code)

	var cachedUser model.UserInfo
	err := u.redisService.Get(ctx, cacheKey, &cachedUser)
	if err == nil {
		log.Printf("Cache HIT for key: %s", cacheKey)
		return &cachedUser, nil
	}

	if err != redis.Nil {
		log.Printf("Cache error: %v", err)
	} else {
		log.Printf("Cache MISS for key: %s", cacheKey)
	}

	user, err := u.userRepo.FindByCode(ctx, code)
	if err != nil || user == nil {
		return nil, errors.ErrGetUserInformation
	}
	if !user.IsActive || !user.IsVerified {
		return nil, errors.ErrUserNotFound
	}

	userInfo := formatUserInfo(user)

	// Store in cache (5 minutes expiration)
	if err := u.redisService.SetWithExpiration(ctx, cacheKey, userInfo, 5*time.Minute); err != nil {
		log.Printf("Failed to cache result: %v", err)
	}

	return userInfo, nil
}

func (u *UserUseCase) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.UserInfo, error) {
	if err := u.userValidator.Validate(req); err != nil {
		return nil, err
	}

	if u.userRepo.ExistByEmail(ctx, req.Email) {
		return nil, errors.ErrEmailAlreadyExists
	}

	hashedPassword, err := u.bcryptService.HashPassword(req.Password)
	if err != nil {
		return nil, errors.ErrProcessUserPassword
	}

	userData, err := entity.NewUser(req.Name, req.Email, hashedPassword, "", entity.Other)
	if err != nil {
		return nil, errors.ErrInvalidInput
	}

	savedUser, err := u.userRepo.Create(ctx, userData)
	if err != nil {
		return nil, errors.ErrCreateUser
	}

	// Invalidate cache for all users list
	if err := u.redisService.DeletePattern(context.Background(), "users:all:*"); err != nil {
		log.Printf("Failed to invalidate users cache: %v", err)
	}

	return formatUserInfo(savedUser), nil
}

func (u *UserUseCase) UpdateUser(ctx context.Context, code string, req *model.UpdateUserRequest) (*model.UserInfo, error) {
	if err := u.userValidator.Validate(req); err != nil {
		return nil, err
	}

	user, err := u.userRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Avatar != nil {
		allowedExtensions := []string{".jpg", ".jpeg", ".png"}

		if !utils.IsValidExtension(req.Avatar.Filename, allowedExtensions) {
			return nil, errors.ErrUnsupportedImageType
		}

		// Choose one of the storage services to upload the avatar
		// path, err := u.cloudinaryService.UploadFile(ctx, req.Avatar.Filename, req.Avatar.Size, *req.Avatar, "users/"+user.Code+"/avatar/")
		// if err != nil || path == nil {
		// 	return nil, errors.ErrUploadImage
		// }

		path, err := u.localStorageService.UploadFile(
			ctx,
			fmt.Sprintf("avatar_%s_%d.jpg", user.ID.String(), time.Now().Unix()),
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
		return nil, errors.ErrUpdateUser
	}

	// Invalidate cache for this user and all users list
	cacheKey := fmt.Sprintf("user:code:%s", code)
	if err := u.redisService.Delete(context.Background(), cacheKey); err != nil {
		log.Printf("Failed to invalidate user cache: %v", err)
	}
	if err := u.redisService.DeletePattern(context.Background(), "users:all:*"); err != nil {
		log.Printf("Failed to invalidate users cache: %v", err)
	}

	return formatUserInfo(updatedUser), nil
}

func (u *UserUseCase) ChangePassword(ctx context.Context, userID string, req *model.ChangePasswordRequest) error {
	if err := u.userValidator.Validate(req); err != nil {
		return err
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	if err := u.bcryptService.ValidatePassword(req.OldPassword, user.Password); err != nil {
		return errors.ErrPasswordNotMatch
	}

	hashedPassword, err := u.bcryptService.HashPassword(req.NewPassword)
	if err != nil {
		return errors.ErrProcessUserPassword
	}

	user.SetPassword(hashedPassword)

	_, err = u.userRepo.Update(ctx, user, user.Code)
	if err != nil {
		return errors.ErrUpdateUserPassword
	}
	return nil
}

func (u *UserUseCase) ChangeStatus(ctx context.Context, code string, req model.ChangeUserStatusRequest) error {
	user, err := u.userRepo.FindByCode(ctx, code)
	if err != nil {
		return errors.ErrUserNotFound
	}

	user.IsActive = req.IsActive

	_, err = u.userRepo.Update(ctx, user, user.Code)
	if err != nil {
		return errors.ErrUpdateUserStatus
	}

	// Invalidate cache for this user and all users list
	cacheKey := fmt.Sprintf("user:code:%s", code)
	if err := u.redisService.Delete(context.Background(), cacheKey); err != nil {
		log.Printf("Failed to invalidate user cache: %v", err)
	}
	if err := u.redisService.DeletePattern(context.Background(), "users:all:*"); err != nil {
		log.Printf("Failed to invalidate users cache: %v", err)
	}

	return nil
}

func (u *UserUseCase) DeleteUser(ctx context.Context, code string) error {
	user, err := u.userRepo.FindByCode(ctx, code)
	if err != nil {
		return errors.ErrUserNotFound
	}

	err = u.refreshTokenRepo.RevokeAllByUserID(ctx, user.ID.String())
	if err != nil {
		return errors.ErrTerminateAllSessions
	}

	err = u.userRepo.Delete(ctx, user.Code)
	if err != nil {
		return errors.ErrDeleteUser
	}

	// Invalidate cache for this user and all users list
	cacheKey := fmt.Sprintf("user:code:%s", code)
	if err := u.redisService.Delete(context.Background(), cacheKey); err != nil {
		log.Printf("Failed to invalidate user cache: %v", err)
	}
	if err := u.redisService.DeletePattern(context.Background(), "users:all:*"); err != nil {
		log.Printf("Failed to invalidate users cache: %v", err)
	}

	return nil
}
