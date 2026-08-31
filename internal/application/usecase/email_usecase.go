package usecase

import (
	"fmt"

	"go-gin-clean/internal/application/port"
	"go-gin-clean/pkg/config"
	"go-gin-clean/pkg/logger"

	"go.uber.org/zap"
)

type EmailUseCase interface {
	SendVerifyEmail(to, name, url string) error
	SendResetPasswordEmail(to, name, url string) error
}

type emailUseCase struct {
	mailer port.Mailer
	cfg    *config.ServerConfig
}

func NewEmailUseCase(mailer port.Mailer, cfg *config.ServerConfig) EmailUseCase {
	return &emailUseCase{
		mailer: mailer,
		cfg:    cfg,
	}
}

func (e *emailUseCase) SendVerifyEmail(to, name, url string) error {
	subject := fmt.Sprintf("Verify your account - %s", e.cfg.AppName)

	data := map[string]any{
		"Name":            name,
		"VerificationURL": url,
	}

	body, err := e.mailer.LoadTemplate("verify_email", data)
	if err != nil {
		logger.Error("failed to load email template: %v", zap.Error(err))
	}

	return e.mailer.SendEmail(to, subject, body)
}

func (e *emailUseCase) SendResetPasswordEmail(to, name, url string) error {
	subject := fmt.Sprintf("Reset your password - %s", e.cfg.AppName)

	data := map[string]any{
		"Name":             name,
		"ResetPasswordURL": url,
	}

	body, err := e.mailer.LoadTemplate("reset_password", data)
	if err != nil {
		logger.Error("failed to load email template: %v", zap.Error(err))
	}

	return e.mailer.SendEmail(to, subject, body)
}
