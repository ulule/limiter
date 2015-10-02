package limiter

import (
	"math/rand"
	"net"
	"net/http"
	"strings"
)

// QueryIP returns real IP address from request.
func QueryIP(r *http.Request) net.IP {
	if r.URL.Path[len(r.URL.Path)-1] == '/' {
		return RemoteAddr(r)
	}

	q := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	if ip := net.ParseIP(q); ip != nil {
		return ip
	}

	ip, err := net.LookupIP(q)
	if err != nil {
		return nil
	}

	if len(ip) == 0 {
		return nil
	}

	return ip[rand.Intn(len(ip))]
}

// RemoteAddr returns remote IP address from request.
func RemoteAddr(r *http.Request) net.IP {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return net.ParseIP(r.RemoteAddr)
	}

	return net.ParseIP(host)
}
