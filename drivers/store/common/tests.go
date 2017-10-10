package common

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ulule/limiter"
)

// TestStoreSequentialAccess verify that store works as expected with a sequential access.
func TestStoreSequentialAccess(t *testing.T, store limiter.Store) {
	is := require.New(t)
	ctx := context.Background()

	limiter := limiter.New(store, limiter.Rate{
		Limit:  3,
		Period: time.Minute,
	})

	for i := 1; i <= 6; i++ {

		if i <= 3 {

			lctx, err := limiter.Peek(ctx, "foo")
			is.NoError(err)
			is.NotZero(lctx)
			is.Equal(int64(3-(i-1)), lctx.Remaining)

		}

		lctx, err := limiter.Get(ctx, "foo")
		is.NoError(err)
		is.NotZero(lctx)

		if i <= 3 {

			is.Equal(int64(3), lctx.Limit)
			is.Equal(int64(3-i), lctx.Remaining)
			is.True(math.Ceil(time.Since(time.Unix(lctx.Reset, 0)).Seconds()) <= 60)

			lctx, err = limiter.Peek(ctx, "foo")
			is.NoError(err)
			is.Equal(int64(3-i), lctx.Remaining)

		} else {

			is.Equal(int64(3), lctx.Limit)
			is.True(lctx.Remaining == 0)
			is.True(math.Ceil(time.Since(time.Unix(lctx.Reset, 0)).Seconds()) <= 60)

		}
	}
}

// TestStoreConcurrentAccess verify that store works as expected with a concurrent access.
func TestStoreConcurrentAccess(t *testing.T, store limiter.Store) {
	is := require.New(t)
	ctx := context.Background()

	limiter := limiter.New(store, limiter.Rate{
		Limit:  100000,
		Period: 10 * time.Minute,
	})

	goroutines := 100
	ops := 200

	wg := &sync.WaitGroup{}
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			for j := 0; j < ops; j++ {
				lctx, err := limiter.Get(ctx, "foo")
				is.NoError(err)
				is.NotZero(lctx)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
