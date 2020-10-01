package echo_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	libEcho "github.com/labstack/echo"
	"github.com/stretchr/testify/require"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/echo"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func TestHTTPMiddleware(t *testing.T) {
	is := require.New(t)

	request, err := http.NewRequest("GET", "/", nil)
	is.NoError(err)
	is.NotNil(request)

	store := memory.NewStore()
	is.NotZero(store)

	rate, err := limiter.NewRateFromFormatted("10-M")
	is.NoError(err)
	is.NotZero(rate)

	middleware := echo.NewMiddleware(limiter.New(store, rate))
	is.NotZero(middleware)

	e := libEcho.New()
	e.Use(middleware)
	e.GET("/", func(c libEcho.Context) error {
		return c.String(http.StatusOK, "hello")
	})

	success := int64(10)
	clients := int64(100)

	//
	// Sequential
	//

	for i := int64(1); i <= clients; i++ {

		resp := httptest.NewRecorder()
		e.ServeHTTP(resp, request)

		if i <= success {
			is.Equal(resp.Code, http.StatusOK)
		} else {
			is.Equal(resp.Code, http.StatusTooManyRequests)
		}
	}

	//
	// Concurrent
	//

	store = memory.NewStore()
	is.NotZero(store)

	middleware = echo.NewMiddleware(limiter.New(store, rate))
	is.NotZero(middleware)

	e = libEcho.New()
	e.Use(middleware)
	e.GET("/", func(c libEcho.Context) error {
		return c.String(http.StatusOK, "hello")
	})

	wg := &sync.WaitGroup{}
	counter := int64(0)

	for i := int64(1); i <= clients; i++ {
		wg.Add(1)
		go func() {

			resp := httptest.NewRecorder()
			e.ServeHTTP(resp, request)

			if resp.Code == http.StatusOK {
				atomic.AddInt64(&counter, 1)
			}

			wg.Done()
		}()
	}

	wg.Wait()
	is.Equal(success, atomic.LoadInt64(&counter))

	//
	// Custom KeyGetter
	//

	store = memory.NewStore()
	is.NotZero(store)

	counter = int64(0)
	keyGetter := func(c libEcho.Context) string {
		v := atomic.AddInt64(&counter, 1)
		return strconv.FormatInt(v, 10)
	}

	middleware = echo.NewMiddleware(limiter.New(store, rate), echo.WithKeyGetter(keyGetter))
	is.NotZero(middleware)

	e = libEcho.New()
	e.Use(middleware)
	e.GET("/", func(c libEcho.Context) error {
		return c.String(http.StatusOK, "hello")
	})

	for i := int64(1); i <= clients; i++ {
		resp := httptest.NewRecorder()
		e.ServeHTTP(resp, request)
		// We should always be ok as the key changes for each request
		is.Equal(http.StatusOK, resp.Code, strconv.FormatInt(i, 10))
	}

	//
	// Test ExcludedKey
	//
	store = memory.NewStore()
	is.NotZero(store)
	counter = int64(0)
	excludedKeyFn := func(key string) bool {
		return key == "1"
	}
	middleware = echo.NewMiddleware(limiter.New(store, rate),
		echo.WithKeyGetter(func(c libEcho.Context) string {
			v := atomic.AddInt64(&counter, 1)
			return strconv.FormatInt(v%2, 10)
		}),
		echo.WithExcludedKey(excludedKeyFn),
	)
	is.NotZero(middleware)

	e = libEcho.New()
	e.Use(middleware)
	e.GET("/", func(c libEcho.Context) error {
		return c.String(http.StatusOK, "hello")
	})
	success = 20
	for i := int64(1); i < clients; i++ {
		resp := httptest.NewRecorder()
		e.ServeHTTP(resp, request)
		if i <= success || i%2 == 1 {
			is.Equal(http.StatusOK, resp.Code, strconv.FormatInt(i, 10))
		} else {
			is.Equal(resp.Code, http.StatusTooManyRequests)
		}
	}
}
