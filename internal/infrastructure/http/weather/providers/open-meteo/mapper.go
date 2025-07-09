package open_meteo

import (
	"time"

	"weather-api/internal/domain"
)

const (
	timeLayout = "2006-01-02T15:04"
)

func toWeather(weatherResponse *WeatherResponse) *domain.Weather {
	return &domain.Weather{
		Temperature: weatherResponse.Current.Temperature,
		Humidity:    float64(weatherResponse.Current.Humidity),
		Description: weatherCodeDescriptions[weatherResponse.Current.WeatherCode],
	}
}

func toWeatherDaily(response *WeatherDailyResponse, city string) *domain.WeatherDaily {
	chanceRain := 0
	if response.Daily.RainSum[0] > 1 {
		chanceRain = 100
	}

	chanceSnow := 0
	if response.Daily.SnowfallSum[0] > 0 {
		chanceSnow = 100
	}
	return &domain.WeatherDaily{
		Location:   city,
		Date:       response.Daily.Time[0],
		MaxTempC:   response.Daily.TemperatureMax[0],
		MinTempC:   response.Daily.TemperatureMin[0],
		AvgTempC:   (response.Daily.TemperatureMax[0] + response.Daily.TemperatureMin[0]) / 2,
		WillItRain: response.Daily.RainSum[0] > 0,
		ChanceRain: chanceRain,
		WillItSnow: response.Daily.SnowfallSum[0] > 0,
		ChanceSnow: chanceSnow,
		Condition:  weatherCodeDescriptions[response.Daily.WeatherCode[0]],
		Icon:       getWeatherIconURL(response.Daily.WeatherCode[0]),
	}
}

func toWeatherHourly(
	response *WeatherHourlyResponse,
	city string, targetTime time.Time,
) *domain.WeatherHourly {
	currentTime := targetTime.Truncate(time.Hour).Format(timeLayout)
	for index, hour := range response.Hourly.Time {
		if hour == currentTime {
			chanceRain := 0
			if response.Hourly.Rain[index] > 1 {
				chanceRain = 100
			}

			chanceSnow := 0
			if response.Hourly.Snowfall[index] > 0 {
				chanceSnow = 100
			}
			return &domain.WeatherHourly{
				Location:   city,
				Time:       hour,
				TempC:      response.Hourly.Temperature[index],
				WillItRain: response.Hourly.Rain[index] > 0,
				ChanceRain: chanceRain,
				WillItSnow: response.Hourly.Snowfall[index] > 0,
				ChanceSnow: chanceSnow,
				Condition:  weatherCodeDescriptions[response.Hourly.WeatherCode[index]],
				Icon:       getWeatherIconURL(response.Hourly.WeatherCode[index]),
			}
		}
	}

	return nil
}
