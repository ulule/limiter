package limiter

import (
	"net"
	"net/http"
	"strings"
)

var (
	// DefaultIPv4Mask defines the default IPv4 mask used to obtain user IP.
	DefaultIPv4Mask = net.CIDRMask(32, 32)
	// DefaultIPv6Mask defines the default IPv6 mask used to obtain user IP.
	DefaultIPv6Mask = net.CIDRMask(128, 128)
)

// GetIP returns IP address from request.
func (limiter *Limiter) GetIP(r *http.Request) net.IP {
	if limiter.Options.TrustForwardHeader {
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

// GetIPWithMask returns IP address from request by applying a mask.
func (limiter *Limiter) GetIPWithMask(r *http.Request) net.IP {
	ip := limiter.GetIP(r)
	if ip.To4() != nil {
		return ip.Mask(limiter.Options.IPv4Mask)
	}
	if ip.To16() != nil {
		return ip.Mask(limiter.Options.IPv6Mask)
	}
	return ip
}

// GetIPKey extracts IP from request and returns hashed IP to use as store key.
func (limiter *Limiter) GetIPKey(r *http.Request) string {
	return limiter.GetIPWithMask(r).String()
}
