package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewProxy returns a reverse proxy for a given backend URL.
func NewProxy(target *url.URL) *httputil.ReverseProxy {
	// Create a reverse proxy that forwards to the target backend
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Optional: You can modify the request before forwarding (e.g., add headers)
	// proxy.ModifyResponse = yourModifyResponseFunc

	return proxy
}

// ServeRequest is a helper that handles proxying to the selected backend
func ServeRequest(proxy *httputil.ReverseProxy, w http.ResponseWriter, r *http.Request) {
	proxy.ServeHTTP(w, r)
}
