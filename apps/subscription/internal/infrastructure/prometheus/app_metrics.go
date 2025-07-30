package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "subscription_service"

var buckets = []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2}

type AppMetrics struct {
	activeSubscriptions prometheus.Gauge
	requestCounter      *prometheus.CounterVec
	requestDuration     *prometheus.HistogramVec
}

func NewAppMetrics() *AppMetrics {
	return &AppMetrics{
		activeSubscriptions: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "active_subscriptions",
			Help:      "Current number of active subscriptions",
		}),
		requestCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "total_requests",
			Help:      "Total number of requests received by the subscription service",
		},
			[]string{"method", "status"},
		),
		requestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_duration_seconds",
			Help:      "Duration of requests in seconds",
			Buckets:   buckets,
		}, []string{"method", "status"}),
	}
}

func (m *AppMetrics) ObserveRequestDuration(method, status string, duration time.Duration) {
	m.requestDuration.WithLabelValues(method, status).Observe(duration.Seconds())
}

func (m *AppMetrics) ObserveTotalRequests(method, status string) {
	m.requestCounter.WithLabelValues(method, status).Inc()
}

func (m *AppMetrics) IncActiveSubscriptions() {
	m.activeSubscriptions.Inc()
}

func (m *AppMetrics) DecActiveSubscriptions() {
	m.activeSubscriptions.Dec()
}
