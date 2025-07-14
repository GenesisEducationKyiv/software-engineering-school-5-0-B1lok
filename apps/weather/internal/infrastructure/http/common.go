package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	internalErrors "weather-service/internal/errors"
	pkgErrors "weather-service/pkg/errors"
)

func Get(ctx context.Context, client *http.Client, endpoint string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, pkgErrors.New(internalErrors.ErrInternal, "failed to create request to API")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, pkgErrors.New(internalErrors.ErrInternal, "failed to connect to API")
	}
	return resp, nil
}

func MockHTTPClient(response MockResponse) *http.Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(response.StatusCode)
		_, _ = w.Write([]byte(response.Body))
	}))

	return &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(server.URL)
			},
		},
	}
}

func MockHTTPClientWithResponses(responses map[string]MockResponse) *http.Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestURL := r.URL.String()

		for urlPattern, response := range responses {
			if strings.Contains(requestURL, urlPattern) {
				w.WriteHeader(response.StatusCode)
				_, _ = w.Write([]byte(response.Body))
				return
			}
		}
	}))

	return &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(server.URL)
			},
		},
	}
}

type MockResponse struct {
	Body       string
	StatusCode int
}

type NoOpLogger struct{}

func (NoOpLogger) LogResponse(_ string, _ *http.Response) {}
