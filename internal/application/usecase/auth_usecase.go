package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-gin-clean/internal/application/port"
	"go-gin-clean/internal/domain/entity"
	"go-gin-clean/internal/dto"
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
	GetOAuthLoginURL(ctx context.Context, provider, appID, platform string) (*dto.OAuthUrlResponse, error)
	GetOAuthRedirectURL(appID, platform string) string
	HandleOAuthCallback(ctx context.Context, provider string, req *dto.OAuthCallbackRequest) (*dto.LoginResponse, string, string, error)
}

type authUseCase struct {
	userRepo         port.UserRepository
	refreshTokenRepo port.RefreshTokenRepository
	outboxUseCase    OutboxUseCase

	jwt       port.TokenMaker
	hasher    port.Hasher
	oauth     port.OAuthProvider
	encryptor port.Encryptor
	cfg       *config.ServerConfig
}

func NewAuthUseCase(
	userRepo port.UserRepository,
	refreshTokenRepo port.RefreshTokenRepository,
	outboxUseCase OutboxUseCase,
	jwt port.TokenMaker,
	hasher port.Hasher,
	oauth port.OAuthProvider,
	encryptor port.Encryptor,
	cfg *config.ServerConfig,
) AuthUseCase {
	return &authUseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		outboxUseCase:    outboxUseCase,
		jwt:              jwt,
		hasher:           hasher,
		oauth:            oauth,
		encryptor:        encryptor,
		cfg:              cfg,
	}
}

func (a *authUseCase) GetOAuthRedirectURL(appID, platform string) string {
	if platform == "mobile" {
		return a.oauth.GetMobileDeepLinkURL(appID)
	}
	return a.oauth.GetFrontendURL(appID)
}

func (a *authUseCase) GetOAuthLoginURL(ctx context.Context, provider, appID, platform string) (*dto.OAuthUrlResponse, error) {
	switch provider {
	case "google":
		return &dto.OAuthUrlResponse{
			AuthURL: a.oauth.GetGoogleAuthURL(appID, platform),
		}, nil
	default:
		return nil, pkgerrors.NewAppError(pkgerrors.Unprocessable, message.ErrInvalidOAuthProvider)
	}
}

func (a *authUseCase) HandleOAuthCallback(ctx context.Context, provider string, req *dto.OAuthCallbackRequest) (*dto.LoginResponse, string, string, error) {
	var (
		user     *entity.User
		appID    string
		platform string
		err      error
	)

	switch provider {
	case "google":
		user, appID, platform, err = a.oauth.HandleGoogleCallback(ctx, req.State, req.Code)
		if err != nil {
			return nil, appID, platform, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrOAuthCallback, err)
		}
	default:
		return nil, appID, platform, pkgerrors.NewAppError(pkgerrors.Unprocessable, message.ErrInvalidOAuthProvider)
	}

	existingUser, err := a.userRepo.FindByOAuthID(ctx, provider, user.OAuthID)
	if err == nil {
		user = existingUser
	} else {
		existingUserByEmail, err := a.userRepo.FindByEmail(ctx, user.Email)
		if err == nil {
			if err = a.userRepo.UpdateOAuthInfo(ctx, existingUserByEmail.ID.String(), provider, user.OAuthID); err != nil {
				return nil, appID, platform, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrLinkOAuth, err)
			}
			user = existingUserByEmail
		} else {
			user, err = a.userRepo.Create(ctx, user)
			if err != nil {
				return nil, appID, platform, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrOAuthSignUp, err)
			}
		}
	}

	accessToken, _, err := a.jwt.GenerateAccessToken(user.ID, user.Role.String())
	if err != nil {
		return nil, appID, platform, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrAccessToken, err)
	}

	refreshToken, expiryAt, err := a.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, appID, platform, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRefreshToken, err)
	}

	hashedRefreshToken, err := a.encryptor.EncryptInternal(refreshToken)
	if err != nil {
		return nil, appID, platform, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrProcessToken, err)
	}

	tokenData := entity.NewRefreshToken(user.ID, hashedRefreshToken, expiryAt, false, user)
	if err := a.refreshTokenRepo.Save(ctx, tokenData); err != nil {
		return nil, appID, platform, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRefreshToken, err)
	}

	return dto.FormatLoginResponse(accessToken, hashedRefreshToken), appID, platform, nil
}

