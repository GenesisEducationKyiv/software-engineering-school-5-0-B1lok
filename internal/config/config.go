package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	DB           DBConfig      `mapstructure:"DB"`
	Server       ServerConfig  `mapstructure:"SERVER"`
	Weather      WeatherConfig `mapstructure:"WEATHER"`
	Email        EmailConfig   `mapstructure:"EMAIL"`
	OpenMeteoUrl string        `mapstructure:"OPEN_METEO_URL"`
	GeoCodingUrl string        `mapstructure:"GEO_CODING_URL"`
	Redis        RedisConfig   `mapstructure:"REDIS"`
}

type DBConfig struct {
	Host     string `mapstructure:"HOST"`
	Port     string `mapstructure:"PORT"`
	User     string `mapstructure:"USER"`
	Password string `mapstructure:"PASSWORD"`
	Name     string `mapstructure:"NAME"`
}

type ServerConfig struct {
	Host string `mapstructure:"HOST"`
	Port string `mapstructure:"PORT"`
}

type WeatherConfig struct {
	ApiUrl string `mapstructure:"API_URL"`
	ApiKey string `mapstructure:"API_KEY"`
}

type EmailConfig struct {
	Host     string `mapstructure:"HOST"`
	Port     int    `mapstructure:"PORT"`
	Username string `mapstructure:"USERNAME"`
	Password string `mapstructure:"PASSWORD"`
	From     string `mapstructure:"FROM"`
}

type RedisConfig struct {
	Address  string `mapstructure:"ADDRESS"`
	Password string `mapstructure:"PASSWORD"`
	DB       int    `mapstructure:"DB"`
}

func LoadConfig() (Config, error) {
	viper.AutomaticEnv()
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Print("No .env file found, binding individual environment variables.")
		bindAllEnvVars()
	}

	bindEnvKeysWithDots()

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

func bindAllEnvVars() {
	_ = viper.BindEnv("DB_HOST")
	_ = viper.BindEnv("DB_PORT")
	_ = viper.BindEnv("DB_USER")
	_ = viper.BindEnv("DB_PASSWORD")
	_ = viper.BindEnv("DB_NAME")
	_ = viper.BindEnv("SERVER_HOST")
	_ = viper.BindEnv("SERVER_PORT")
	_ = viper.BindEnv("WEATHER_API_URL")
	_ = viper.BindEnv("WEATHER_API_KEY")
	_ = viper.BindEnv("EMAIL_HOST")
	_ = viper.BindEnv("EMAIL_PORT")
	_ = viper.BindEnv("EMAIL_USERNAME")
	_ = viper.BindEnv("EMAIL_PASSWORD")
	_ = viper.BindEnv("EMAIL_FROM")
	_ = viper.BindEnv("OPEN_METEO_URL")
	_ = viper.BindEnv("GEO_CODING_URL")
	_ = viper.BindEnv("REDIS_ADDRESS")
	_ = viper.BindEnv("REDIS_PASSWORD")
	_ = viper.BindEnv("REDIS_DB")
}

func bindEnvKeysWithDots() {
	for _, key := range viper.AllKeys() {
		dottedKey := strings.ToLower(strings.ReplaceAll(key, "_", "."))
		viper.Set(dottedKey, viper.Get(key))
	}
}
