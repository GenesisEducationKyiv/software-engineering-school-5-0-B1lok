package mapper

import (
	"weather-service/internal/application/query"
	"weather-service/internal/interface/rest/weather/dto/response"
)

func ToWeatherResponse(weatherResult *query.WeatherResult) *response.WeatherResponse {
	return &response.WeatherResponse{
		Temperature: weatherResult.Temperature,
		Humidity:    weatherResult.Humidity,
		Description: weatherResult.Description,
	}
}

func ToWeatherDailyResponse(w *query.WeatherDailyResult) *response.WeatherDailyResponse {
	return &response.WeatherDailyResponse{
		Location:   w.Location,
		Date:       w.Date,
		MaxTempC:   w.MaxTempC,
		MinTempC:   w.MinTempC,
		AvgTempC:   w.AvgTempC,
		WillItRain: w.WillItRain,
		ChanceRain: w.ChanceRain,
		WillItSnow: w.WillItSnow,
		ChanceSnow: w.ChanceSnow,
		Condition:  w.Condition,
		Icon:       w.Icon,
	}
}

func ToWeatherHourlyResponse(w *query.WeatherHourlyResult) *response.WeatherHourlyResponse {
	return &response.WeatherHourlyResponse{
		Location:   w.Location,
		Time:       w.Time,
		TempC:      w.TempC,
		WillItRain: w.WillItRain,
		ChanceRain: w.ChanceRain,
		WillItSnow: w.WillItSnow,
		ChanceSnow: w.ChanceSnow,
		Condition:  w.Condition,
		Icon:       w.Icon,
	}
}
