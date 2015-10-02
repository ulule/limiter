package limiter

import (
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestQueryIP tests QueryIP() function.
func TestQueryIP(t *testing.T) {
	expected := net.ParseIP("8.8.8.8")

	r := http.Request{
		URL:        &url.URL{Path: "/"},
		Header:     http.Header{},
		RemoteAddr: "8.8.8.8:8888",
	}

	ip := QueryIP(&r)
	assert.Equal(t, expected, ip)
}
