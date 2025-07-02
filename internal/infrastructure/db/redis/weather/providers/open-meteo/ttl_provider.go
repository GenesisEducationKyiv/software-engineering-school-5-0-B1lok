package open_meteo

import (
	"time"

	"weather-api/internal/infrastructure/db/redis/weather"
)

type TTLProvider struct{}

func NewTTLProvider() *TTLProvider {
	return &TTLProvider{}
}

func (p *TTLProvider) TTL(forecastType weather.ForecastType) time.Duration {
	switch forecastType {
	case weather.ForecastCurrent:
		return 15 * time.Minute
	case weather.ForecastHourly:
		endOfHour := time.Now().Truncate(time.Hour).Add(time.Hour)
		return time.Until(endOfHour)
	case weather.ForecastDaily:
		tomorrow := time.Now().AddDate(0, 0, 1)
		midnight := time.Date(
			tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
			0, 0, 0, 0, time.Now().Location(),
		)
		return time.Until(midnight)
	default:
		return 1 * time.Minute
	}
}
