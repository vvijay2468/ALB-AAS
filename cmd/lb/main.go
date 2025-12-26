package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
	"math/rand"
	// "http"
	"github.com/vvijay2468/load-balancer/internal/backend"
	"github.com/vvijay2468/load-balancer/internal/balancer"
	"github.com/vvijay2468/load-balancer/internal/metrics"
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

	// HTTPS mux (real LB)
	httpsMux := http.NewServeMux()

	httpsMux.Handle("/metrics", metrics.Handler())

	httpsMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

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
			http.Error(w, "No healthy backends available", http.StatusServiceUnavailable)
			metrics.RequestCount.WithLabelValues(r.Method, "503").Inc()
			return
		}

		b.IncrementConnections()
		defer b.DecrementConnections()
		b.Serve(w, r)

		duration := time.Since(start)

		b.UpdateLatency(duration)
		metrics.BackendLatencyEWMA.
			WithLabelValues(b.URL.Host).
			Set(b.GetLatencyEWMA())

		rec := &statusRecorder{
			ResponseWriter: w,
			status:         200,
		}
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
