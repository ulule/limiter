package limiter_test

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ulule/limiter"
)

func TestGetIP(t *testing.T) {
	is := require.New(t)

	request1 := &http.Request{
		URL:        &url.URL{Path: "/"},
		Header:     http.Header{},
		RemoteAddr: "8.8.8.8:8888",
	}

	request2 := &http.Request{
		URL:        &url.URL{Path: "/foo"},
		Header:     http.Header{},
		RemoteAddr: "8.8.8.8:8888",
	}
	request2.Header.Add("X-Forwarded-For", "9.9.9.9, 7.7.7.7, 6.6.6.6")

	request3 := &http.Request{
		URL:        &url.URL{Path: "/bar"},
		Header:     http.Header{},
		RemoteAddr: "8.8.8.8:8888",
	}
	request3.Header.Add("X-Real-IP", "6.6.6.6")

	scenarios := []struct {
		request  *http.Request
		hasProxy bool
		expected net.IP
	}{
		{
			//
			// Scenario #1 : RemoteAddr without proxy.
			//
			request:  request1,
			hasProxy: false,
			expected: net.ParseIP("8.8.8.8"),
		},
		{
			//
			// Scenario #2 : X-Forwarded-For without proxy.
			//
			request:  request2,
			hasProxy: false,
			expected: net.ParseIP("8.8.8.8"),
		},
		{
			//
			// Scenario #3 : X-Real-IP without proxy.
			//
			request:  request3,
			hasProxy: false,
			expected: net.ParseIP("8.8.8.8"),
		},
		{
			//
			// Scenario #4 : RemoteAddr with proxy.
			//
			request:  request1,
			hasProxy: true,
			expected: net.ParseIP("8.8.8.8"),
		},
		{
			//
			// Scenario #5 : X-Forwarded-For with proxy.
			//
			request:  request2,
			hasProxy: true,
			expected: net.ParseIP("9.9.9.9"),
		},
		{
			//
			// Scenario #6 : X-Real-IP with proxy.
			//
			request:  request3,
			hasProxy: true,
			expected: net.ParseIP("6.6.6.6"),
		},
	}

	for i, scenario := range scenarios {
		message := fmt.Sprintf("Scenario #%d", (i + 1))
		ip := limiter.GetIP(scenario.request, scenario.hasProxy)
		is.Equal(scenario.expected, ip, message)
	}
}

func TestGetIPKey(t *testing.T) {
	is := require.New(t)

	request1 := &http.Request{
		URL:        &url.URL{Path: "/"},
		Header:     http.Header{},
		RemoteAddr: "8.8.8.8:8888",
	}

	request2 := &http.Request{
		URL:        &url.URL{Path: "/foo"},
		Header:     http.Header{},
		RemoteAddr: "8.8.8.8:8888",
	}
	request2.Header.Add("X-Forwarded-For", "9.9.9.9, 7.7.7.7, 6.6.6.6")

	request3 := &http.Request{
		URL:        &url.URL{Path: "/bar"},
		Header:     http.Header{},
		RemoteAddr: "8.8.8.8:8888",
	}
	request3.Header.Add("X-Real-IP", "6.6.6.6")

	scenarios := []struct {
		request  *http.Request
		hasProxy bool
		expected string
	}{
		{
			//
			// Scenario #1 : RemoteAddr without proxy.
			//
			request:  request1,
			hasProxy: false,
			expected: "8.8.8.8",
		},
		{
			//
			// Scenario #2 : X-Forwarded-For without proxy.
			//
			request:  request2,
			hasProxy: false,
			expected: "8.8.8.8",
		},
		{
			//
			// Scenario #3 : X-Real-IP without proxy.
			//
			request:  request3,
			hasProxy: false,
			expected: "8.8.8.8",
		},
		{
			//
			// Scenario #4 : RemoteAddr without proxy.
			//
			request:  request1,
			hasProxy: true,
			expected: "8.8.8.8",
		},
		{
			//
			// Scenario #5 : X-Forwarded-For without proxy.
			//
			request:  request2,
			hasProxy: true,
			expected: "9.9.9.9",
		},
		{
			//
			// Scenario #6 : X-Real-IP without proxy.
			//
			request:  request3,
			hasProxy: true,
			expected: "6.6.6.6",
		},
	}

	for i, scenario := range scenarios {
		message := fmt.Sprintf("Scenario #%d", (i + 1))
		key := limiter.GetIPKey(scenario.request, scenario.hasProxy)
		is.Equal(scenario.expected, key, message)
	}
}
