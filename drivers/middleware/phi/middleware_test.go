package phi_test

import (
	libphi "github.com/fate-lovely/phi"
	"github.com/stretchr/testify/require"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/phi"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
)

func TestFasthttpMiddleware(t *testing.T) {
	is := require.New(t)
	req := fasthttp.AcquireRequest()
	req.Header.SetHost("localhost:8080")
	req.Header.SetRequestURI("/")

	store := memory.NewStore()
	is.NotZero(store)

	rate, err := limiter.NewRateFromFormatted("10-M")
	is.NoError(err)
	is.NotZero(rate)

	middleware := phi.NewMiddleware(limiter.New(store, rate))

	router := libphi.NewRouter()
	router.Use(middleware)
	router.Get("/", func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBodyString("hello")
	})

	success := int64(10)
	clients := int64(100)

	//
	// Sequential
	//

	for i := int64(1); i <= clients; i++ {
		resp := fasthttp.AcquireResponse()
		err := serve(router.ServeFastHTTP, req, resp)
		is.Nil(err)

		if i <= success {
			is.Equal(resp.StatusCode(), fasthttp.StatusOK)
		} else {
			is.Equal(resp.StatusCode(), fasthttp.StatusTooManyRequests)
		}
	}

	//
	// Concurrent
	//

	store = memory.NewStore()
	is.NotZero(store)

	middleware = phi.NewMiddleware(limiter.New(store, rate))

	router = libphi.NewRouter()
	router.Use(middleware)
	router.Get("/", func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBodyString("hello")
	})

	wg := &sync.WaitGroup{}
	counter := int64(0)

	for i := int64(1); i <= clients; i++ {
		wg.Add(1)

		go func() {
			resp := fasthttp.AcquireResponse()
			err := serve(router.ServeFastHTTP, req, resp)
			is.Nil(err)

			if resp.StatusCode() == fasthttp.StatusOK {
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

	j := 0
	KeyGetter := func(ctx *fasthttp.RequestCtx) string {
		j++
		return strconv.Itoa(j)
	}
	middleware = phi.NewMiddleware(limiter.New(store, rate), phi.WithKeyGetter(KeyGetter))

	is.NotZero(middleware)

	router = libphi.NewRouter()
	router.Use(middleware)
	router.Get("/", func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBodyString("hello")
	})

	for i := int64(1); i <= clients; i++ {
		resp := fasthttp.AcquireResponse()
		err := serve(router.ServeFastHTTP, req, resp)
		is.Nil(err)
		is.Equal(fasthttp.StatusOK, resp.StatusCode(), strconv.Itoa(int(i)))
	}
}

func serve(handler fasthttp.RequestHandler, req *fasthttp.Request, res *fasthttp.Response) error {
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close()

	go func() {
		err := fasthttp.Serve(ln, handler)
		if err != nil {
			panic(err)
		}
	}()

	client := fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}

	return client.Do(req, res)
}
