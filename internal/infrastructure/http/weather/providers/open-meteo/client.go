package open_meteo

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"weather-api/internal/domain"
	appHttp "weather-api/internal/infrastructure/http"
	"weather-api/pkg/errors"
)

type Logger interface {
	LogResponse(provider string, resp *http.Response)
}

type Client struct {
	client       *http.Client
	clock        appHttp.Clock
	openMeteoUrl string
	geoCodingUrl string
	logger       Logger
}

const (
	providerName   = "open-meteo"
	defaultTimeout = 10 * time.Second
	current        = "current"
	daily          = "daily"
	hourly         = "hourly"
)

func NewClient(openMeteoUrl string, geoCodingUrl string, logger Logger) *Client {
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	return &Client{
		client:       client,
		openMeteoUrl: openMeteoUrl,
		clock:        appHttp.SystemClock{},
		geoCodingUrl: geoCodingUrl,
		logger:       logger,
	}
}

func (h *Client) SetClock(clock appHttp.Clock) {
	h.clock = clock
}

func (h *Client) GetWeather(city string) (*domain.Weather, error) {
	coords, err := h.fetchCoordinates(city)
	if err != nil {
		return nil, err
	}

	resp, err := appHttp.Get(h.client, h.buildRequestUrl(coords, currentWeatherParams, current))
	h.logger.LogResponse(providerName, resp)
	if err != nil {
		return nil, errors.Wrap(
			err, "failed to connect to weather API", http.StatusServiceUnavailable,
		)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("Error closing response body:", err)
		}
	}()

	if err := h.handleAPIResponse(resp); err != nil {
		return nil, err
	}

	var apiResponse WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(
			err, "failed to parse weather response", http.StatusInternalServerError,
		)
	}
	return toWeather(&apiResponse), nil
}

func (h *Client) GetDailyForecast(city string) (*domain.WeatherDaily, error) {
	coords, err := h.fetchCoordinates(city)
	if err != nil {
		return nil, err
	}

	resp, err := appHttp.Get(h.client, h.buildRequestUrl(coords, dailyForecastParams, daily))
	h.logger.LogResponse(providerName, resp)
	if err != nil {
		return nil, errors.Wrap(
			err, "failed to connect to weather API", http.StatusServiceUnavailable,
		)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("Error closing response body:", err)
		}
	}()

	if err := h.handleAPIResponse(resp); err != nil {
		return nil, err
	}

	var apiResponse WeatherDailyResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(
			err, "failed to parse weather response", http.StatusInternalServerError,
		)
	}
	return toWeatherDaily(&apiResponse, city), nil
}

func (h *Client) GetHourlyForecast(city string) (*domain.WeatherHourly, error) {
	coords, err := h.fetchCoordinates(city)
	if err != nil {
		return nil, err
	}

	resp, err := appHttp.Get(h.client, h.buildRequestUrl(coords, hourlyForecastParams, hourly))
	h.logger.LogResponse(providerName, resp)
	if err != nil {
		return nil, errors.Wrap(
			err, "failed to connect to weather API", http.StatusServiceUnavailable,
		)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("Error closing response body:", err)
		}
	}()

	if err := h.handleAPIResponse(resp); err != nil {
		return nil, err
	}

	var apiResponse WeatherHourlyResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(
			err, "failed to parse weather response", http.StatusInternalServerError,
		)
	}
	return toWeatherHourly(&apiResponse, city, h.clock.Now()), nil
}

type coordinates struct {
	latitude  float64
	longitude float64
}

func (h *Client) fetchCoordinates(city string) (*coordinates, error) {
	endpoint := fmt.Sprintf("%s/search?name=%s&count=1", h.geoCodingUrl, city)

	resp, err := appHttp.Get(h.client, endpoint)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("Error closing response body:", err)
		}
	}()

	if err := h.handleAPIResponse(resp); err != nil {
		return nil, err
	}

	var apiResponse GeolocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(
			err, "failed to parse geolocation response", http.StatusInternalServerError,
		)
	}
	if len(apiResponse.Results) == 0 {
		return nil, errors.New("city not found", http.StatusNotFound)
	}

	return &coordinates{
		latitude:  apiResponse.Results[0].Latitude,
		longitude: apiResponse.Results[0].Longitude,
	}, nil
}

func (h *Client) handleAPIResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error  bool   `json:"error"`
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return errors.Wrap(
				err, "failed to parse error response from open-meteo API", http.StatusBadGateway,
			)
		}

		if errResp.Error {
			return errors.New(errResp.Reason, http.StatusBadRequest)
		}

		return errors.New("unexpected error from open-meteo API", http.StatusBadGateway)
	}

	return nil
}

func (h *Client) buildRequestUrl(coords *coordinates, params []string, forecast string) string {
	baseURL := fmt.Sprintf("%s/forecast", h.openMeteoUrl)

	values := url.Values{}
	values.Set("latitude", fmt.Sprintf("%f", coords.latitude))
	values.Set("longitude", fmt.Sprintf("%f", coords.longitude))
	values.Set(forecast, strings.Join(params, ","))
	if forecast != current {
		values.Set("forecast_days", "1")
	}

	return fmt.Sprintf("%s?%s", baseURL, values.Encode())
}
