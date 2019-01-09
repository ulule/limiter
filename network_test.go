package limiter_test

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ulule/limiter/v3"
)

func TestGetIP(t *testing.T) {
	is := require.New(t)

	limiter1 := New(limiter.WithTrustForwardHeader(false))
	limiter2 := New(limiter.WithTrustForwardHeader(true))
	limiter3 := New(limiter.WithIPv4Mask(net.CIDRMask(24, 32)))
	limiter4 := New(limiter.WithIPv6Mask(net.CIDRMask(48, 128)))

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

	request4 := &http.Request{
		URL:        &url.URL{Path: "/"},
		Header:     http.Header{},
		RemoteAddr: "[2001:db8:cafe:1234:beef::fafa]:8888",
	}

	scenarios := []struct {
		request  *http.Request
		limiter  *limiter.Limiter
		expected net.IP
	}{
		{
			//
			// Scenario #1 : RemoteAddr without proxy.
			//
			request:  request1,
			limiter:  limiter1,
			expected: net.ParseIP("8.8.8.8").To4(),
		},
		{
			//
			// Scenario #2 : X-Forwarded-For without proxy.
			//
			request:  request2,
			limiter:  limiter1,
			expected: net.ParseIP("8.8.8.8").To4(),
		},
		{
			//
			// Scenario #3 : X-Real-IP without proxy.
			//
			request:  request3,
			limiter:  limiter1,
			expected: net.ParseIP("8.8.8.8").To4(),
		},
		{
			//
			// Scenario #4 : RemoteAddr with proxy.
			//
			request:  request1,
			limiter:  limiter2,
			expected: net.ParseIP("8.8.8.8").To4(),
		},
		{
			//
			// Scenario #5 : X-Forwarded-For with proxy.
			//
			request:  request2,
			limiter:  limiter2,
			expected: net.ParseIP("9.9.9.9").To4(),
		},
		{
			//
			// Scenario #6 : X-Real-IP with proxy.
			//
			request:  request3,
			limiter:  limiter2,
			expected: net.ParseIP("6.6.6.6").To4(),
		},
		{
			//
			// Scenario #7 : IPv4 with mask.
			//
			request:  request1,
			limiter:  limiter3,
			expected: net.ParseIP("8.8.8.0").To4(),
		},
		{
			//
			// Scenario #8 : IPv6 with mask.
			//
			request:  request4,
			limiter:  limiter4,
			expected: net.ParseIP("2001:db8:cafe::").To16(),
		},
	}

	for i, scenario := range scenarios {
		message := fmt.Sprintf("Scenario #%d", (i + 1))
		ip := scenario.limiter.GetIPWithMask(scenario.request)
		is.Equal(scenario.expected, ip, message)
	}
}

func TestGetIPKey(t *testing.T) {
	is := require.New(t)

	limiter1 := New(limiter.WithTrustForwardHeader(false))
	limiter2 := New(limiter.WithTrustForwardHeader(true))
	limiter3 := New(limiter.WithIPv4Mask(net.CIDRMask(24, 32)))
	limiter4 := New(limiter.WithIPv6Mask(net.CIDRMask(48, 128)))

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

	request4 := &http.Request{
		URL:        &url.URL{Path: "/"},
		Header:     http.Header{},
		RemoteAddr: "[2001:db8:cafe:1234:beef::fafa]:8888",
	}

	scenarios := []struct {
		request  *http.Request
		limiter  *limiter.Limiter
		expected string
	}{
		{
			//
			// Scenario #1 : RemoteAddr without proxy.
			//
			request:  request1,
			limiter:  limiter1,
			expected: "8.8.8.8",
		},
		{
			//
			// Scenario #2 : X-Forwarded-For without proxy.
			//
			request:  request2,
			limiter:  limiter1,
			expected: "8.8.8.8",
		},
		{
			//
			// Scenario #3 : X-Real-IP without proxy.
			//
			request:  request3,
			limiter:  limiter1,
			expected: "8.8.8.8",
		},
		{
			//
			// Scenario #4 : RemoteAddr without proxy.
			//
			request:  request1,
			limiter:  limiter2,
			expected: "8.8.8.8",
		},
		{
			//
			// Scenario #5 : X-Forwarded-For without proxy.
			//
			request:  request2,
			limiter:  limiter2,
			expected: "9.9.9.9",
		},
		{
			//
			// Scenario #6 : X-Real-IP without proxy.
			//
			request:  request3,
			limiter:  limiter2,
			expected: "6.6.6.6",
		},
		{
			//
			// Scenario #7 : IPv4 with mask.
			//
			request:  request1,
			limiter:  limiter3,
			expected: "8.8.8.0",
		},
		{
			//
			// Scenario #8 : IPv6 with mask.
			//
			request:  request4,
			limiter:  limiter4,
			expected: "2001:db8:cafe::",
		},
	}

	for i, scenario := range scenarios {
		message := fmt.Sprintf("Scenario #%d", (i + 1))
		key := scenario.limiter.GetIPKey(scenario.request)
		is.Equal(scenario.expected, key, message)
	}
}
