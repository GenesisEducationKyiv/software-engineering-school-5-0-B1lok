package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type СacheMetrics struct {
	cacheHit  prometheus.Counter
	cacheMiss prometheus.Counter
}

func NewCacheMetrics(namespace, subsystem string) *СacheMetrics {
	return &СacheMetrics{
		cacheHit: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "cache_hit",
			Help:      "Total number of cache hits",
		}),
		cacheMiss: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "cache_miss",
			Help:      "Total number of cache misses",
		}),
	}
}

func (m *СacheMetrics) CacheHit() {
	m.cacheHit.Inc()
}

func (m *СacheMetrics) CacheMiss() {
	m.cacheMiss.Inc()
}
