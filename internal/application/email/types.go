package email

import (
	"weather-api/internal/domain"
)

type ConfirmationEmail struct {
	To        string
	City      string
	Frequency string
	URL       string
}

type WeatherDailyEmail struct {
	To             string
	Frequency      string
	UnsubscribeURL string
	WeatherDaily   *domain.WeatherDaily
}

type WeatherHourlyEmail struct {
	To             string
	Frequency      string
	UnsubscribeURL string
	WeatherHourly  *domain.WeatherHourly
}
