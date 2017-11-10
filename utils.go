package limiter

import (
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

// GetIP returns IP address from request.
func GetIP(r *http.Request, trustForwardHeader ...bool) net.IP {
	if len(trustForwardHeader) >= 1 && trustForwardHeader[0] {
		ip := r.Header.Get("X-Forwarded-For")
		if ip != "" {
			parts := strings.SplitN(ip, ",", 2)
			part := strings.TrimSpace(parts[0])
			return net.ParseIP(part)
		}

		ip = strings.TrimSpace(r.Header.Get("X-Real-IP"))
		if ip != "" {
			return net.ParseIP(ip)
		}
	}

	remoteAddr := strings.TrimSpace(r.RemoteAddr)
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return net.ParseIP(remoteAddr)
	}

	return net.ParseIP(host)
}

// GetIPKey extracts IP from request and returns hashed IP to use as store key.
func GetIPKey(r *http.Request, trustForwardHeader ...bool) string {
	return GetIP(r, trustForwardHeader...).String()
}

// Random return a random integer between min and max.
func Random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
