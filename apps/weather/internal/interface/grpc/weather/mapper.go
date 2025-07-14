package weather

import (
	"weather-service/internal/application/query"
)

func toProtoWeather(w *query.WeatherResult) *Weather {
	return &Weather{
		Temperature: w.Temperature,
		Humidity:    w.Humidity,
		Description: w.Description,
	}
}

//nolint:gosec
func toProtoWeatherDaily(w *query.WeatherDailyResult) *WeatherDaily {
	return &WeatherDaily{
		Location:   w.Location,
		Date:       w.Date,
		MaxTempC:   w.MaxTempC,
		MinTempC:   w.MinTempC,
		AvgTempC:   w.AvgTempC,
		WillItRain: w.WillItRain,
		ChanceRain: int32(w.ChanceRain),
		WillItSnow: w.WillItSnow,
		ChanceSnow: int32(w.ChanceSnow),
		Condition:  w.Condition,
		Icon:       w.Icon,
	}
}

//nolint:gosec
func toProtoWeatherHourly(w *query.WeatherHourlyResult) *WeatherHourly {
	return &WeatherHourly{
		Location:   w.Location,
		Time:       w.Time,
		TempC:      w.TempC,
		WillItRain: w.WillItRain,
		ChanceRain: int32(w.ChanceRain),
		WillItSnow: w.WillItSnow,
		ChanceSnow: int32(w.ChanceSnow),
		Condition:  w.Condition,
		Icon:       w.Icon,
	}
}
