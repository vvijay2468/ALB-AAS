package balancer

import (
	"sync"
	"github.com/vvijay2468/load-balancer/internal/backend"
)

var leastMux sync.Mutex

// NextBackendLeast returns the backend with the least active connections.
// It skips unhealthy backends, returns nil if none available.
func NextBackendLeast() *backend.Backend {
	leastMux.Lock()
	defer leastMux.Unlock()

	backends := backend.GetServerPool()
	var selected *backend.Backend
	minConn := int(^uint(0) >> 1) // big int

	for _, b := range backends {
		if !b.IsAlive() {
			continue
		}
		// count active connections stored in backend struct (implement below)
		if b.ActiveConnections() < minConn {
			minConn = b.ActiveConnections()
			selected = b
		}
	}

	return selected
}
