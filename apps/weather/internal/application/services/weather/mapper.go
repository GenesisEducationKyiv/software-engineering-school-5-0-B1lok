package weather

import (
	"weather-service/internal/application/query"
	"weather-service/internal/domain"
)

func toNewWeatherResult(weather *domain.Weather) *query.WeatherResult {
	return &query.WeatherResult{
		Temperature: weather.Temperature,
		Humidity:    weather.Humidity,
		Description: weather.Description,
	}
}

func toNewWeatherDailyResult(weather *domain.WeatherDaily) *query.WeatherDailyResult {
	return &query.WeatherDailyResult{
		Location:   weather.Location,
		Date:       weather.Date,
		MaxTempC:   weather.MaxTempC,
		MinTempC:   weather.MinTempC,
		AvgTempC:   weather.AvgTempC,
		WillItRain: weather.WillItRain,
		ChanceRain: weather.ChanceRain,
		WillItSnow: weather.WillItSnow,
		ChanceSnow: weather.ChanceSnow,
		Condition:  weather.Condition,
		Icon:       weather.Icon,
	}
}

func toNewWeatherHourlyResult(weather *domain.WeatherHourly) *query.WeatherHourlyResult {
	return &query.WeatherHourlyResult{
		Location:   weather.Location,
		Time:       weather.Time,
		TempC:      weather.TempC,
		WillItRain: weather.WillItRain,
		ChanceRain: weather.ChanceRain,
		WillItSnow: weather.WillItSnow,
		ChanceSnow: weather.ChanceSnow,
		Condition:  weather.Condition,
		Icon:       weather.Icon,
	}
}
