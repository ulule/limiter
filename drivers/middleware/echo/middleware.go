package echo

import (
	"strconv"

	"github.com/labstack/echo"
	"github.com/ulule/limiter/v3"
)

// Middleware is the middleware for echo.
type Middleware struct {
	Limiter        *limiter.Limiter
	OnError        ErrorHandler
	OnLimitReached LimitReachedHandler
	KeyGetter      KeyGetter
	ExcludedKey    func(string) bool
}

// NewMiddleware return a new instance of a echo middleware.
func NewMiddleware(limiter *limiter.Limiter, options ...Option) echo.MiddlewareFunc {
	middleware := &Middleware{
		Limiter:        limiter,
		OnError:        DefaultErrorHandler,
		OnLimitReached: DefaultLimitReachedHandler,
		KeyGetter:      DefaultKeyGetter,
		ExcludedKey:    nil,
	}

	for _, option := range options {
		option.apply(middleware)
	}

	return middleware.Handle
}

// Handle echo request.
func (m *Middleware) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		key := m.KeyGetter(c)
		if m.ExcludedKey != nil && m.ExcludedKey(key) {
			return next(c)
		}
		context, err := m.Limiter.Get(c.Request().Context(), key)
		if err != nil {
			m.OnError(c, err)

			return next(c)
		}

		h := c.Response().Header()
		h.Set("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		h.Set("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		h.Set("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		if context.Reached {
			return m.OnLimitReached(c)
		}

		return next(c)
	}
}
