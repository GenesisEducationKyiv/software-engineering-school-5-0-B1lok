package ttl

import (
	"time"

	"weather-service/internal/infrastructure/db/redis/weather"
)

type Clock interface {
	Now() time.Time
}

type Provider struct {
	clock      Clock
	currentTTL time.Duration
}

func NewTTLProvider(currentTTL time.Duration, clock Clock) *Provider {
	return &Provider{
		currentTTL: currentTTL,
		clock:      clock,
	}
}

func (p *Provider) TTL(forecastType weather.ForecastType) time.Duration {
	switch forecastType {
	case weather.ForecastCurrent:
		return p.currentTTL
	case weather.ForecastHourly:
		endOfHour := p.clock.Now().Truncate(time.Hour).Add(time.Hour)
		return time.Until(endOfHour)
	case weather.ForecastDaily:
		tomorrow := p.clock.Now().AddDate(0, 0, 1)
		midnight := time.Date(
			tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
			0, 0, 0, 0, p.clock.Now().Location(),
		)
		return time.Until(midnight)
	default:
		return 1 * time.Minute
	}
}
