package limiter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHTTPMiddleware tests the HTTP middleware.
func TestHTTPMiddleware(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "178.1.2.3:128"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	mw := NewHTTPMiddleware(newRedisLimiter("5-M", "limitertests:http")).Handler(handler)

	i := 1
	for i <= 10 {
		res := httptest.NewRecorder()
		mw.ServeHTTP(res, req)
		if i <= 5 {
			assert.Equal(t, res.Code, 200)
		} else {
			assert.Equal(t, res.Code, 429)
		}
		i++
	}
}
