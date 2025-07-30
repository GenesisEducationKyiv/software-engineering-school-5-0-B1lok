package config

import (
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/iamolegga/enviper"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	defaultEnvFile = ".env"
	tagName        = "config"
)

type Config struct {
	ServerPort              string `config:"server_port"`
	WeatherServiceAddr      string `config:"weather_service"`
	SubscriptionServiceAddr string `config:"subscription_service"`
}

func LoadConfig() (Config, error) {
	if err := loadEnvFile(defaultEnvFile); err != nil {
		log.Warn().Err(err).Msg("failed to load .env file")
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
