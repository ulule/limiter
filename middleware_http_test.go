package limiter

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
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

// TestHTTPMiddlewareWithRaceCondition tests the HTTP middleware under race condition.
func TestHTTPMiddlewareWithRaceCondition(t *testing.T) {
	runtime.GOMAXPROCS(4)

	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "178.1.2.28:128"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	mw := NewHTTPMiddleware(newRedisLimiter("11-M", "limitertests:http")).Handler(handler)

	nbRequests := 100
	successCount := 0

	var wg sync.WaitGroup
	wg.Add(nbRequests)

	for i := 1; i <= nbRequests; i++ {
		go func() {
			res := httptest.NewRecorder()
			mw.ServeHTTP(res, req)
			if res.Code == 200 {
				successCount++
			}
			wg.Done()
		}()
	}

	wg.Wait()

	assert.Equal(t, 11, successCount)
}
