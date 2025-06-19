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
	"weather-api/pkg/errors"
)

type WeatherRepository struct {
	apiKey  string
	client  *http.Client
	clock   Clock
	baseUrl string
}

func NewWeatherRepository(apiUrl string, apiKey string) *WeatherRepository {
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	return &WeatherRepository{apiKey: apiKey, client: client, clock: SystemClock{}, baseUrl: apiUrl}
}

const (
	currentEndpoint  = "/current.json"
	forecastEndpoint = "/forecast.json"
	defaultTimeout   = 10 * time.Second
)

func (r *WeatherRepository) SetClock(clock Clock) {
	r.clock = clock
}

func (r *WeatherRepository) GetWeather(ctx context.Context, city string) (*domain.Weather, error) {
	endpoint := fmt.Sprintf("%s%s?key=%s&q=%s",
		r.baseUrl,
		currentEndpoint,
		r.apiKey,
		url.QueryEscape(city),
	)
	resp, err := r.requestWeatherAPI(ctx, endpoint)
	if err != nil {
		return nil, err
	}
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

func (r *WeatherRepository) GetDailyForecast(
	ctx context.Context, city string,
) (*domain.WeatherDaily, error) {
	endpoint := fmt.Sprintf("%s%s?key=%s&q=%s&days=1",
		r.baseUrl,
		forecastEndpoint,
		r.apiKey,
		url.QueryEscape(city),
	)
	resp, err := r.requestWeatherAPI(ctx, endpoint)
	if err != nil {
		return nil, err
	}
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

func (r *WeatherRepository) GetHourlyForecast(
	ctx context.Context, city string,
) (*domain.WeatherHourly, error) {
	endpoint := fmt.Sprintf("%s%s?key=%s&q=%s&days=1",
		r.baseUrl,
		forecastEndpoint,
		r.apiKey,
		url.QueryEscape(city),
	)
	resp, err := r.requestWeatherAPI(ctx, endpoint)
	if err != nil {
		return nil, err
	}
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

func (r *WeatherRepository) requestWeatherAPI(
	ctx context.Context, endpoint string,
) (*http.Response, error) {
	resp, err := r.client.Get(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to weather API", http.StatusServiceUnavailable)
	}
	return resp, nil
}

func (r *WeatherRepository) handleAPIResponse(resp *http.Response) error {
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
