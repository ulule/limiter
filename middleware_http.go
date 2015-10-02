package limiter

// HTTPMiddleware is the middleware for basic http.Handler.
import (
	"fmt"
	"net/http"
)

// HTTPMiddleware is the basic HTTP middleware.
type HTTPMiddleware struct {
	Limiter *Limiter
}

// NewHTTPMiddleware return a new instance of go-json-rest middleware.
func NewHTTPMiddleware(limiter *Limiter) *HTTPMiddleware {
	return &HTTPMiddleware{Limiter: limiter}
}

// Handler the middleware handler.
func (m *HTTPMiddleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context, err := m.Limiter.Get(fmt.Sprintf("%s", GetIP(r)))
		if err != nil {
			panic(err)
		}

		w.Header().Add("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
		w.Header().Add("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		w.Header().Add("X-RateLimit-Reset", fmt.Sprintf("%d", context.Reset))

		if context.Reached {
			http.Error(w, "Limit exceeded", 429)
			return
		}

		h.ServeHTTP(w, r)
	})
}
