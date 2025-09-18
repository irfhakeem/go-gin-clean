package usecases

import (
	"fmt"
	"go-gin-clean/internal/core/ports"
)

type EmailUseCase struct {
	application string
	smtp        ports.MailerService
}

func NewEmailUseCase(smtp ports.MailerService) ports.EmailUseCase {
	return &EmailUseCase{
		application: "Go Gin Clean App",
		smtp:        smtp,
	}
}

func (e *EmailUseCase) SendVerifyEmail(to, name, url string) error {
	subject := fmt.Sprintf("Verify User %s Email", e.application)

	data := map[string]any{
		"Name":            name,
		"VerificationURL": url,
	}

	body, err := e.smtp.LoadTemplate("verify_email", data)
	if err != nil {
		return fmt.Errorf("failed to load email template: %v", err)
	}

	return e.smtp.SendEmail(to, subject, body)
}

func (e *EmailUseCase) SendResetPasswordEmail(to, name, url string) error {
	subject := fmt.Sprintf("Reset %s Account Password", e.application)

	data := map[string]any{
		"Name":     name,
		"ResetURL": url,
	}

	body, err := e.smtp.LoadTemplate("reset_password", data)
	if err != nil {
		return fmt.Errorf("failed to load password reset request email template: %v", err)
	}

	return e.smtp.SendEmail(to, subject, body)
}
