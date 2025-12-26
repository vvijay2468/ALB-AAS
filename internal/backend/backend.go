package backend

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// Backend represents a backend server
type Backend struct {
	URL   *url.URL
	Alive bool
	Proxy *httputil.ReverseProxy
	// mux             sync.RWMutex
	currConnections int32
	LatencyEWMA     float64
	mu              sync.RWMutex
}

// serverPool holds all registered backends
var (
	serverPool []*Backend
	poolMu     sync.RWMutex
)

// AddBackend adds a backend URL to the server pool
func AddBackend(rawURL string) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		log.Printf("Invalid backend URL %s: %v\n", rawURL, err)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(parsed)

	server := &Backend{
		URL:         parsed,
		Proxy:       proxy,
		Alive:       true,
		LatencyEWMA: 50.0,
	}

	poolMu.Lock()
	serverPool = append(serverPool, server)
	poolMu.Unlock()

	log.Printf("Added backend: %s\n", rawURL)
}

// HealthCheckAll checks the health of all backends
func HealthCheckAll(path string) {
	poolMu.RLock()
	defer poolMu.RUnlock()

	for _, b := range serverPool {
		go func(b *Backend) {
			client := http.Client{Timeout: 2 * time.Second}
			resp, err := client.Get(b.URL.String() + path)

			b.mu.Lock()
			if err != nil || resp.StatusCode != http.StatusOK {
				b.Alive = false
			} else {
				b.Alive = true
			}
			b.mu.Unlock()
		}(b)
	}
}
func GetAliveBackends() []*Backend {
	poolMu.RLock()
	defer poolMu.RUnlock()

	alive := make([]*Backend, 0)
	for _, b := range serverPool {
		b.mu.RLock()
		isAlive := b.Alive
		b.mu.RUnlock()

		if isAlive {
			alive = append(alive, b)
		}
	}
	return alive
}

// GetServerPool returns all backends
func GetServerPool() []*Backend {
	return serverPool
}

// IsAlive returns whether a backend is healthy
func (b *Backend) IsAlive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Alive
}

// Serve proxies the request to this backend
func (b *Backend) Serve(w http.ResponseWriter, r *http.Request) {
	b.Proxy.ServeHTTP(w, r)
}

// increment/decrement connection counters
func (b *Backend) IncrementConnections() {
	b.mu.Lock()
	b.currConnections++
	b.mu.Unlock()
}

func (b *Backend) DecrementConnections() {
	b.mu.Lock()
	if b.currConnections > 0 {
		b.currConnections--
	}
	b.mu.Unlock()
}

func (b *Backend) GetConnections() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return int64(b.currConnections)
}


func (b *Backend) ActiveConnections() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return int(b.currConnections)
}



func (b *Backend) UpdateLatency(d time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	const alpha = 0.2
	latencyMs := float64(d.Milliseconds())

	if b.LatencyEWMA == 0 {
		b.LatencyEWMA = latencyMs
		return
	}
	b.LatencyEWMA = alpha*latencyMs + (1-alpha)*b.LatencyEWMA
}

func (b *Backend) GetLatencyEWMA() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.LatencyEWMA
}

