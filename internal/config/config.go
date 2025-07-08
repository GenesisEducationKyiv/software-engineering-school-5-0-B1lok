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
	DB           DBConfig      `config:"db"`
	Server       ServerConfig  `config:"server"`
	Weather      WeatherConfig `config:"weather"`
	Email        EmailConfig   `config:"email"`
	OpenMeteoURL string        `config:"open_meteo_url"`
	GeoCodingURL string        `config:"geo_coding_url"`
	Redis        RedisConfig   `config:"redis"`
}

type DBConfig struct {
	Host     string `config:"host"`
	Port     string `config:"port"`
	User     string `config:"user"`
	Password string `config:"password"`
	Name     string `config:"name"`
}

type ServerConfig struct {
	Host string `config:"host"`
	Port string `config:"port"`
}

type WeatherConfig struct {
	ApiURL string `config:"api_url"`
	ApiKey string `config:"api_key"`
}

type EmailConfig struct {
	Host     string `config:"host"`
	Port     int    `config:"port"`
	Username string `config:"username"`
	Password string `config:"password"`
	From     string `config:"from"`
}

type RedisConfig struct {
	Address  string `config:"address"`
	Password string `config:"password"`
	DB       int    `config:"db"`
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
