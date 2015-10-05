package limiter

import (
	"math"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func TestLimiterMemory(t *testing.T) {
	rate, err := NewRateFromFormatted("3-M")
	assert.Nil(t, err)

	store := NewMemoryStore("limitertests:memory", 30*time.Second)

	limiter := NewLimiter(store, rate)

	i := 1
	for i <= 5 {
		ctx, err := limiter.Get("boo")
		assert.Nil(t, err)

		if i <= 3 {
			assert.Equal(t, int64(3), ctx.Limit)
			assert.Equal(t, int64(3-i), ctx.Remaining)
			assert.True(t, math.Ceil(time.Since(time.Unix(ctx.Reset, 0)).Seconds()) <= 60)

		} else {
			assert.Equal(t, int64(3), ctx.Limit)
			assert.True(t, ctx.Remaining == 0)
			assert.True(t, math.Ceil(time.Since(time.Unix(ctx.Reset, 0)).Seconds()) <= 60)
		}

		i++
	}
}

// TestLimiterRedis tests ratelimit.Limiter with Redis store.
func TestLimiterRedis(t *testing.T) {
	rate, err := NewRateFromFormatted("3-M")
	assert.Nil(t, err)

	store, err := NewRedisStore(newRedisPool(), "limitertests:redis")
	assert.Nil(t, err)

	limiter := NewLimiter(store, rate)

	i := 1
	for i <= 5 {
		ctx, err := limiter.Get("boo")
		assert.Nil(t, err)

		if i <= 3 {
			assert.Equal(t, int64(3), ctx.Limit)
			assert.Equal(t, int64(3-i), ctx.Remaining)
			assert.True(t, math.Ceil(time.Since(time.Unix(ctx.Reset, 0)).Seconds()) <= 60)

		} else {
			assert.Equal(t, int64(3), ctx.Limit)
			assert.True(t, ctx.Remaining == 0)
			assert.True(t, math.Ceil(time.Since(time.Unix(ctx.Reset, 0)).Seconds()) <= 60)
		}

		i++
	}
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// newRedisPool returns
func newRedisPool() *redis.Pool {
	return redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", ":6379")
		if err != nil {
			return nil, err
		}
		return c, err
	}, 100)
}

// newRedisLimiter returns an instance of limiter with redis backend.
func newRedisLimiter(formattedQuota string, prefix string) *Limiter {
	rate, err := NewRateFromFormatted(formattedQuota)
	if err != nil {
		panic(err)
	}

	store, err := NewRedisStore(newRedisPool(), prefix)
	if err != nil {
		panic(err)
	}

	return NewLimiter(store, rate)
}
