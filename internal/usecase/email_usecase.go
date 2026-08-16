package usecase

import (
	"fmt"
	"go-gin-clean/internal/gateway/mailer"
	"go-gin-clean/pkg/config"
	"go-gin-clean/pkg/logger"

	"go.uber.org/zap"
)

type EmailUseCaseInterface interface {
	SendVerifyEmail(to, name, url string) error
	SendResetPasswordEmail(to, name, url string) error
}

type EmailUseCase struct {
	smtp mailer.MailerServiceInterface
	cfg  *config.ServerConfig
}

func NewEmailUseCase(smtp mailer.MailerServiceInterface, cfg *config.ServerConfig) EmailUseCaseInterface {
	return &EmailUseCase{
		smtp: smtp,
		cfg:  cfg,
	}
}

func (e *EmailUseCase) SendVerifyEmail(to, name, url string) error {
	subject := fmt.Sprintf("Verify your account - %s", e.cfg.AppName)

	data := map[string]any{
		"Name":            name,
		"VerificationURL": url,
	}

	body, err := e.smtp.LoadTemplate("verify_email", data)
	if err != nil {
		logger.Error("failed to load email template: %v", zap.Error(err))
	}

	return e.smtp.SendEmail(to, subject, body)
}

func (e *EmailUseCase) SendResetPasswordEmail(to, name, url string) error {
	subject := fmt.Sprintf("Reset your password - %s", e.cfg.AppName)

	data := map[string]any{
		"Name":             name,
		"ResetPasswordURL": url,
	}

	body, err := e.smtp.LoadTemplate("reset_password", data)
	if err != nil {
		logger.Error("failed to load email template: %v", zap.Error(err))
	}

	return e.smtp.SendEmail(to, subject, body)
}
