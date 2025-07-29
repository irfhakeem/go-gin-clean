package usecases

import (
	"context"
	"go-gin-clean/internal/core/domain/entities"
	"go-gin-clean/internal/core/domain/errors"
	vo "go-gin-clean/internal/core/domain/valueobjects"
	"go-gin-clean/internal/core/dto"
	"go-gin-clean/internal/core/ports"
	"time"
)

type UserUseCase struct {
	userRepo         ports.UserRepository
	refreshTokenRepo ports.RefreshTokenRepository
	jwtService       ports.JWTService
	bcryptService    ports.BcryptService
	emailService     ports.EmailService
}

func NewUserUseCase(
	userRepo ports.UserRepository,
	refreshTokenRepo ports.RefreshTokenRepository,
	jwtService ports.JWTService,
	bcryptService ports.BcryptService,
) ports.UserUseCase {
	return &UserUseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
		bcryptService:    bcryptService,
	}
}

func (uc *UserUseCase) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if !user.IsActive {
		return nil, errors.ErrUserNotFound
	}

	if err := uc.bcryptService.ValidatePassword(req.Password, user.Password.Value()); err != nil {
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

	token := &entities.RefreshToken{
		UserID:   user.ID,
		Token:    refreshToken,
		ExpiryAt: expiryAt,
	}

	if err := uc.refreshTokenRepo.Save(ctx, token); err != nil {
		return nil, err
	}

	userInfo := dto.UserInfo{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email.Value(),
		Avatar:   user.Avatar,
		Gender:   user.Gender,
		IsActive: user.IsActive,
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userInfo,
	}, nil
}

func (uc *UserUseCase) Register(ctx context.Context, req *dto.RegisterRequest) error {
	if uc.userRepo.ExistsByEmail(ctx, req.Email) {
		return errors.ErrEmailAlreadyExists
	}

	hashedPassword, err := uc.bcryptService.HashPassword(req.Password)
	if err != nil {
		return err
	}

	email, err := vo.NewEmailAddress(req.Email)
	if err != nil {
		return err
	}

	password, err := vo.NewPassword(hashedPassword)
	if err != nil {
		return err
	}

	user := &entities.User{
		Name:     req.Name,
		Email:    email,
		Password: password,
		Audit: entities.Audit{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	savedUser, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		return err
	}

	return uc.SendVerifyEmail(ctx, savedUser.Email.Value())
}

func (uc *UserUseCase) RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshTokenResponse, error) {
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

	token := &entities.RefreshToken{
		UserID:   user.ID,
		Token:    newRefreshToken,
		ExpiryAt: expiryAt,
	}

	if err := uc.refreshTokenRepo.Save(ctx, token); err != nil {
		return nil, err
	}

	return &dto.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (uc *UserUseCase) Logout(ctx context.Context, userID int64) error {
	return uc.refreshTokenRepo.RevokeAllByUserID(ctx, userID)
}

func (uc *UserUseCase) VerifyEmail(ctx context.Context, token string) error {
	claims, err := uc.jwtService.ValidateAccessToken(token)
	if err != nil {
		return errors.ErrTokenInvalid
	}

	user, err := uc.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	user.Activate()
	user.UpdateTimestamp()

	_, err = uc.userRepo.Update(ctx, user)
	return err
}

func (uc *UserUseCase) SendVerifyEmail(ctx context.Context, email string) error {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return errors.ErrUserNotFound
	}

	token, _, err := uc.jwtService.GenerateAccessToken(user)
	if err != nil {
		return err
	}

	return uc.emailService.SendVerifyEmail(user.Email.Value(), user.Name, token)
}

func (uc *UserUseCase) SendResetPassword(ctx context.Context, email string) error {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return errors.ErrUserNotFound
	}

	token, _, err := uc.jwtService.GenerateAccessToken(user)
	if err != nil {
		return err
	}

	return uc.emailService.SendResetPasswordEmail(user.Email.Value(), user.Name, token)
}

func (uc *UserUseCase) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return errors.ErrUserNotFound
	}

	hashedPassword, err := uc.bcryptService.HashPassword("temporary_password")
	if err != nil {
		return err
	}

	if err := user.ChangePassword(hashedPassword); err != nil {
		return err
	}

	user.UpdateTimestamp()
	_, err = uc.userRepo.Update(ctx, user)
	return err
}

func (uc *UserUseCase) GetAllUsers(ctx context.Context, page, pageSize int, search string) (*dto.PaginationResponse[dto.UserInfo], error) {
	offset := dto.Offset(page, pageSize)
	users, total, err := uc.userRepo.FindAll(ctx, pageSize, offset, search)
	if err != nil {
		return nil, err
	}

	userInfos := make([]dto.UserInfo, len(users))
	for i, user := range users {
		userInfos[i] = dto.UserInfo{
			ID:       user.ID,
			Name:     user.Name,
			Email:    user.Email.Value(),
			Avatar:   user.Avatar,
			Gender:   user.Gender,
			IsActive: user.IsActive,
		}
	}

	return dto.NewPaginationResponse(userInfos, page, pageSize, int(total)), nil
}

func (uc *UserUseCase) GetUserByID(ctx context.Context, userID int64) (*dto.UserInfo, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	return &dto.UserInfo{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email.Value(),
		Avatar:   user.Avatar,
		Gender:   user.Gender,
		IsActive: user.IsActive,
	}, nil
}

func (uc *UserUseCase) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserInfo, error) {
	if uc.userRepo.ExistsByEmail(ctx, req.Email) {
		return nil, errors.ErrEmailAlreadyExists
	}

	hashedPassword, err := uc.bcryptService.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	email, err := vo.NewEmailAddress(req.Email)
	if err != nil {
		return nil, err
	}

	password, err := vo.NewPassword(hashedPassword)
	if err != nil {
		return nil, err
	}

	user := &entities.User{
		Name:     req.Name,
		Email:    email,
		Password: password,
		Gender:   req.Gender,
		IsActive: true,
		Audit: entities.Audit{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	savedUser, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return &dto.UserInfo{
		ID:       savedUser.ID,
		Name:     savedUser.Name,
		Email:    savedUser.Email.Value(),
		Avatar:   savedUser.Avatar,
		Gender:   savedUser.Gender,
		IsActive: savedUser.IsActive,
	}, nil
}

func (uc *UserUseCase) UpdateUser(ctx context.Context, userID int64, req *dto.UpdateUserRequest) (*dto.UserInfo, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}

	if req.Gender != nil {
		user.Gender = *req.Gender
	}

	user.UpdateTimestamp()

	updatedUser, err := uc.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return &dto.UserInfo{
		ID:       updatedUser.ID,
		Name:     updatedUser.Name,
		Email:    updatedUser.Email.Value(),
		Avatar:   updatedUser.Avatar,
		Gender:   updatedUser.Gender,
		IsActive: updatedUser.IsActive,
	}, nil
}

func (uc *UserUseCase) ChangePassword(ctx context.Context, userID int64, req *dto.ChangePasswordRequest) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	if err := uc.bcryptService.ValidatePassword(req.OldPassword, user.Password.Value()); err != nil {
		return errors.ErrPasswordNotMatch
	}

	hashedPassword, err := uc.bcryptService.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	if err := user.ChangePassword(hashedPassword); err != nil {
		return err
	}

	user.UpdateTimestamp()
	_, err = uc.userRepo.Update(ctx, user)
	return err
}

func (uc *UserUseCase) DeleteUser(ctx context.Context, userID int64) error {
	return uc.userRepo.Delete(ctx, userID)
}