func (a *authUseCase) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := a.userRepo.FindByEmail(ctx, req.Email)
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

	if err := a.hasher.ValidatePassword(req.Password, user.Password); err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Unauthorized, message.ErrLoginFailed, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrPasswordNotMatch))
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
		return pkgerrors.WrapAppError(pkgerrors.Conflict, message.ErrRegisterFailed, pkgerrors.NewAppError(pkgerrors.Conflict, message.ErrEmailAlreadyExists))
	}

	hashedPassword, err := a.hasher.HashPassword(req.Password)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRegisterFailed, err)
	}

	userData, err := entity.NewUser(req.Name, req.Email, hashedPassword)
	if err != nil || userData == nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrRegisterFailed, err)
	}

	savedUser, err := a.userRepo.Create(ctx, userData)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRegisterFailed, err)
	}

	plainText := fmt.Sprintf("%s_%s", savedUser.ID.String(), time.Now().Add(24*time.Hour).Format(time.RFC3339))
	token, err := a.encryptor.EncryptURLSafe(plainText)
	if err != nil {
		logger.Error("failed to prepare verification email", zap.Error(err))
		return nil
	}

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", a.cfg.AppUrl, token)
	message := entity.UserRegisterEvent{
		UserEvent:       entity.UserEvent{UserID: savedUser.ID, Name: savedUser.Name},
		Email:           savedUser.Email,
		VerificationURL: verificationURL,
	}

	if err := a.outboxUseCase.SaveOutboxMessage(ctx, "user", savedUser.ID.String(), entity.EventUserRegistered, message); err != nil {
		logger.Error("failed to save outbox message for register event", zap.Error(err))
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
		return nil, pkgerrors.WrapAppError(pkgerrors.Unauthorized, message.ErrRefreshToken, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrTokenInvalid))
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

	if err := a.refreshTokenRepo.RevokeByToken(ctx, refreshToken); err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRefreshToken, err)
	}

	newHashedRefreshToken, err := a.encryptor.EncryptInternal(newRefreshToken)
	if err != nil {
		return nil, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrRefreshToken, err)
	}

	tokenData := entity.NewRefreshToken(user.ID, newHashedRefreshToken, expiryAt, false, user)
	if err := a.refreshTokenRepo.Save(ctx, tokenData); err != nil {
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
	message := entity.UserRegisterEvent{
		UserEvent:       entity.UserEvent{UserID: user.ID, Name: user.Name},
		Email:           user.Email,
		VerificationURL: verificationURL,
	}

	if err := a.outboxUseCase.SaveOutboxMessage(ctx, "user", user.ID.String(), entity.EventUserRegistered, message); err != nil {
		logger.Error("failed to save outbox message for verify email event", zap.Error(err))
	}

	return nil
}

func (a *authUseCase) VerifyEmail(ctx context.Context, token string) error {
	decrypted, err := a.encryptor.DecryptURLSafe(token)
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

	user, err := a.userRepo.FindByID(ctx, payload[0])
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
	message := entity.UserResetPasswordEvent{
		UserEvent: entity.UserEvent{UserID: user.ID, Name: user.Name},
		Email:     user.Email,
		ResetURL:  resetURL,
	}

	if err := a.outboxUseCase.SaveOutboxMessage(ctx, "user", user.ID.String(), entity.EventUserResetPassword, message); err != nil {
		logger.Error("failed to save outbox message for reset password event", zap.Error(err))
	}

	return nil
}

func (a *authUseCase) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	decrypted, err := a.encryptor.DecryptURLSafe(req.Token)
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

	user, err := a.userRepo.FindByEmail(ctx, payload[0])
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrResetPasswordFailed, err)
	}

	hashedPassword, err := a.hasher.HashPassword(req.NewPassword)
	if err != nil {
		return pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrResetPasswordFailed, err)
	}

	user.SetPassword(hashedPassword)
	if _, err = a.userRepo.Update(ctx, user); err != nil {
		return pkgerrors.WrapAppError(pkgerrors.BadRequest, message.ErrResetPasswordFailed, err)
	}

	return nil
}
