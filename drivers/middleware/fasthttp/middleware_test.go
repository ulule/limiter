package fasthttp_test

import (
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	libfasthttp "github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/fasthttp"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func TestFasthttpMiddleware(t *testing.T) {
	is := require.New(t)

	store := memory.NewStore()
	is.NotZero(store)

	rate, err := limiter.NewRateFromFormatted("10-M")
	is.NoError(err)
	is.NotZero(rate)

	middleware := fasthttp.NewMiddleware(limiter.New(store, rate))

	requestHandler := func(ctx *libfasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/":
			ctx.SetStatusCode(libfasthttp.StatusOK)
			ctx.SetBodyString("hello")
		}
	}

	success := int64(10)
	clients := int64(100)

	//
	// Sequential
	//

	for i := int64(1); i <= clients; i++ {
		resp := libfasthttp.AcquireResponse()
		req := libfasthttp.AcquireRequest()
		req.Header.SetHost("localhost:8081")
		req.Header.SetRequestURI("/")
		err := serve(middleware.Handle(requestHandler), req, resp)
		is.NoError(err)

		if i <= success {
			is.Equal(resp.StatusCode(), libfasthttp.StatusOK)
		} else {
			is.Equal(resp.StatusCode(), libfasthttp.StatusTooManyRequests)
		}
	}

	//
	// Concurrent
	//

	store = memory.NewStore()
	is.NotZero(store)

	middleware = fasthttp.NewMiddleware(limiter.New(store, rate))

	requestHandler = func(ctx *libfasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/":
			ctx.SetStatusCode(libfasthttp.StatusOK)
			ctx.SetBodyString("hello")
		}
	}

	wg := &sync.WaitGroup{}
	counter := int64(0)

	for i := int64(1); i <= clients; i++ {
		wg.Add(1)

		go func() {
			resp := libfasthttp.AcquireResponse()
			req := libfasthttp.AcquireRequest()
			req.Header.SetHost("localhost:8081")
			req.Header.SetRequestURI("/")
			err := serve(middleware.Handle(requestHandler), req, resp)
			is.NoError(err)

			if resp.StatusCode() == libfasthttp.StatusOK {
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
	keyGetter := func(c *libfasthttp.RequestCtx) string {
		v := atomic.AddInt64(&counter, 1)
		return strconv.FormatInt(v, 10)
	}

	middleware = fasthttp.NewMiddleware(limiter.New(store, rate), fasthttp.WithKeyGetter(keyGetter))
	is.NotZero(middleware)

	requestHandler = func(ctx *libfasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/":
			ctx.SetStatusCode(libfasthttp.StatusOK)
			ctx.SetBodyString("hello")
		}
	}

	for i := int64(1); i <= clients; i++ {
		resp := libfasthttp.AcquireResponse()
		req := libfasthttp.AcquireRequest()
		req.Header.SetHost("localhost:8081")
		req.Header.SetRequestURI("/")
		err := serve(middleware.Handle(requestHandler), req, resp)
		is.NoError(err)
		is.Equal(libfasthttp.StatusOK, resp.StatusCode(), strconv.FormatInt(i, 10))
	}
}

func serve(handler libfasthttp.RequestHandler, req *libfasthttp.Request, res *libfasthttp.Response) error {
	ln := fasthttputil.NewInmemoryListener()
	defer func() {
		err := ln.Close()
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		err := libfasthttp.Serve(ln, handler)
		if err != nil {
			panic(err)
		}
	}()

	client := libfasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}

	return client.Do(req, res)
}
