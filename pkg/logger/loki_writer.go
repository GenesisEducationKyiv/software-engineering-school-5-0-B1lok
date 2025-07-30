package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	clientTimeout    = 5 * time.Second
	lokiPushEndpoint = "/loki/api/v1/push"
)

type LokiWriter struct {
	client  *http.Client
	host    string
	labels  map[string]string
	onError func(error)
	mu      sync.Mutex
}

type lokiPayload struct {
	Streams []Stream `json:"streams"`
}

type Stream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

func NewLokiWriter(host string, labels map[string]string, onError func(err error)) *LokiWriter {
	return &LokiWriter{
		client:  &http.Client{Timeout: clientTimeout},
		host:    host,
		labels:  labels,
		onError: onError,
	}
}

func (l *LokiWriter) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var logEntry map[string]interface{}
	if err := json.Unmarshal(p, &logEntry); err != nil {
		if l.onError != nil {
			l.onError(fmt.Errorf("parse error: %w", err))
		}
		return len(p), nil
	}

	timestamp := time.Now()
	if s, ok := logEntry["time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			timestamp = t
		}
	}

	line, err := json.Marshal(logEntry)
	if err != nil {
		if l.onError != nil {
			l.onError(fmt.Errorf("marshal error: %w", err))
		}
		return len(p), nil
	}

	payload := lokiPayload{Streams: []Stream{{
		Stream: l.labels,
		Values: [][]string{
			{strconv.FormatInt(timestamp.UnixNano(), 10), string(line)},
		},
	}}}

	if err := l.send(payload); err != nil {
		if l.onError != nil {
			l.onError(fmt.Errorf("send error: %w", err))
		}
	}

	return len(p), nil
}

func (l *LokiWriter) send(payload lokiPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s%s", l.host, lokiPushEndpoint),
		bytes.NewReader(data),
	)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		return fmt.Errorf("http send: %w", err)
	}

	defer func() {
		if _, err = io.Copy(io.Discard, resp.Body); err != nil {
			log.Printf("failed to discard response body: %v", err)
		}
		if err = resp.Body.Close(); err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("loki responded with %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
