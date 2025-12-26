package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Total number of requests your LB has processed
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lb_http_requests_total",
			Help: "Total number of HTTP requests processed by the load balancer",
		},
		[]string{"method", "code"},
	)

	// Number of backends currently marked alive
	AliveBackends = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "lb_backends_alive",
			Help: "Current number of healthy backends",
		},
	)

	// Request durations in seconds (latency)
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "lb_request_duration_seconds",
			Help:    "Histogram of response time for load balancer",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "code"},
	)
	BackendLatencyEWMA = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lb_backend_latency_ewma_ms",
			Help: "EWMA latency per backend in milliseconds",
		},
		[]string{"backend"},
	)
	RateLimitedRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lb_rate_limited_requests_total",
		Help: "Total number of rate limited requests",
	})
	BackendCircuitState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lb_backend_circuit_state",
			Help: "Circuit breaker state per backend",
		},
		[]string{"backend"},
	)
)

func init() {
	// Register metrics with Prometheus's default registry
	prometheus.MustRegister(RequestCount)
	prometheus.MustRegister(AliveBackends)
	prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(BackendLatencyEWMA)
}

// Handler returns an http.Handler that exposes /metrics endpoint
func Handler() http.Handler {
	return promhttp.Handler()
}
