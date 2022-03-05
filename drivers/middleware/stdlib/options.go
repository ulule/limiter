package stdlib

import (
	"net/http"

	"github.com/ulule/limiter/v3"
)

// Option is used to define Middleware configuration.
type Option interface {
	apply(*Middleware)
}

type option func(*Middleware)

func (o option) apply(middleware *Middleware) {
	o(middleware)
}

// ErrorHandler is an handler used to inform when an error has occurred.
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

// WithErrorHandler will configure the Middleware to use the given ErrorHandler.
func WithErrorHandler(handler ErrorHandler) Option {
	return option(func(middleware *Middleware) {
		middleware.OnError = handler
	})
}

// DefaultErrorHandler is the default ErrorHandler used by a new Middleware.
func DefaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	panic(err)
}

// LimitReachedHandler is an handler used to inform when the limit has exceeded.
type LimitReachedHandler func(w http.ResponseWriter, r *http.Request)

// WithLimitReachedHandler will configure the Middleware to use the given LimitReachedHandler.
func WithLimitReachedHandler(handler LimitReachedHandler) Option {
	return option(func(middleware *Middleware) {
		middleware.OnLimitReached = handler
	})
}

// DefaultLimitReachedHandler is the default LimitReachedHandler used by a new Middleware.
func DefaultLimitReachedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Limit exceeded", http.StatusTooManyRequests)
}

// KeyGetter will define the rate limiter key given the gin Context.
type KeyGetter func(r *http.Request) string

// WithKeyGetter will configure the Middleware to use the given KeyGetter.
func WithKeyGetter(handler KeyGetter) Option {
	return option(func(middleware *Middleware) {
		middleware.KeyGetter = handler
	})
}

// DefaultKeyGetter is the default KeyGetter used by a new Middleware.
// It returns the Client IP address.
func DefaultKeyGetter(limiter *limiter.Limiter) func(r *http.Request) string {
	return func(r *http.Request) string {
		return limiter.GetIPKey(r)
	}
}

// WithExcludedKey will configure the Middleware to ignore key(s) using the given function.
func WithExcludedKey(handler func(string) bool) Option {
	return option(func(middleware *Middleware) {
		middleware.ExcludedKey = handler
	})
}
