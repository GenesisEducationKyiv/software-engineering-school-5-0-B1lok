package publisher

type BaseEmailMessage struct {
	To           string `json:"to"`
	TemplateName string `json:"template_name"`
}

type ConfirmationEmailMessage struct {
	BaseEmailMessage
	TemplateData ConfirmationEmailTemplate `json:"template_data"`
}

type ConfirmationEmailTemplate struct {
	City      string `json:"city"`
	Frequency string `json:"frequency"`
	URL       string `json:"url"`
}

type WeatherTemplateData struct {
	UnsubscribeURL string      `json:"unsubscribe_url"`
	Frequency      string      `json:"frequency"`
	Weather        interface{} `json:"weather"`
}

type HourlyUpdateMessage struct {
	BaseEmailMessage
	TemplateData WeatherTemplateData `json:"template_data"`
}

type HourlyUpdateTemplate struct {
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

type DailyUpdateMessage struct {
	BaseEmailMessage
	TemplateData WeatherTemplateData `json:"template_data"`
}

type DailyUpdateTemplate struct {
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
