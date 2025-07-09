package validator

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"weather-api/pkg/errors"
)

type CityValidator struct {
	apiUrl string
	apiKey string
	client *http.Client
}

func NewCityValidator(apiUrl, apiKey string) *CityValidator {
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	return &CityValidator{apiUrl: apiUrl, apiKey: apiKey, client: client}
}

const (
	searchEndpoint = "/search.json"
	defaultTimeout = 10 * time.Second
)

func (c CityValidator) Validate(city string) (*string, error) {
	endpoint := fmt.Sprintf("%s%s?key=%s&q=%s",
		c.apiUrl,
		searchEndpoint,
		c.apiKey,
		url.QueryEscape(city),
	)
	resp, err := c.client.Get(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to weather API", http.StatusServiceUnavailable)
	}
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

	if len(results) == 0 {
		return nil, errors.New("Invalid input", http.StatusBadRequest)
	}

	return &results[0].Name, nil
}

type cityResult struct {
	Name string `json:"name"`
}
