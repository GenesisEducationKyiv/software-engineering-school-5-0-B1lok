package response

type WeatherResponse struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Description string  `json:"description"`
}

type WeatherDailyResponse struct {
	Location   string  `json:"location"`
	Date       string  `json:"date"`
	MaxTempC   float64 `json:"max_temp_c"`
	MinTempC   float64 `json:"min_temp_c"`
	AvgTempC   float64 `json:"avg_temp_c"`
	WillItRain bool    `json:"will_it_rain"`
	ChanceRain int     `json:"chance_rain"`
	WillItSnow bool    `json:"will_it_snow"`
	ChanceSnow int     `json:"chance_snow"`
	Condition  string  `json:"condition"`
	Icon       string  `json:"icon"`
}

type WeatherHourlyResponse struct {
	Location   string  `json:"location"`
	Time       string  `json:"time"`
	TempC      float64 `json:"temp_c"`
	WillItRain bool    `json:"will_it_rain"`
	ChanceRain int     `json:"chance_rain"`
	WillItSnow bool    `json:"will_it_snow"`
	ChanceSnow int     `json:"chance_snow"`
	Condition  string  `json:"condition"`
	Icon       string  `json:"icon"`
}
