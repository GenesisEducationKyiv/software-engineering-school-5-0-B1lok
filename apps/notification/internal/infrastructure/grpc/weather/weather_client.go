package weather

import (
	"context"
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
) (*WeatherDaily, error) {
	return c.client.GetDailyWeather(ctx, &CityRequest{City: city})
}

func (c *Client) HourlyUpdate(
	ctx context.Context,
	city string,
) (*WeatherHourly, error) {
	return c.client.GetHourlyWeather(ctx, &CityRequest{City: city})

}
