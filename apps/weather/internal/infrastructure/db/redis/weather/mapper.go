package weather

import (
	"weather-service/internal/domain"
)

func ToDomainWeather(dto *Weather) *domain.Weather {
	if dto == nil {
		return nil
	}
	return &domain.Weather{
		Temperature: dto.Temperature,
		Humidity:    dto.Humidity,
		Description: dto.Description,
	}
}

func ToDTOWeather(domain *domain.Weather) *Weather {
	if domain == nil {
		return nil
	}
	return &Weather{
		Temperature: domain.Temperature,
		Humidity:    domain.Humidity,
		Description: domain.Description,
	}
}

func ToDomainWeatherDaily(dto *WeatherDaily) *domain.WeatherDaily {
	if dto == nil {
		return nil
	}
	return &domain.WeatherDaily{
		Location:   dto.Location,
		Date:       dto.Date,
		MaxTempC:   dto.MaxTempC,
		MinTempC:   dto.MinTempC,
		AvgTempC:   dto.AvgTempC,
		WillItRain: dto.WillItRain,
		ChanceRain: dto.ChanceRain,
		WillItSnow: dto.WillItSnow,
		ChanceSnow: dto.ChanceSnow,
		Condition:  dto.Condition,
		Icon:       dto.Icon,
	}
}

func ToDTOWeatherDaily(domain *domain.WeatherDaily) *WeatherDaily {
	if domain == nil {
		return nil
	}
	return &WeatherDaily{
		Location:   domain.Location,
		Date:       domain.Date,
		MaxTempC:   domain.MaxTempC,
		MinTempC:   domain.MinTempC,
		AvgTempC:   domain.AvgTempC,
		WillItRain: domain.WillItRain,
		ChanceRain: domain.ChanceRain,
		WillItSnow: domain.WillItSnow,
		ChanceSnow: domain.ChanceSnow,
		Condition:  domain.Condition,
		Icon:       domain.Icon,
	}
}
func ToDomainWeatherHourly(dto *WeatherHourly) *domain.WeatherHourly {
	if dto == nil {
		return nil
	}
	return &domain.WeatherHourly{
		Location:   dto.Location,
		Time:       dto.Time,
		TempC:      dto.TempC,
		WillItRain: dto.WillItRain,
		ChanceRain: dto.ChanceRain,
		WillItSnow: dto.WillItSnow,
		ChanceSnow: dto.ChanceSnow,
		Condition:  dto.Condition,
		Icon:       dto.Icon,
	}
}

func ToDTOWeatherHourly(domain *domain.WeatherHourly) *WeatherHourly {
	if domain == nil {
		return nil
	}
	return &WeatherHourly{
		Location:   domain.Location,
		Time:       domain.Time,
		TempC:      domain.TempC,
		WillItRain: domain.WillItRain,
		ChanceRain: domain.ChanceRain,
		WillItSnow: domain.WillItSnow,
		ChanceSnow: domain.ChanceSnow,
		Condition:  domain.Condition,
		Icon:       domain.Icon,
	}
}
