package memory_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ulule/limiter/drivers/store/memory"
)

func TestCacheIncrementSequential(t *testing.T) {
	is := require.New(t)

	key := "foobar"
	cache := memory.NewCache(10 * time.Nanosecond)
	duration := 50 * time.Millisecond
	deleted := time.Now().Add(duration).UnixNano()
	epsilon := 0.001

	x, expire := cache.Increment(key, 1, duration)
	is.Equal(int64(1), x)
	is.InEpsilon(deleted, expire.UnixNano(), epsilon)

	x, expire = cache.Increment(key, 2, duration)
	is.Equal(int64(3), x)
	is.InEpsilon(deleted, expire.UnixNano(), epsilon)

	time.Sleep(duration)

	deleted = time.Now().Add(duration).UnixNano()
	x, expire = cache.Increment(key, 1, duration)
	is.Equal(int64(1), x)
	is.InEpsilon(deleted, expire.UnixNano(), epsilon)
}

func TestCacheIncrementConcurrent(t *testing.T) {
	is := require.New(t)

	goroutines := 300
	ops := 500

	expected := int64(0)
	for i := 0; i < goroutines; i++ {
		if (i % 3) == 0 {
			for j := 0; j < ops; j++ {
				expected += int64(i + j)
			}
		}
	}

	key := "foobar"
	cache := memory.NewCache(10 * time.Nanosecond)

	wg := &sync.WaitGroup{}
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			if (i % 3) == 0 {
				time.Sleep(1 * time.Second)
				for j := 0; j < ops; j++ {
					cache.Increment(key, int64(i+j), (1 * time.Second))
				}
			} else {
				time.Sleep(50 * time.Millisecond)
				stopAt := time.Now().Add(500 * time.Millisecond)
				for time.Now().Before(stopAt) {
					cache.Increment(key, int64(i), (75 * time.Millisecond))
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	value, expire := cache.Get(key, (100 * time.Millisecond))
	is.Equal(expected, value)
	is.True(time.Now().Before(expire))
}

func TestCacheGet(t *testing.T) {
	is := require.New(t)

	key := "foobar"
	cache := memory.NewCache(10 * time.Nanosecond)
	duration := 50 * time.Millisecond
	deleted := time.Now().Add(duration).UnixNano()
	epsilon := 0.001

	x, expire := cache.Get(key, duration)
	is.Equal(int64(0), x)
	is.InEpsilon(deleted, expire.UnixNano(), epsilon)

}

func TestCacheReset(t *testing.T) {
	is := require.New(t)

	key := "foobar"
	cache := memory.NewCache(10 * time.Nanosecond)
	duration := 50 * time.Millisecond
	deleted := time.Now().Add(duration).UnixNano()
	epsilon := 0.001

	x, expire := cache.Get(key, duration)
	is.Equal(int64(0), x)
	is.InEpsilon(deleted, expire.UnixNano(), epsilon)

	x, expire = cache.Get(key, duration)
	is.Equal(int64(1), x)

	x, expire = cache.Reset(key, duration)
	is.Equal(int64(0), x)

	x, expire = cache.Get(key, duration)
	is.Equal(int64(1), x)

}
