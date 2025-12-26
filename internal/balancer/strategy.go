package balancer

import (
	"math"
	"sync"
	"math/rand"
	"github.com/vvijay2468/load-balancer/internal/backend"
)

var (
	rrIndex int
	mux     sync.Mutex
)

// NextBackend returns the next healthy backend using Round-Robin.
// If no healthy backend is found, returns nil.
func NextBackend() *backend.Backend {
	mux.Lock()
	defer mux.Unlock()

	backends := backend.GetAllBackends()

	if len(backends) == 0 {
		return nil
	}

	// Try up to the number of backends in the pool
	for i := 0; i < len(backends); i++ {
		// Circular increment of rrIndex
		rrIndex = (rrIndex + 1) % len(backends)
		b := backends[rrIndex]

		// Check health
		if b.IsAlive() {
			return b
		}
	}

	// None healthy
	return nil
}
func NextBackendAdaptive() *backend.Backend {
	backends := backend.GetAliveBackends()
	if len(backends) == 0 {
		return nil
	}

	var best *backend.Backend
	bestScore := math.MaxFloat64

	for _, b := range backends {
		latency := b.GetLatencyEWMA()
		conns := b.GetConnections()

		score := latency * float64(conns+1)+rand.Float64()

		if score < bestScore {
			bestScore = score
			best = b
		}
	}

	return best
}
