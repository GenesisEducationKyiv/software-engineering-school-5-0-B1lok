package logger

import (
	"common/logger"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"subscription-service/internal/config"
)

var labels = map[string]string{
	"job": "subscription-service",
}

func Configure(cfg config.Config) {
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}

	var writers []io.Writer
	writers = append(writers, consoleWriter)

	lokiWriter := logger.NewLokiWriter(cfg.LokiHost, labels, func(err error) {
		log.Printf("Loki error: %v", err)
	})
	writers = append(writers, lokiWriter)

	log.Logger = zerolog.New(zerolog.MultiLevelWriter(writers...)).
		With().
		Timestamp().
		Logger()

	if cfg.LogSampling.Enabled {
		sampler := zerolog.LevelSampler{
			TraceSampler: NewPercentageSampler(cfg.LogSampling.Trace),
			DebugSampler: NewPercentageSampler(cfg.LogSampling.Debug),
			InfoSampler:  NewPercentageSampler(cfg.LogSampling.Info),
			WarnSampler:  NewPercentageSampler(cfg.LogSampling.Warn),
			ErrorSampler: NewPercentageSampler(cfg.LogSampling.Error),
		}
		log.Logger = log.Logger.Sample(sampler)
	}
}
