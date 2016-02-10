package limiter

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// TestLimiterMemory tests Limiter with memory store.
func TestLimiterMemory(t *testing.T) {
	rate, err := NewRateFromFormatted("3-M")
	assert.Nil(t, err)

	store := NewMemoryStoreWithOptions(StoreOptions{
		Prefix:          "limitertests:memory",
		CleanUpInterval: 30 * time.Second,
	})

	testLimiter(t, store, rate)
}

// TestLimiterRedis tests Limiter with Redis store.
func TestLimiterRedis(t *testing.T) {
	rate, err := NewRateFromFormatted("3-M")
	assert.Nil(t, err)

	randPrefix := RandStringRunes(10)
	store, err := NewRedisStoreWithOptions(
		newRedisPool(),
		StoreOptions{Prefix: "limitertests:redis_" + randPrefix, MaxRetry: 3})

	assert.Nil(t, err)

	testLimiter(t, store, rate)
}

func testLimiter(t *testing.T, store Store, rate Rate) {
	limiter := NewLimiter(store, rate)

	i := 1
	for i <= 5 {
		if i <= 3 {
			ctx, err := limiter.Peek("boo")
			assert.NoError(t, err)
			assert.Equal(t, int64(3-(i-1)), ctx.Remaining)
		}

		ctx, err := limiter.Get("boo")
		assert.NoError(t, err)

		if i <= 3 {
			assert.Equal(t, int64(3), ctx.Limit)
			assert.Equal(t, int64(3-i), ctx.Remaining)
			assert.True(t, math.Ceil(time.Since(time.Unix(ctx.Reset, 0)).Seconds()) <= 60)
			ctx, err := limiter.Peek("boo")
			assert.NoError(t, err)
			assert.Equal(t, int64(3-i), ctx.Remaining)
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

	store, err := NewRedisStoreWithOptions(
		newRedisPool(),
		StoreOptions{Prefix: prefix, MaxRetry: 3})

	if err != nil {
		panic(err)
	}

	return NewLimiter(store, rate)
}
