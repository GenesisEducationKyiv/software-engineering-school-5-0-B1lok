package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"notification/internal/config"
)

func Configure(cfg config.Config) {
	log.Logger = zerolog.New(os.Stdout).
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
