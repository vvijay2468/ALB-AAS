package balancer

import (
	"crypto/md5"
	"encoding/binary"
	"net"
	"net/http"

	"github.com/vvijay2468/load-balancer/internal/backend"
)

func NextBackendSticky(r *http.Request) *backend.Backend {
	backends := backend.GetServerPool()
	// get client IP
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	hash := md5.Sum([]byte(ip))
	idx := int(binary.BigEndian.Uint32(hash[:4]) % uint32(len(backends)))

	for i := 0; i < len(backends); i++ {
		b := backends[(idx+i)%len(backends)]
		if b.IsAlive() {
			return b
		}
	}
	return nil
}
