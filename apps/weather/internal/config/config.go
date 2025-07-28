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
	Server       ServerConfig     `config:"server"`
	Weather      WeatherConfig    `config:"weather"`
	OpenMeteoURL string           `config:"open_meteo_url"`
	GeoCodingURL string           `config:"geo_coding_url"`
	Redis        RedisConfig      `config:"redis"`
	MetricsPort  string           `config:"metrics_port"`
	LogSampling  LogSamplingRates `config:"log_sampling"`
}

type ServerConfig struct {
	HttpPort string `config:"http_port"`
	GrpcPort string `config:"grpc_port"`
}

type WeatherConfig struct {
	ApiURL string `config:"api_url"`
	ApiKey string `config:"api_key"`
}

type RedisConfig struct {
	Address  string `config:"address"`
	Password string `config:"password"`
	DB       int    `config:"db"`
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
		log.Warn().Err(err).Msg("failed to load .env file, using default values")
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
