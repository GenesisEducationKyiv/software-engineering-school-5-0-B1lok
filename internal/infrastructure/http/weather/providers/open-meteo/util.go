package open_meteo

import "fmt"

var currentWeatherParams = []string{"temperature_2m", "relative_humidity_2m", "weather_code"}
var dailyForecastParams = []string{"weather_code", "temperature_2m_max",
	"temperature_2m_min", "rain_sum", "snowfall_sum",
}
var hourlyForecastParams = []string{"temperature_2m", "rain", "snowfall", "weather_code"}

var weatherCodeDescriptions = map[int]string{
	0:  "Clear sky",
	1:  "Mainly clear",
	2:  "Partly cloudy",
	3:  "Overcast",
	45: "Fog",
	48: "Depositing rime fog",

	51: "Drizzle: Light",
	53: "Drizzle: Moderate",
	55: "Drizzle: Dense",
	56: "Freezing Drizzle: Light",
	57: "Freezing Drizzle: Dense",

	61: "Rain: Slight",
	63: "Rain: Moderate",
	65: "Rain: Heavy",
	66: "Freezing Rain: Light",
	67: "Freezing Rain: Heavy",

	71: "Snow fall: Slight",
	73: "Snow fall: Moderate",
	75: "Snow fall: Heavy",
	77: "Snow grains",

	80: "Rain showers: Slight",
	81: "Rain showers: Moderate",
	82: "Rain showers: Violent",

	85: "Snow showers: Slight",
	86: "Snow showers: Heavy",

	95: "Thunderstorm: Slight or moderate",
	96: "Thunderstorm with slight hail",
	99: "Thunderstorm with heavy hail",
}

const iconSnow = "13d"

func getWeatherIconURL(weatherCode int) string { //nolint:gocyclo
	var iconCode string

	switch weatherCode {
	case 0:
		iconCode = "01d"
	case 1:
		iconCode = "02d"
	case 2:
		iconCode = "03d"
	case 3:
		iconCode = "04d"
	case 45, 48:
		iconCode = "50d"
	case 51, 53, 55, 56, 57:
		iconCode = "09d"
	case 61, 63, 65:
		iconCode = "10d"
	case 66, 67:
		iconCode = iconSnow
	case 71, 73, 75:
		iconCode = iconSnow
	case 77:
		iconCode = iconSnow
	case 80, 81, 82:
		iconCode = "09d"
	case 85, 86:
		iconCode = iconSnow
	case 95, 96, 99:
		iconCode = "11d"
	default:
		iconCode = "01d"
	}

	return fmt.Sprintf("http://openweathermap.org/img/wn/%s@2x.png", iconCode)
}
