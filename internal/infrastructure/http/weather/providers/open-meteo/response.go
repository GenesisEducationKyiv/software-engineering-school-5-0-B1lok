package open_meteo

type GeolocationResponse struct {
	Results []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

type WeatherResponse struct {
	Current struct {
		Temperature float64 `json:"temperature_2m"`
		Humidity    int     `json:"relative_humidity_2m"`
		WeatherCode int     `json:"weather_code"`
	} `json:"current"`
}

type WeatherDailyResponse struct {
	Daily struct {
		Time           []string  `json:"time"`
		WeatherCode    []int     `json:"weather_code"`
		TemperatureMax []float64 `json:"temperature_2m_max"`
		TemperatureMin []float64 `json:"temperature_2m_min"`
		RainSum        []float64 `json:"rain_sum"`
		SnowfallSum    []float64 `json:"snowfall_sum"`
	} `json:"daily"`
}

type WeatherHourlyResponse struct {
	Hourly struct {
		Time        []string  `json:"time"`
		Temperature []float64 `json:"temperature_2m"`
		Rain        []float64 `json:"rain"`
		Snowfall    []float64 `json:"snowfall"`
		WeatherCode []int     `json:"weather_code"`
	} `json:"hourly"`
}
