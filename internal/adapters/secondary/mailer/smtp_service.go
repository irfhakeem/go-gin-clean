package mailer

import (
	"bytes"
	"fmt"
	"go-gin-clean/internal/core/ports"
	"go-gin-clean/pkg/config"
	"os"
	"path/filepath"
	"text/template"

	"gopkg.in/gomail.v2"
)

type SMTPService struct {
	cfg *config.MailerConfig
}

func NewSMTPService(cfg *config.MailerConfig) ports.MailerService {
	return &SMTPService{cfg: cfg}
}

func (s *SMTPService) SendEmail(to, subject, body string) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", s.cfg.Sender)
	mailer.SetHeader("To", to)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/html", body)

	dialer := gomail.NewDialer(
		s.cfg.Host,
		s.cfg.Port,
		s.cfg.Auth,
		s.cfg.Password,
	)

	err := dialer.DialAndSend(mailer)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func (s *SMTPService) LoadTemplate(templateName string, data any) (string, error) {
	templatePath := filepath.Join("internal", "adapters", "secondary", "mailer", "templates", templateName+".html")

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("template file not found: %s", templatePath)
	}

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %v", err)
	}

	var renderedBody string
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}
	renderedBody = buf.String()

	return renderedBody, nil
}
