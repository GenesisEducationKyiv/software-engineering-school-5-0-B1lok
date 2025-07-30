package config

import (
	"fmt"

	"github.com/rs/zerolog/log"

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
	DB               DBConfig         `config:"db"`
	Server           ServerConfig     `config:"server"`
	ValidatorAddress string           `config:"validator_address"`
	RabbitMqURL      string           `config:"rabbitmq_url"`
	MetricsPort      string           `config:"metrics_port"`
	LogSampling      LogSamplingRates `config:"log_sampling"`
	LokiHost         string           `config:"loki_host"`
}

type DBConfig struct {
	Host     string `config:"host"`
	Port     string `config:"port"`
	User     string `config:"user"`
	Password string `config:"password"`
	Name     string `config:"name"`
}

type ServerConfig struct {
	Host     string `config:"host"`
	HttpPort string `config:"http_port"`
	GrpcPort string `config:"grpc_port"`
}

type LogSamplingRates struct {
	Enabled bool    `config:"enabled"`
	Trace   float64 `config:"trace"`
	Debug   float64 `config:"debug"`
	Info    float64 `config:"info"`
	Warn    float64 `config:"warn"`
	Error   float64 `config:"error"`
}

func LoadConfig() (Config, error) {
	if err := loadEnvFile(defaultEnvFile); err != nil {
		log.Warn().Err(err).Msgf("loading config")
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
