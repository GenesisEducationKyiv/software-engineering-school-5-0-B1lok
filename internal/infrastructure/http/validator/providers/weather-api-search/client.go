package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	internalErrors "weather-api/internal/errors"
	appHttp "weather-api/internal/infrastructure/http"
	pkgErrors "weather-api/pkg/errors"
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
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to validate city: non-200 response",
		)
	}

	var results []cityResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to parse city validation response",
		)
	}

	if len(results) == 0 || !strings.EqualFold(results[0].Name, city) {
		return nil, pkgErrors.New(internalErrors.ErrInvalidInput, "Invalid input")
	}

	return &results[0].Name, nil
}

type cityResult struct {
	Name string `json:"name"`
}
