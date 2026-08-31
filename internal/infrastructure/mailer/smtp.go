package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"go-gin-clean/internal/application/port"
	"go-gin-clean/pkg/config"

	"gopkg.in/gomail.v2"
)

//go:embed templates/*.html
var templateFS embed.FS

type EmailDialer interface {
	DialAndSend(m ...*gomail.Message) error
}

type smtpMailer struct {
	cfg    *config.EmailConfig
	dialer EmailDialer
}

func NewSMTPMailer(cfg *config.EmailConfig) port.Mailer {
	dialer := gomail.NewDialer(
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
	)
	return &smtpMailer{
		cfg:    cfg,
		dialer: dialer,
	}
}

func (s *smtpMailer) SendEmail(to, subject, body string) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", s.cfg.From)
	mailer.SetHeader("To", to)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/html", body)

	if err := s.dialer.DialAndSend(mailer); err != nil {
		return err
	}
	return nil
}

func (s *smtpMailer) LoadTemplate(templateName string, data any) (string, error) {
	templatePath := fmt.Sprintf("templates/email/%s.html", templateName)

	templateContent, err := templateFS.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	tmpl, err := template.New(templateName).Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var renderedTemplate bytes.Buffer
	err = tmpl.Execute(&renderedTemplate, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return renderedTemplate.String(), nil
}
