package limiter

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/stretchr/testify/assert"
)

// TestRate tests ratelimit.Rate methods.
func TestGJRMiddleware(t *testing.T) {
	api := rest.NewApi()

	api.Use(NewGJRMiddleware(newRedisLimiter("10-M", "limitertests:gjr")))

	var reset int64

	api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
		reset = r.Env["ratelimit:reset"].(int64)
		w.WriteJson(map[string]string{"message": "ok"})
	}))

	handler := api.MakeHandler()
	req := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	req.RemoteAddr = "178.1.2.3:124"

	i := 1
	for i < 20 {
		recorded := test.RunRequest(t, handler, req)
		assert.True(t, math.Ceil(time.Since(time.Unix(reset, 0)).Seconds()) <= 60)
		if i <= 10 {
			recorded.BodyIs(`{"message":"ok"}`)
			recorded.HeaderIs("X-Ratelimit-Limit", "10")
			recorded.HeaderIs("X-Ratelimit-Remaining", fmt.Sprintf("%d", 10-i))
			recorded.CodeIs(200)
		} else {
			recorded.BodyIs(`{"Error":"Limit exceeded"}`)
			recorded.HeaderIs("X-Ratelimit-Limit", "10")
			recorded.HeaderIs("X-Ratelimit-Remaining", "0")
			recorded.CodeIs(429)
		}
		i++
	}
}
