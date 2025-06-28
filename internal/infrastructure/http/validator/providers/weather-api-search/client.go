package weather_api_search

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	appHttp "weather-api/internal/infrastructure/http"
	"weather-api/pkg/errors"
)

type Logger interface {
	LogResponse(provider string, resp *http.Response)
}

type Client struct {
	apiUrl string
	apiKey string
	client *http.Client
	logger Logger
}

func NewClient(apiUrl, apiKey string, logger Logger) *Client {
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	return &Client{apiUrl: apiUrl, apiKey: apiKey, client: client, logger: logger}
}

const (
	providerName   = "weather-api-search"
	searchEndpoint = "/search.json"
	defaultTimeout = 10 * time.Second
)

func (h *Client) Validate(city string) (*string, error) {
	endpoint := fmt.Sprintf("%s%s?key=%s&q=%s",
		h.apiUrl,
		searchEndpoint,
		h.apiKey,
		url.QueryEscape(city),
	)
	resp, err := appHttp.Get(h.client, endpoint)
	if err != nil {
		return nil, err
	}
	h.logger.LogResponse(providerName, resp)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("Error closing response body:", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to validate city: non-200 response",
			http.StatusInternalServerError,
		)
	}

	var results []cityResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, errors.Wrap(err, "failed to parse city validation response",
			http.StatusInternalServerError,
		)
	}

	if len(results) == 0 || !strings.EqualFold(results[0].Name, city) {
		return nil, errors.New("Invalid input", http.StatusBadRequest)
	}

	return &results[0].Name, nil
}

type cityResult struct {
	Name string `json:"name"`
}
