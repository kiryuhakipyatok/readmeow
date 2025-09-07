package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusSetup struct {
	Registry            *prometheus.Registry
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPErrorTotal      *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
}

func NewPrometheusSetup() *PrometheusSetup {
	registry := prometheus.NewRegistry()
	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method"},
	)
	registry.MustRegister(httpRequestsTotal)
	httpRequestsDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_sec",
			Help:    "Duration of requests in seconds",
			Buckets: prometheus.LinearBuckets(0.1, 0.1, 10),
		},
		[]string{"path", "method", "status"},
	)
	registry.MustRegister(httpRequestsDuration)
	httpErrorTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_error_total",
			Help: "Total number of HTTP errors",
		},
		[]string{"path", "method", "status", "error"},
	)
	registry.MustRegister(httpErrorTotal)
	return &PrometheusSetup{
		Registry:            registry,
		HTTPRequestsTotal:   httpRequestsTotal,
		HTTPRequestDuration: httpRequestsDuration,
		HTTPErrorTotal:      httpErrorTotal,
	}
}
