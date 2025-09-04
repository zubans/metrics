package middlewares

import (
	"net"
	"net/http"
)

// TrustedSubnetMiddleware checks that X-Real-IP belongs to the configured CIDR.
// If cidr is empty, all requests pass through.
func TrustedSubnetMiddleware(cidr string) func(http.Handler) http.Handler {
	var _, ipnet *net.IPNet
	var parseErr error
	if cidr != "" {
		var n *net.IPNet
		_, n, parseErr = net.ParseCIDR(cidr)
		ipnet = n
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cidr == "" || parseErr != nil {
				// No restriction or invalid CIDR config -> allow
				next.ServeHTTP(w, r)
				return
			}
			realIP := r.Header.Get("X-Real-IP")
			ip := net.ParseIP(realIP)
			if ip == nil || !ipnet.Contains(ip) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
