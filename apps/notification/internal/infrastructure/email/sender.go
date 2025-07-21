package email

import (
	"bytes"
	"fmt"
	"gopkg.in/gomail.v2"
	"html/template"

	"notification/internal/config"
)

const (
	templatePath = "templates"
)

type Sender struct {
	config config.EmailConfig
	dialer *gomail.Dialer
}

func NewEmailSender(config config.EmailConfig) *Sender {
	dialer := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)

	return &Sender{
		config: config,
		dialer: dialer,
	}
}

func (s *Sender) Send(templateName, to, subject string, data any) error {
	htmlBody, err := renderTemplate(fmt.Sprintf("%s/%s", templatePath, templateName), data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	message := gomail.NewMessage()
	message.SetHeader("From", s.config.From)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", htmlBody)

	if err := s.dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func renderTemplate(templatePath string, data any) (string, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse file: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute: %w", err)
	}

	return buf.String(), nil
}
