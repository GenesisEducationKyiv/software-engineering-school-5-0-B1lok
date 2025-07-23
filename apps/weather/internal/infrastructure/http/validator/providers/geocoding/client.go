package geocoding

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	internalErrors "weather-service/internal/errors"
	appHttp "weather-service/internal/infrastructure/http"
	pkgErrors "weather-service/pkg/errors"
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
	providerName   = "geocoding"
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

func (h *Client) Validate(ctx context.Context, city string) (*string, error) {
	endpoint := fmt.Sprintf("%s/search?name=%s&count=1", h.geoCodingUrl, city)

	resp, err := appHttp.Get(ctx, h.client, endpoint)
	if err != nil {
		return nil, err
	}
	h.logger.LogResponse(providerName, resp)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close response body")
		}
	}()

	if err := h.handleAPIResponse(resp); err != nil {
		return nil, err
	}

	var apiResponse CityResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, pkgErrors.New(
			internalErrors.ErrInternal, "failed to parse geolocation response",
		)
	}
	if len(apiResponse.Results) == 0 || !strings.EqualFold(apiResponse.Results[0].Name, city) {
		return nil, pkgErrors.New(internalErrors.ErrInvalidInput, "Invalid city input")
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
			return pkgErrors.New(
				internalErrors.ErrServiceUnavailable,
				"failed to parse error response from geo-coding API",
			)
		}

		if errResp.Error {
			return pkgErrors.New(internalErrors.ErrInvalidInput, errResp.Reason)
		}

		return pkgErrors.New(
			internalErrors.ErrServiceUnavailable, "unexpected error from geo-coding API",
		)
	}

	return nil
}
