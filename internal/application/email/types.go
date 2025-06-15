package email

import (
	"weather-api/internal/domain"
)

type ConfirmationEmail struct {
	To        string
	City      string
	Frequency string
	Url       string
}

type WeatherDailyEmail struct {
	To             string
	Frequency      string
	UnsubscribeUrl string
	WeatherDaily   *domain.WeatherDaily
}

type WeatherHourlyEmail struct {
	To             string
	Frequency      string
	UnsubscribeUrl string
	WeatherHourly  *domain.WeatherHourly
}
