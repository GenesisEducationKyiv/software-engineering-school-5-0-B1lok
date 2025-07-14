package config

import (
	"fmt"
	"log"

	"github.com/go-viper/mapstructure/v2"
	"github.com/iamolegga/enviper"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const (
	defaultEnvFile = ".env"
	tagName        = "config"
)

type Config struct {
	Email       EmailConfig `config:"email"`
	ServerPort  string      `config:"server_port"`
	RabbitMqURL string      `config:"rabbitmq_url"`
}

type EmailConfig struct {
	Host     string `config:"host"`
	Port     int    `config:"port"`
	Username string `config:"username"`
	Password string `config:"password"`
	From     string `config:"from"`
}

func LoadConfig() (Config, error) {
	if err := loadEnvFile(defaultEnvFile); err != nil {
		log.Printf("warning: %v", err)
	}

	var config Config
	if err := readConfig(&config); err != nil {
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	return config, nil
}

func loadEnvFile(path string) error {
	if err := godotenv.Load(path); err != nil {
		return fmt.Errorf("error loading .env file from %s: %w", path, err)
	}

	return nil
}

func readConfig(config *Config) error {
	v := enviper.New(viper.GetViper()).WithTagName(tagName)

	confOption := func(c *mapstructure.DecoderConfig) {
		c.Squash = true
	}

	if err := v.Unmarshal(config, confOption); err != nil {
		return fmt.Errorf("unmarshaling: %w", err)
	}

	return nil
}
