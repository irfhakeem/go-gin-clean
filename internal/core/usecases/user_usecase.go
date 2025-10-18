package usecases

import (
	"context"
	"fmt"
	"go-gin-clean/internal/core/contracts"
	"go-gin-clean/internal/core/domain/entities"
	"go-gin-clean/internal/core/domain/enums"
	"go-gin-clean/internal/core/domain/errors"
	"go-gin-clean/internal/core/ports"
	"go-gin-clean/pkg/config"
	"log"
	"strconv"
	"strings"
	"time"
)

type UserUseCase struct {
	userRepo            ports.UserRepository
	email               ports.EmailUseCase
	refreshTokenRepo    ports.RefreshTokenRepository
	jwtService          ports.JWTService
	bcryptService       ports.BcryptService
	aesService          ports.EncryptionService
	localStorageService ports.MediaService
}

func NewUserUseCase(
	userRepo ports.UserRepository,
	email ports.EmailUseCase,
	refreshTokenRepo ports.RefreshTokenRepository,
	jwtService ports.JWTService,
	bcryptService ports.BcryptService,
	aesService ports.EncryptionService,
	localStorageService ports.MediaService,
) ports.UserUseCase {
	return &UserUseCase{
		userRepo:            userRepo,
		email:               email,
		refreshTokenRepo:    refreshTokenRepo,
		jwtService:          jwtService,
		bcryptService:       bcryptService,
		aesService:          aesService,
		localStorageService: localStorageService,
	}
}

func FormatUserInfo(user *entities.User) *contracts.UserInfo {
	return &contracts.UserInfo{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Gender:   user.Gender,
		Avatar:   user.Avatar,
		IsActive: user.IsActive,
	}
}

func (uc *UserUseCase) Login(ctx context.Context, req *contracts.LoginRequest) (*contracts.LoginResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if !user.IsActive {
		return nil, errors.ErrUserNotFound
	}

	if err := uc.bcryptService.ValidatePassword(req.Password, user.Password); err != nil {
		return nil, errors.ErrPasswordNotMatch
	}

	accessToken, _, err := uc.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, expiryAt, err := uc.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	hashedRefreshToken, err := uc.aesService.EncryptInternal(refreshToken)
	if err != nil {
		return nil, err
	}

	token := entities.NewRefreshToken(user.ID, hashedRefreshToken, expiryAt, false, *user)

	if err := uc.refreshTokenRepo.Save(ctx, token); err != nil {
		return nil, err
	}

	return &contracts.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *FormatUserInfo(user),
	}, nil
}

func (uc *UserUseCase) Register(ctx context.Context, req *contracts.RegisterRequest) error {
	if uc.userRepo.ExistsByEmail(ctx, req.Email) {
		return errors.ErrEmailAlreadyExists
	}

	hashedPassword, err := uc.bcryptService.HashPassword(req.Password)
	if err != nil {
		return err
	}

	user, err := entities.NewUser(req.Name, req.Email, hashedPassword, "", enums.Unknown)
	if err != nil {
		return err
	}

	savedUser, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		return err
	}

	plainText := fmt.Sprintf("%d_%s", savedUser.ID, time.Now().Add(24*time.Hour).Format(time.RFC3339))

	token, err := uc.aesService.EncryptURLSafe(plainText)
	if err != nil {
		return err
	}

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", config.GetAppURL(), token)

	go func() {
		if err := uc.email.SendVerifyEmail(user.Email, user.Name, verificationURL); err != nil {
			log.Printf("Failed to send verification email to %s: %v", user.Email, err)
		}
	}()

	return nil
}

