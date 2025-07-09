package email

import (
	"weather-api/internal/config"
)

func CreateConfig(cfg config.EmailConfig) EmailConfig {
	return EmailConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Username: cfg.Username,
		Password: cfg.Password,
		From:     cfg.From,
	}
}
