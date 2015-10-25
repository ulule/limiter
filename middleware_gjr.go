package limiter

import (
	"strconv"

	"github.com/ant0ine/go-json-rest/rest"
)

// GJRMiddleware is the go-json-rest middleware.
type GJRMiddleware struct {
	Limiter *Limiter
}

// NewGJRMiddleware returns a new instance of go-json-rest middleware.
func NewGJRMiddleware(limiter *Limiter) *GJRMiddleware {
	return &GJRMiddleware{
		Limiter: limiter,
	}
}

// MiddlewareFunc is the middleware method (handler).
func (m *GJRMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		context, err := m.Limiter.Get(GetIPKey(r.Request))
		if err != nil {
			panic(err)
		}

		w.Header().Add("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		w.Header().Add("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		w.Header().Add("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		// That can be useful to access rate limit context in views.
		r.Env["ratelimit:limit"] = context.Limit
		r.Env["ratelimit:remaining"] = context.Remaining
		r.Env["ratelimit:reset"] = context.Reset

		if context.Reached {
			rest.Error(w, "Limit exceeded", 429)
			return
		}

		h(w, r)
	}
}
