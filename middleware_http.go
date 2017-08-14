package limiter

// HTTPMiddleware is the middleware for basic http.Handler.
import (
	"net/http"
	"strconv"
)

// HTTPMiddleware is the basic HTTP middleware.
type HTTPMiddleware struct {
	Limiter *Limiter
}

// NewHTTPMiddleware return a new instance of a basic HTTP middleware.
func NewHTTPMiddleware(limiter *Limiter) *HTTPMiddleware {
	return &HTTPMiddleware{Limiter: limiter}
}

// Handler the middleware handler.
func (m *HTTPMiddleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context, err := m.Limiter.Get(GetIPKey(r))
		if err != nil {
			panic(err)
		}

		w.Header().Add("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		w.Header().Add("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		w.Header().Add("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		if context.Reached {
			http.Error(w, "Limit exceeded", 429)
			return
		}

		h.ServeHTTP(w, r)
	})
}
