package limiter

import (
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

// GetIP returns IP address from request.
func GetIP(r *http.Request) net.IP {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		return net.ParseIP(parts[0])
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return net.ParseIP(ip)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return net.ParseIP(r.RemoteAddr)
	}

	return net.ParseIP(host)
}

// GetIPKey extracts IP from request and returns hashed IP to use as store key.
func GetIPKey(r *http.Request) string {
	return GetIP(r).String()
}

// Random return a random integer between min and max.
func Random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
