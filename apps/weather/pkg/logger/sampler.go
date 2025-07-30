package logger

import (
	"math/rand"

	"github.com/rs/zerolog"
)

type PercentageSampler struct {
	rate float64
}

func NewPercentageSampler(rate float64) *PercentageSampler {
	if rate < 0.0 {
		rate = 0.0
	} else if rate > 1.0 {
		rate = 1.0
	}
	return &PercentageSampler{rate: rate}
}

func (s *PercentageSampler) Sample(_ zerolog.Level) bool {
	return rand.Float64() < s.rate //nolint:gosec
}
