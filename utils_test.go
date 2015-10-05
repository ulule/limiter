package limiter

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetIP tests GetIP() function.
func TestGetIP(t *testing.T) {
	//
	// RemoteAddr
	//

	expected := net.ParseIP("8.8.8.8")

	r := http.Request{
		URL:        &url.URL{Path: "/"},
		Header:     http.Header{},
		RemoteAddr: "8.8.8.8:8888",
	}

	ip := GetIP(&r)
	assert.Equal(t, expected, ip)

	//
	// X-Forwarded-For
	//

	expected = net.ParseIP("9.9.9.9")

	r = http.Request{
		URL:        &url.URL{Path: "/foo"},
		Header:     http.Header{},
		RemoteAddr: "8.8.8.8:8888",
	}

	r.Header.Add("X-Forwarded-For", "9.9.9.9, 7.7.7.7, 6.6.6.6")
	ip = GetIP(&r)
	assert.Equal(t, expected, ip)

	//
	// X-Real-IP
	//

	expected = net.ParseIP("6.6.6.6")

	r = http.Request{
		URL:        &url.URL{Path: "/bar"},
		Header:     http.Header{},
		RemoteAddr: "8.8.8.8:8888",
	}

	r.Header.Add("X-Real-IP", "6.6.6.6")
	ip = GetIP(&r)
	assert.Equal(t, expected, ip)
}

// TestGetIPKey tests GetIPKey() function.
func TestGetIPKey(t *testing.T) {
	ip := net.ParseIP("8.8.8.8")

	r := http.Request{
		URL:        &url.URL{Path: "/"},
		Header:     http.Header{},
		RemoteAddr: "8.8.8.8:8888",
	}

	h := sha256.New()
	h.Write([]byte(string(ip)))
	expected := hex.EncodeToString(h.Sum(nil))

	assert.Equal(t, expected, GetIPKey(&r))
}
