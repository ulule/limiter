package phi

import (
	"github.com/fate-lovely/phi"
	"github.com/ulule/limiter/v3"
	"github.com/valyala/fasthttp"
	"strconv"
)

// Middleware is the middleware for fasthttp.
type Middleware struct {
	Limiter        *limiter.Limiter
	OnError        ErrorHandler
	OnLimitReached LimitReachedHandler
	KeyGetter      KeyGetter
}

// NewMiddleware return a new instance of a fasthttp middleware.
func NewMiddleware(limiter *limiter.Limiter, options ...Option) phi.Middleware {
	middleware := &Middleware{
		Limiter:        limiter,
		OnError:        DefaultErrorHandler,
		OnLimitReached: DefaultLimitReachedHandler,
		KeyGetter:      DefaultKeyGetter,
	}

	for _, option := range options {
		option.apply(middleware)
	}

	return func(handlerFunc phi.HandlerFunc) phi.HandlerFunc {
		return middleware.Handle(handlerFunc)
	}
}

// Handle gin request.
func (middleware *Middleware) Handle(next phi.HandlerFunc) phi.HandlerFunc {
	return func(ctx *fasthttp.RequestCtx) {
		key := middleware.KeyGetter(ctx)
		context, err := middleware.Limiter.Get(ctx, key)
		if err != nil {
			middleware.OnError(ctx, err)
			return
		}

		ctx.Response.Header.Set("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		ctx.Response.Header.Set("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		ctx.Response.Header.Set("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		if context.Reached {
			middleware.OnLimitReached(ctx)
			return
		}

		next(ctx)
	}
}
