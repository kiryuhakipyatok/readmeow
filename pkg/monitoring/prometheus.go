package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusSetup struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPErrorTotal      *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
}

func NewPrometheusSetup() *PrometheusSetup {
	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method"},
	)
	prometheus.MustRegister(httpRequestsTotal)
	httpRequestsDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_sec",
			Help:    "Duration of requests in seconds",
			Buckets: prometheus.LinearBuckets(0.1, 0.1, 10),
		},
		[]string{"path", "method", "status"},
	)
	prometheus.MustRegister(httpRequestsDuration)
	httpErrorTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_error_total",
			Help: "Total number of HTTP errors",
		},
		[]string{"path", "method", "status", "error"},
	)
	prometheus.MustRegister(httpErrorTotal)
	return &PrometheusSetup{
		HTTPRequestsTotal:   httpRequestsTotal,
		HTTPRequestDuration: httpRequestsDuration,
		HTTPErrorTotal:      httpErrorTotal,
	}
}
