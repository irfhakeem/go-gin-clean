package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-gin-clean/internal/application/port"
	"go-gin-clean/internal/domain/entity"
	"go-gin-clean/internal/dto"
	"go-gin-clean/internal/dto/event"
	"go-gin-clean/pkg/config"
	pkgerrors "go-gin-clean/pkg/errors"
	"go-gin-clean/pkg/logger"
	"go-gin-clean/pkg/message"

	"go.uber.org/zap"
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

type authUseCase struct {
	txManager        port.TransactionManager
	userRepo         port.UserRepository
	refreshTokenRepo port.RefreshTokenRepository
	outboxUseCase    OutboxUseCase

	jwt       port.TokenMaker
	hasher    port.Hasher
	encryptor port.Encryptor
	cfg       *config.ServerConfig
}

func NewAuthUseCase(
	txManager port.TransactionManager,
	userRepo port.UserRepository,
	refreshTokenRepo port.RefreshTokenRepository,
	outboxUseCase OutboxUseCase,
	jwt port.TokenMaker,
	hasher port.Hasher,
	encryptor port.Encryptor,
	cfg *config.ServerConfig,
) AuthUseCase {
	return &authUseCase{
		txManager:        txManager,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		outboxUseCase:    outboxUseCase,
		jwt:              jwt,
		hasher:           hasher,
		encryptor:        encryptor,
		cfg:              cfg,
	}
}

func (a *authUseCase) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := a.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Unauthorized, message.ErrLoginFailed, err)
	}

	if !user.IsActive {
		return nil, pkgerrors.NewAppError(pkgerrors.NotFound, message.ErrUserNotFound)
	}

	if !user.IsVerified {
		return nil, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrEmailNotVerified)
	}

	if err := a.hasher.ValidatePassword(req.Password, user.Password); err != nil {
		return nil, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrPasswordNotMatch)
	}

	accessToken, _, err := a.jwt.GenerateAccessToken(user.ID, user.Role.String())
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrLoginFailed, err)
	}

	refreshToken, expiryAt, err := a.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrLoginFailed, err)
	}

	hashedRefreshToken, err := a.encryptor.EncryptInternal(refreshToken)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrLoginFailed, err)
	}

	tokenData := entity.NewRefreshToken(user.ID, hashedRefreshToken, expiryAt, false, user)
	if err := a.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrLoginFailed, err)
	}

	return dto.FormatLoginResponse(accessToken, hashedRefreshToken), nil
}

func (a *authUseCase) Register(ctx context.Context, req *dto.RegisterRequest) error {
	if exist := a.userRepo.ExistByEmail(ctx, req.Email); exist {
		return pkgerrors.NewAppError(pkgerrors.Conflict, message.ErrEmailAlreadyExists)
	}

	hashedPassword, err := a.hasher.HashPassword(req.Password)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRegisterFailed, err)
	}

	userData, err := entity.NewUser(req.Name, req.Email, hashedPassword)
	if err != nil || userData == nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrRegisterFailed, err)
	}

	err = a.txManager.WithTx(ctx, func(txCtx context.Context) error {
		savedUser, err := a.userRepo.Create(txCtx, userData)
		if err != nil {
			return err
		}

		plainText := fmt.Sprintf("%s_%s", savedUser.ID.String(), time.Now().Add(24*time.Hour).Format(time.RFC3339))
		token, err := a.encryptor.EncryptURLSafe(plainText)
		if err != nil {
			return err
		}

		verificationURL := fmt.Sprintf("%s/verify-email?token=%s", a.cfg.AppUrl, token)
		registerEvent := event.UserRegisterEvent{
			UserEvent:       event.UserEvent{UserID: savedUser.ID, Name: savedUser.Name},
			Email:           savedUser.Email,
			VerificationURL: verificationURL,
		}

		return a.outboxUseCase.SaveOutboxMessage(txCtx, "user", savedUser.ID.String(), entity.EventUserRegistered, registerEvent)
	})
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRegisterFailed, err)
	}

	return nil
}

func (a *authUseCase) RefreshToken(ctx context.Context, hashedRefreshToken string) (*dto.RefreshTokenResponse, error) {
	refreshToken, err := a.encryptor.DecryptInternal(hashedRefreshToken)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Unauthorized, message.ErrRefreshToken, err)
	}

	claims, err := a.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, pkgerrors.AsAppError(pkgerrors.Unauthorized, message.ErrRefreshToken, err)
	}

	if !a.refreshTokenRepo.IsTokenValid(ctx, refreshToken) {
		return nil, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrTokenInvalid)
	}

	user, err := a.userRepo.FindByID(ctx, claims.UserID.String())
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Unauthorized, message.ErrRefreshToken, err)
	}

	newAccessToken, _, err := a.jwt.GenerateAccessToken(user.ID, user.Role.String())
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRefreshToken, err)
	}

	newRefreshToken, expiryAt, err := a.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRefreshToken, err)
	}

	newHashedRefreshToken, err := a.encryptor.EncryptInternal(newRefreshToken)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRefreshToken, err)
	}

	tokenData := entity.NewRefreshToken(user.ID, newHashedRefreshToken, expiryAt, false, user)
	err = a.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := a.refreshTokenRepo.RevokeByToken(txCtx, refreshToken); err != nil {
			return err
		}
		return a.refreshTokenRepo.Save(txCtx, tokenData)
	})
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRefreshToken, err)
	}

	return &dto.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (a *authUseCase) Logout(ctx context.Context, id string) error {
	if err := a.refreshTokenRepo.RevokeAllByUserID(ctx, id); err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrLogoutFailed, err)
	}
	return nil
}

