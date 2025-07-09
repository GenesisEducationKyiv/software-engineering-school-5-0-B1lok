package weather_api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"weather-api/internal/domain"
	internalErrors "weather-api/internal/errors"
	appHttp "weather-api/internal/infrastructure/http"
	pkgErrors "weather-api/pkg/errors"
)

type Logger interface {
	LogResponse(provider string, resp *http.Response)
}

type Clock interface {
	Now() time.Time
}

type Client struct {
	apiKey  string
	client  *http.Client
	clock   Clock
	baseURL string
	logger  Logger
}

func NewClient(apiURL string, apiKey string, logger Logger, clock Clock) *Client {
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	return &Client{apiKey: apiKey,
		client:  client,
		clock:   clock,
		baseURL: apiURL,
		logger:  logger,
	}
}

const (
	providerName     = "weather-api"
	currentEndpoint  = "/current.json"
	forecastEndpoint = "/forecast.json"
	defaultTimeout   = 10 * time.Second
)

func (c *Client) GetWeather(ctx context.Context, city string) (*domain.Weather, error) {
	endpoint := fmt.Sprintf("%s%s?key=%s&q=%s",
		c.baseURL,
		currentEndpoint,
		c.apiKey,
		url.QueryEscape(city),
	)
	resp, err := appHttp.Get(ctx, c.client, endpoint)
	if err != nil {
		return nil, err
	}
	c.logger.LogResponse(providerName, resp)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("Error closing response body:", err)
		}
	}()

	if err := c.handleAPIResponse(resp); err != nil {
		return nil, err
	}
	var apiResponse WeatherRepositoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to parse weather data",
		)
	}
	return toWeather(&apiResponse), nil
}

func (c *Client) GetDailyForecast(ctx context.Context, city string) (*domain.WeatherDaily, error) {
	endpoint := fmt.Sprintf("%s%s?key=%s&q=%s&days=1",
		c.baseURL,
		forecastEndpoint,
		c.apiKey,
		url.QueryEscape(city),
	)
	resp, err := appHttp.Get(ctx, c.client, endpoint)
	if err != nil {
		return nil, err
	}
	c.logger.LogResponse(providerName, resp)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("Error closing response body:", err)
		}
	}()

	if err := c.handleAPIResponse(resp); err != nil {
		return nil, err
	}

	var apiResponse WeatherDailyResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to parse weather data",
		)
	}
	return toWeatherDaily(&apiResponse), nil
}

func (c *Client) GetHourlyForecast(
	ctx context.Context,
	city string,
) (*domain.WeatherHourly, error) {
	endpoint := fmt.Sprintf("%s%s?key=%s&q=%s&days=1",
		c.baseURL,
		forecastEndpoint,
		c.apiKey,
		url.QueryEscape(city),
	)
	resp, err := appHttp.Get(ctx, c.client, endpoint)
	if err != nil {
		return nil, err
	}
	c.logger.LogResponse(providerName, resp)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("Error closing response body:", err)
		}
	}()

	if err := c.handleAPIResponse(resp); err != nil {
		return nil, err
	}

	var apiResponse WeatherHourlyResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to parse weather data",
		)
	}
	return toWeatherHourly(&apiResponse, c.clock.Now()), nil
}

func (c *Client) handleAPIResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return pkgErrors.New(
				internalErrors.ErrServiceUnavailable,
				"failed to parse error response from weather API",
			)
		}

		if errResp.Error.Code == 1006 {
			return pkgErrors.New(internalErrors.ErrNotFound, "City not found")
		}

		return pkgErrors.New(
			internalErrors.ErrServiceUnavailable,
			fmt.Sprintf("weather API error: %s", errResp.Error.Message),
		)
	}

	return nil
}
