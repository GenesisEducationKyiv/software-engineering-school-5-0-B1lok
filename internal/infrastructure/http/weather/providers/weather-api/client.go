package weather_api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"weather-api/internal/domain"
	appHttp "weather-api/internal/infrastructure/http"
	"weather-api/pkg/errors"
)

type Logger interface {
	LogResponse(provider string, resp *http.Response)
}

type Client struct {
	apiKey  string
	client  *http.Client
	clock   appHttp.Clock
	baseUrl string
	logger  Logger
}

func NewClient(apiUrl string, apiKey string, logger Logger) *Client {
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	return &Client{apiKey: apiKey,
		client:  client,
		clock:   appHttp.SystemClock{},
		baseUrl: apiUrl,
		logger:  logger,
	}
}

const (
	providerName     = "weather-api"
	currentEndpoint  = "/current.json"
	forecastEndpoint = "/forecast.json"
	defaultTimeout   = 10 * time.Second
)

func (r *Client) SetClock(clock appHttp.Clock) {
	r.clock = clock
}

func (r *Client) GetWeather(city string) (*domain.Weather, error) {
	endpoint := fmt.Sprintf("%s%s?key=%s&q=%s",
		r.baseUrl,
		currentEndpoint,
		r.apiKey,
		url.QueryEscape(city),
	)
	resp, err := appHttp.Get(r.client, endpoint)
	if err != nil {
		return nil, err
	}
	r.logger.LogResponse(providerName, resp)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("Error closing response body:", err)
		}
	}()

	if err := r.handleAPIResponse(resp); err != nil {
		return nil, err
	}
	var apiResponse WeatherRepositoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(
			err, "failed to parse weather data", http.StatusInternalServerError,
		)
	}
	return toWeather(&apiResponse), nil
}

func (r *Client) GetDailyForecast(city string) (*domain.WeatherDaily, error) {
	endpoint := fmt.Sprintf("%s%s?key=%s&q=%s&days=1",
		r.baseUrl,
		forecastEndpoint,
		r.apiKey,
		url.QueryEscape(city),
	)
	resp, err := appHttp.Get(r.client, endpoint)
	if err != nil {
		return nil, err
	}
	r.logger.LogResponse(providerName, resp)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("Error closing response body:", err)
		}
	}()

	if err := r.handleAPIResponse(resp); err != nil {
		return nil, err
	}

	var apiResponse WeatherDailyResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to parse weather data", http.StatusInternalServerError)
	}
	return toWeatherDaily(&apiResponse), nil
}

func (r *Client) GetHourlyForecast(city string) (*domain.WeatherHourly, error) {
	endpoint := fmt.Sprintf("%s%s?key=%s&q=%s&days=1",
		r.baseUrl,
		forecastEndpoint,
		r.apiKey,
		url.QueryEscape(city),
	)
	resp, err := appHttp.Get(r.client, endpoint)
	if err != nil {
		return nil, err
	}
	r.logger.LogResponse(providerName, resp)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("Error closing response body:", err)
		}
	}()

	if err := r.handleAPIResponse(resp); err != nil {
		return nil, err
	}

	var apiResponse WeatherHourlyResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(err, "failed to parse weather data", http.StatusInternalServerError)
	}
	return toWeatherHourly(&apiResponse, r.clock.Now()), nil
}

func (r *Client) handleAPIResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return errors.Wrap(err, "failed to parse error response from weather API", http.StatusBadGateway)
		}

		if errResp.Error.Code == 1006 {
			return errors.New("City not found", http.StatusNotFound)
		}

		return errors.New(fmt.Sprintf(
			"weather API error: %s", errResp.Error.Message), http.StatusBadGateway,
		)
	}

	return nil
}
