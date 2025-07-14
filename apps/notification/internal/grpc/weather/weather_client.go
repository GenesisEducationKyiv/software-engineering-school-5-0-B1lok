package weather

import (
	"context"

	"notification/internal/rabbitmq/publisher"
)

type Client struct {
	client WeatherServiceClient
}

func NewClient(client WeatherServiceClient) *Client {
	return &Client{client: client}
}

func (c *Client) DailyUpdate(
	ctx context.Context,
	city string,
) (publisher.DailyUpdateTemplate, error) {
	weather, err := c.client.GetDailyWeather(ctx, &CityRequest{City: city})

	if err != nil {
		return publisher.DailyUpdateTemplate{}, err
	}

	return publisher.DailyUpdateTemplate{
		Location:   weather.Location,
		Date:       weather.Date,
		MaxTempC:   weather.MaxTempC,
		MinTempC:   weather.MinTempC,
		AvgTempC:   weather.AvgTempC,
		WillItRain: weather.WillItRain,
		ChanceRain: int(weather.ChanceRain),
		WillItSnow: weather.WillItSnow,
		ChanceSnow: int(weather.ChanceSnow),
		Condition:  weather.Condition,
		Icon:       weather.Icon,
	}, nil
}

func (c *Client) HourlyUpdate(
	ctx context.Context,
	city string,
) (publisher.HourlyUpdateTemplate, error) {
	weather, err := c.client.GetHourlyWeather(ctx, &CityRequest{City: city})

	if err != nil {
		return publisher.HourlyUpdateTemplate{}, err
	}

	return publisher.HourlyUpdateTemplate{
		Location:   weather.Location,
		Time:       weather.Time,
		TempC:      weather.TempC,
		WillItRain: weather.WillItRain,
		ChanceRain: int(weather.ChanceRain),
		WillItSnow: weather.WillItSnow,
		ChanceSnow: int(weather.ChanceSnow),
		Condition:  weather.Condition,
		Icon:       weather.Icon,
	}, nil
}
