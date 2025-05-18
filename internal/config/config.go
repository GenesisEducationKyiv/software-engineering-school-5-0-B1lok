package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBHost        string `mapstructure:"DB_HOST"`
	DBPort        string `mapstructure:"DB_PORT"`
	DBUser        string `mapstructure:"DB_USER"`
	DBPassword    string `mapstructure:"DB_PASSWORD"`
	DBName        string `mapstructure:"DB_NAME"`
	ServerHost    string `mapstructure:"SERVER_HOST"`
	ServerPort    string `mapstructure:"SERVER_PORT"`
	WeatherApiKey string `mapstructure:"WEATHER_API_KEY"`
	EmailHost     string `mapstructure:"EMAIL_HOST"`
	EmailPort     int    `mapstructure:"EMAIL_PORT"`
	EmailUser     string `mapstructure:"EMAIL_USERNAME"`
	EmailPassword string `mapstructure:"EMAIL_PASSWORD"`
}

func LoadConfig() (config Config, err error) {
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
