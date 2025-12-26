package backend

var (
	backends []*Backend
)

// addBackendUnsafe assumes caller holds lock
func addBackendUnsafe(b *Backend) {
	backends = append(backends, b)
}

// AddBackendToPool registers a backend
func AddBackendToPool(b *Backend) {
	poolMu.Lock()
	defer poolMu.Unlock()
	addBackendUnsafe(b)
}

// GetAllBackends returns a COPY of all backends
func GetAllBackends() []*Backend {
	poolMu.RLock()
	defer poolMu.RUnlock()
	return append([]*Backend{}, backends...)
}

