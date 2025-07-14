package weather

func mapCurrentWeather(w *Weather) currentResponse {
	return currentResponse{
		Temperature: w.Temperature,
		Humidity:    w.Humidity,
		Description: w.Description,
	}
}

func mapDailyWeather(w *WeatherDaily) dailyResponse {
	return dailyResponse{
		Location:   w.Location,
		Date:       w.Date,
		MaxTempC:   w.MaxTempC,
		MinTempC:   w.MinTempC,
		AvgTempC:   w.AvgTempC,
		WillItRain: w.WillItRain,
		ChanceRain: int(w.ChanceRain),
		WillItSnow: w.WillItSnow,
		ChanceSnow: int(w.ChanceSnow),
		Condition:  w.Condition,
		Icon:       w.Icon,
	}
}

func mapHourlyWeather(w *WeatherHourly) hourlyResponse {
	return hourlyResponse{
		Location:   w.Location,
		Time:       w.Time,
		TempC:      w.TempC,
		WillItRain: w.WillItRain,
		ChanceRain: int(w.ChanceRain),
		WillItSnow: w.WillItSnow,
		ChanceSnow: int(w.ChanceSnow),
		Condition:  w.Condition,
		Icon:       w.Icon,
	}
}
