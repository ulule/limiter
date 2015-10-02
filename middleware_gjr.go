package limiter

import (
	"fmt"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ulule/ulule-api/utils/request"
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
		context, err := m.Limiter.Get(fmt.Sprintf("%s", request.QueryIP(r.Request)))
		if err != nil {
			panic(err)
		}

		w.Header().Add("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
		w.Header().Add("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		w.Header().Add("X-RateLimit-Reset", fmt.Sprintf("%d", context.Reset))

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
