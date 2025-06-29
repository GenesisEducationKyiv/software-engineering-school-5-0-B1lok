package geo_coding_api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	appHttp "weather-api/internal/infrastructure/http"
	"weather-api/pkg/errors"
)

type Logger interface {
	LogResponse(provider string, resp *http.Response)
}

type Client struct {
	geoCodingUrl string
	client       *http.Client
	logger       Logger
}

const (
	providerName   = "geo-coding-api"
	defaultTimeout = 10 * time.Second
)

func NewClient(geoCodingUrl string, logger Logger) *Client {
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	return &Client{client: client, geoCodingUrl: geoCodingUrl, logger: logger}
}

type CityResponse struct {
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
}

func (h *Client) Validate(city string) (*string, error) {
	endpoint := fmt.Sprintf("%s/search?name=%s&count=1", h.geoCodingUrl, city)

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

	if err := h.handleAPIResponse(resp); err != nil {
		return nil, err
	}

	var apiResponse CityResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.Wrap(
			err, "failed to parse geolocation response", http.StatusInternalServerError,
		)
	}
	if len(apiResponse.Results) == 0 || !strings.EqualFold(apiResponse.Results[0].Name, city) {
		return nil, errors.New("Invalid input", http.StatusBadRequest)
	}

	return &apiResponse.Results[0].Name, nil
}

func (h *Client) handleAPIResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error  bool   `json:"error"`
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return errors.Wrap(
				err, "failed to parse error response from geo-coding API", http.StatusBadGateway,
			)
		}

		if errResp.Error {
			return errors.New(errResp.Reason, http.StatusBadRequest)
		}

		return errors.New("unexpected error from geo-coding API", http.StatusBadGateway)
	}

	return nil
}
