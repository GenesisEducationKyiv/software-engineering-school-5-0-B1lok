package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

type FileLogger struct {
	logger zerolog.Logger
}

func NewFileLogger(logFilepath string) (*FileLogger, error) {
	dir := filepath.Dir(logFilepath)

	// #nosec G301
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(
		logFilepath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644) // #nosec G302 G304
	if err != nil {
		return nil, err
	}

	zlogger := zerolog.New(file).
		With().
		Timestamp().
		Logger()

	return &FileLogger{logger: zlogger}, nil
}

func (l *FileLogger) LogResponse(provider string, resp *http.Response) {
	if resp == nil || resp.Body == nil {
		l.logger.Warn().
			Str("provider", provider).
			Msg("Empty response")
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		l.logger.Error().
			Str("provider", provider).
			Err(err).
			Msg("Failed to read body")
		return
	}

	l.logger.Info().
		Str("provider", provider).
		RawJSON("body", bodyBytes).
		Msg("Received response")

	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
}
