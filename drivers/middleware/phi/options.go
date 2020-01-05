package phi

import (
	"github.com/valyala/fasthttp"
)

// Option is used to define Middleware configuration
type Option interface {
	apply(middleware *Middleware)
}

type option func(*Middleware)

func (o option) apply(middleware *Middleware) {
	o(middleware)
}

// ErrorHandler is an handler used to inform when an error has occurred.
type ErrorHandler func(ctx *fasthttp.RequestCtx, err error)

// WithErrorHandler will configure the Middleware to use the given ErrorHandler.
func WithErrorHandler(handler ErrorHandler) Option {
	return option(func(middleware *Middleware) {
		middleware.OnError = handler
	})
}

// DefaultErrorHandler is the default ErrorHandler used by a new Middleware.
func DefaultErrorHandler(ctx *fasthttp.RequestCtx, err error) {
	panic(err)
}

// LimitReachedHandler is an handler used to inform when the limit has exceeded.
type LimitReachedHandler func(ctx *fasthttp.RequestCtx)

// WithLimitReachedHandler will configure the Middleware to use the given LimitReachedHandler.
func WithLimitReachedHandler(handler LimitReachedHandler) Option {
	return option(func(middleware *Middleware) {
		middleware.OnLimitReached = handler
	})
}

// DefaultLimitReachedHandler is the default LimitReachedHandler used by a new Middleware.
func DefaultLimitReachedHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusTooManyRequests)
	ctx.Response.SetBodyString("Limit exceeded")
}

// KeyGetter will define the rate limiter key given the gin Context
type KeyGetter func(ctx *fasthttp.RequestCtx) string

// WithKeyGetter will configure the Middleware to use the given KeyGetter
func WithKeyGetter(KeyGetter KeyGetter) Option {
	return option(func(middleware *Middleware) {
		middleware.KeyGetter = KeyGetter
	})
}

// DefaultKeyGetter is the default KeyGetter used by a new Middleware
// It returns the Client IP address
func DefaultKeyGetter(ctx *fasthttp.RequestCtx) string {
	return ctx.RemoteIP().String()
}