func (a *authUseCase) SendVerifyEmail(ctx context.Context, req *dto.SendVerifyEmailRequest) error {
	user, err := a.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrPrepareVerificationEmail, err)
	}

	plainText := fmt.Sprintf("%s_%s", user.ID.String(), time.Now().Add(24*time.Hour).Format(time.RFC3339))
	token, err := a.encryptor.EncryptURLSafe(plainText)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrPrepareVerificationEmail, err)
	}

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", a.cfg.AppUrl, token)
	verifyEvent := event.UserRegisterEvent{
		UserEvent:       event.UserEvent{UserID: user.ID, Name: user.Name},
		Email:           user.Email,
		VerificationURL: verificationURL,
	}

	if err := a.outboxUseCase.SaveOutboxMessage(ctx, "user", user.ID.String(), entity.EventUserRegistered, verifyEvent); err != nil {
		logger.Error("failed to save outbox message for verify email event", zap.Error(err))
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrPrepareVerificationEmail, err)
	}

	return nil
}

func (a *authUseCase) VerifyEmail(ctx context.Context, token string) error {
	decrypted, err := a.encryptor.DecryptURLSafe(token)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrVerifyEmailFailed, err)
	}

	lastUnderscore := strings.LastIndex(decrypted, "_")
	if lastUnderscore == -1 {
		return pkgerrors.NewAppError(pkgerrors.BadRequest, message.ErrTokenInvalid)
	}
	userID := decrypted[:lastUnderscore]
	expiryStr := decrypted[lastUnderscore+1:]

	expiryTime, err := time.Parse(time.RFC3339, expiryStr)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrVerifyEmailFailed, err)
	}

	if time.Now().After(expiryTime) {
		return pkgerrors.NewAppError(pkgerrors.BadRequest, message.ErrTokenExpired)
	}

	user, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrVerifyEmailFailed, err)
	}

	user.VerifyEmail()
	if _, err = a.userRepo.Update(ctx, user); err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrVerifyEmailFailed, err)
	}

	return nil
}

func (a *authUseCase) SendResetPassword(ctx context.Context, req *dto.SendResetPasswordRequest) error {
	user, err := a.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrPrepareForgotPasswordEmail, err)
	}

	plainText := fmt.Sprintf("%s_%s", user.Email, time.Now().Add(1*time.Hour).Format(time.RFC3339))
	token, err := a.encryptor.EncryptURLSafe(plainText)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrPrepareForgotPasswordEmail, err)
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", a.cfg.AppUrl, token)
	resetEvent := event.UserResetPasswordEvent{
		UserEvent: event.UserEvent{UserID: user.ID, Name: user.Name},
		Email:     user.Email,
		ResetURL:  resetURL,
	}

	if err := a.outboxUseCase.SaveOutboxMessage(ctx, "user", user.ID.String(), entity.EventUserResetPassword, resetEvent); err != nil {
		logger.Error("failed to save outbox message for reset password event", zap.Error(err))
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrPrepareForgotPasswordEmail, err)
	}
	return nil
}

func (a *authUseCase) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	decrypted, err := a.encryptor.DecryptURLSafe(req.Token)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrResetPasswordFailed, err)
	}

	lastUnderscore := strings.LastIndex(decrypted, "_")
	if lastUnderscore == -1 {
		return pkgerrors.NewAppError(pkgerrors.BadRequest, message.ErrTokenInvalid)
	}
	email := decrypted[:lastUnderscore]
	expiryStr := decrypted[lastUnderscore+1:]

	expiryTime, err := time.Parse(time.RFC3339, expiryStr)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrResetPasswordFailed, err)
	}

	if time.Now().After(expiryTime) {
		return pkgerrors.NewAppError(pkgerrors.BadRequest, message.ErrTokenExpired)
	}

	user, err := a.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrResetPasswordFailed, err)
	}

	hashedPassword, err := a.hasher.HashPassword(req.NewPassword)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrResetPasswordFailed, err)
	}

	user.SetPassword(hashedPassword)
	return a.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := a.userRepo.Update(txCtx, user); err != nil {
			return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrResetPasswordFailed, err)
		}

		if err := a.refreshTokenRepo.RevokeAllByUserID(txCtx, user.ID.String()); err != nil {
			return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrResetPasswordFailed, err)
		}

		return nil
	})
}
