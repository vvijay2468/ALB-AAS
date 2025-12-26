package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	// "http"
	"github.com/vvijay2468/load-balancer/internal/backend"
	"github.com/vvijay2468/load-balancer/internal/balancer"
	"github.com/vvijay2468/load-balancer/internal/metrics"
	"github.com/vvijay2468/load-balancer/internal/ratelimit"
)

// Config holds JSON config values
type Config struct {
	Port                string   `json:"port"`
	TLSPort             string   `json:"tls_port"`
	CertFile            string   `json:"cert_file"`
	KeyFile             string   `json:"key_file"`
	Strategy            string   `json:"strategy"`
	Backends            []string `json:"backends"`
	HealthCheckPath     string   `json:"health_check_path"`
	HealthCheckInterval int      `json:"health_check_interval"`
}
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}



var httpsMux = http.NewServeMux()

func main() {
	// 1) Load config.json
	rand.Seed(time.Now().UnixNano())
	configFile, err := os.Open("config/config.json")
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer configFile.Close()

	var config Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		log.Fatalf("Error decoding config: %v", err)
	}

	fmt.Printf("Starting load balancer on port %s...\n", config.Port)
	currentStrategy := config.Strategy
	// 2) Initialize backend pool
	for _, url := range config.Backends {
		backend.AddBackend(url)
	}

	// 3) Start health checks
	go func() {
		for {
			backend.HealthCheckAll(config.HealthCheckPath)
			time.Sleep(time.Duration(config.HealthCheckInterval) * time.Second)
		}
	}()
	// 4) Setup HTTP handlers

	rl := ratelimit.NewManager()

	httpsMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 1️ Rate limiting (FIRST)
		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "invalid client address", http.StatusBadRequest)
			return
		}

		if !rl.Allow(clientIP) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			metrics.RateLimitedRequests.Inc()
			return
		}

		// 2️ Pick backend
		var b *backend.Backend
		switch currentStrategy {
		case "least_conn":
			b = balancer.NextBackendLeast()
		case "sticky":
			b = balancer.NextBackendSticky(r)
		case "adaptive":
			b = balancer.NextBackendAdaptive()
		default:
			b = balancer.NextBackend()
		}

		if b == nil {
			http.Error(w, "no healthy backends", http.StatusServiceUnavailable)
			return
		}

		// 3️ Proxy with status recording
		b.IncrementConnections()
		defer b.DecrementConnections()

		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		b.Serve(rec, r)

		// 4️ Circuit breaker feedback
		if rec.status >= 500 {
			b.RecordFailure()
		} else {
			b.RecordSuccess()
		}

		// 5️ Metrics + latency
		duration := time.Since(start)

		b.UpdateLatency(duration)

		metrics.BackendLatencyEWMA.
			WithLabelValues(b.URL.Host).
			Set(b.GetLatencyEWMA())

		metrics.RequestCount.
			WithLabelValues(r.Method, fmt.Sprintf("%d", rec.status)).
			Inc()

		metrics.RequestDuration.
			WithLabelValues(r.Method, fmt.Sprintf("%d", rec.status)).
			Observe(duration.Seconds())
	})

	// HTTP mux (metrics + redirect)
	httpMux := http.NewServeMux()

	httpMux.Handle("/metrics", metrics.Handler())

	httpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		httpsHost := net.JoinHostPort(host, config.TLSPort)
		http.Redirect(w, r, "https://"+httpsHost+r.URL.RequestURI(), http.StatusMovedPermanently)
	})

	// Start HTTP
	go func() {
		log.Println("HTTP server listening on :8080")
		log.Fatal(http.ListenAndServe(":8080", httpMux))
	}()

	// Start HTTPS (blocks)
	log.Println("HTTPS server listening on :8443")
	log.Fatal(http.ListenAndServeTLS(":8443", config.CertFile, config.KeyFile, httpsMux))

}