func (uc *UserUseCase) RefreshToken(ctx context.Context, refreshToken string) (*contracts.RefreshTokenResponse, error) {
	claims, err := uc.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	if !uc.refreshTokenRepo.IsTokenValid(ctx, refreshToken) {
		return nil, errors.ErrTokenInvalid
	}

	user, err := uc.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	newAccessToken, _, err := uc.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, expiryAt, err := uc.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	if err := uc.refreshTokenRepo.RevokeByToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	hashedRefreshToken, err := uc.aesService.EncryptInternal(newRefreshToken)
	if err != nil {
		return nil, err
	}

	token := entities.NewRefreshToken(user.ID, hashedRefreshToken, expiryAt, false, *user)

	if err := uc.refreshTokenRepo.Save(ctx, token); err != nil {
		return nil, err
	}

	return &contracts.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (uc *UserUseCase) Logout(ctx context.Context, userID int64) error {
	return uc.refreshTokenRepo.RevokeAllByUserID(ctx, userID)
}

func (uc *UserUseCase) VerifyEmail(ctx context.Context, token string) error {
	token, err := uc.aesService.DecryptURLSafe(token)
	if err != nil {
		return errors.ErrTokenInvalid
	}

	payloads := strings.Split(token, "_")
	if len(payloads) != 2 {
		return errors.ErrTokenInvalid
	}

	expiryAt, err := time.Parse(time.RFC3339, payloads[1])
	if err != nil {
		return errors.ErrTokenInvalid
	}

	if time.Now().After(expiryAt) {
		return errors.ErrTokenExpired
	}

	userID, err := strconv.ParseInt(payloads[0], 10, 64)
	if err != nil {
		return errors.ErrInvalidIDFormat
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	user.Activate()

	_, err = uc.userRepo.Update(ctx, user)
	return err
}

func (uc *UserUseCase) SendVerifyEmail(ctx context.Context, email string) error {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return errors.ErrUserNotFound
	}

	plainText := fmt.Sprintf("%d_%s", user.ID, time.Now().Add(24*time.Hour).Format(time.RFC3339))

	token, err := uc.aesService.EncryptURLSafe(plainText)
	if err != nil {
		return err
	}

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", config.GetAppURL(), token)

	go func() {
		if err := uc.email.SendVerifyEmail(user.Email, user.Name, verificationURL); err != nil {
			log.Printf("Failed to send verification email to %s: %v", user.Email, err)
		}
	}()

	return nil
}

func (uc *UserUseCase) SendResetPassword(ctx context.Context, email string) error {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return errors.ErrUserNotFound
	}

	plainText := fmt.Sprintf("%s_%s", user.Email, time.Now().Add(1*time.Hour).Format(time.RFC3339))

	token, err := uc.aesService.EncryptURLSafe(plainText)
	if err != nil {
		return err
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", config.GetAppURL(), token)

	go func() {
		if err := uc.email.SendResetPasswordEmail(user.Email, user.Name, resetURL); err != nil {
			log.Printf("Failed to send password reset email to %s: %v", user.Email, err)
		}
	}()

	return nil
}

func (uc *UserUseCase) ResetPassword(ctx context.Context, req *contracts.ResetPasswordRequest) error {
	paylaod, err := uc.aesService.DecryptURLSafe(req.Token)
	if err != nil {
		return errors.ErrTokenInvalid
	}

	parts := strings.Split(paylaod, "_")
	if len(parts) != 2 {
		return errors.ErrTokenInvalid
	}

	email := parts[0]
	expiryAt, err := time.Parse(time.RFC3339, parts[1])
	if err != nil {
		return errors.ErrTokenInvalid
	}

	if time.Now().After(expiryAt) {
		return errors.ErrTokenExpired
	}

	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return errors.ErrUserNotFound
	}

	hashedPassword, err := uc.bcryptService.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	if err := user.ChangePassword(hashedPassword); err != nil {
		return err
	}

	_, err = uc.userRepo.Update(ctx, user)
	return err
}

func (uc *UserUseCase) GetAllUsers(ctx context.Context, page, pageSize int, search string) (*contracts.PaginationResponse[contracts.UserInfo], error) {
	offset := contracts.Offset(page, pageSize)
	users, total, err := uc.userRepo.FindAll(ctx, pageSize, offset, search)
	if err != nil {
		return nil, err
	}

	userInfos := make([]contracts.UserInfo, len(users))
	for i, user := range users {
		userInfos[i] = *FormatUserInfo(user)
	}

	return contracts.NewPaginationResponse(userInfos, page, pageSize, int(total)), nil
}

func (uc *UserUseCase) GetUserByID(ctx context.Context, userID int64) (*contracts.UserInfo, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	return FormatUserInfo(user), nil
}

func (uc *UserUseCase) CreateUser(ctx context.Context, req *contracts.CreateUserRequest) (*contracts.UserInfo, error) {
	if uc.userRepo.ExistsByEmail(ctx, req.Email) {
		return nil, errors.ErrEmailAlreadyExists
	}

	hashedPassword, err := uc.bcryptService.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user, err := entities.NewUser(req.Name, req.Email, hashedPassword, "", enums.Unknown)
	if err != nil {
		return nil, err
	}

	user.IsActive = true

	savedUser, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return FormatUserInfo(savedUser), nil
}

func (uc *UserUseCase) UpdateUser(ctx context.Context, userID int64, req *contracts.UpdateUserRequest) (*contracts.UserInfo, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Avatar != nil {
		path := fmt.Sprintf("avatars/user_%d/", user.ID)

		filePath, err := uc.localStorageService.UploadFile(req.Avatar.Filename, req.Avatar.Size, req.Avatar.Content, path)
		if err != nil {
			return nil, err
		}

		user.Avatar = *filePath
	}

	if req.Gender != nil {
		user.Gender = *req.Gender
	}

	updatedUser, err := uc.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return FormatUserInfo(updatedUser), nil
}

func (uc *UserUseCase) ChangePassword(ctx context.Context, userID int64, req *contracts.ChangePasswordRequest) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	if err := uc.bcryptService.ValidatePassword(req.OldPassword, user.Password); err != nil {
		return errors.ErrPasswordNotMatch
	}

	hashedPassword, err := uc.bcryptService.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	if err := user.ChangePassword(hashedPassword); err != nil {
		return err
	}

	_, err = uc.userRepo.Update(ctx, user)
	return err
}

func (uc *UserUseCase) DeleteUser(ctx context.Context, userID int64) error {
	return uc.userRepo.Delete(ctx, userID)
}
