package ratelimit

import "sync"

type Manager struct {
	buckets map[string]*TokenBucket
	mu      sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		buckets: make(map[string]*TokenBucket),
	}
}

func (m *Manager) Allow(ip string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	bucket, ok := m.buckets[ip]
	if !ok {
		bucket = NewTokenBucket(5000, 1000) // 50 burst, 10 rps
		m.buckets[ip] = bucket
	}
	return bucket.Allow()
}
