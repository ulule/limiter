package memory_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/store/memory"
)

func TestCacheIncrementSequential(t *testing.T) {
	clock := time.Unix(1533930608, 0)
	limiter.Now = func() time.Time {
		return clock
	}
	defer func() {
		limiter.Now = time.Now
	}()

	is := require.New(t)

	key := "foobar"
	cache := memory.NewCache(10 * time.Nanosecond)
	duration := 50 * time.Millisecond
	deleted := limiter.Now().Add(duration).UnixNano()
	epsilon := 0.001

	x, expire := cache.Increment(key, 1, duration)
	is.Equal(int64(1), x)
	is.InEpsilon(deleted, expire.UnixNano(), epsilon)

	x, expire = cache.Increment(key, 2, duration)
	is.Equal(int64(3), x)
	is.InEpsilon(deleted, expire.UnixNano(), epsilon)

	clock = clock.Add(duration + 1)

	deleted = limiter.Now().Add(duration).UnixNano()
	x, expire = cache.Increment(key, 1, duration)
	is.Equal(int64(1), x)
	is.InEpsilon(deleted, expire.UnixNano(), epsilon)
}

func TestCacheIncrementConcurrent(t *testing.T) {
	clock := time.Unix(1533930608, 0)
	limiter.Now = func() time.Time {
		return clock
	}
	defer func() {
		limiter.Now = time.Now
	}()
	var clockMutex sync.Mutex

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

	clocks := make([]time.Time, goroutines)
	for i := range clocks {
		clocks[i] = clock
	}

	wg := &sync.WaitGroup{}
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			c := &clocks[i]
			if (i % 3) != 0 {
				*c = c.Add(50 * time.Millisecond)
				for j := 0; j < 500; j++ {
					*c = c.Add(1 * time.Millisecond)
					clockMutex.Lock()
					t := clock
					clock = *c
					cache.Increment(key, int64(i), (75 * time.Millisecond))
					clock = t
					clockMutex.Unlock()
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	wg = &sync.WaitGroup{}
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			c := &clocks[i]
			if (i % 3) == 0 {
				*c = c.Add(1 * time.Second)
				for j := 0; j < ops; j++ {
					*c = c.Add(1 * time.Millisecond)
					clockMutex.Lock()
					t := clock
					clock = *c
					cache.Increment(key, int64(i+j), (1 * time.Second))
					clock = t
					clockMutex.Unlock()
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	for _, t := range clocks {
		if t.After(clock) {
			clock = t
		}
	}

	value, expire := cache.Get(key, (100 * time.Millisecond))
	is.Equal(expected, value)
	is.True(limiter.Now().Before(expire))
}

func TestCacheGet(t *testing.T) {
	clock := time.Unix(1533930608, 0)
	limiter.Now = func() time.Time {
		return clock
	}
	defer func() {
		limiter.Now = time.Now
	}()

	is := require.New(t)

	key := "foobar"
	cache := memory.NewCache(10 * time.Nanosecond)
	duration := 50 * time.Millisecond
	deleted := limiter.Now().Add(duration).UnixNano()
	epsilon := 0.001

	x, expire := cache.Get(key, duration)
	is.Equal(int64(0), x)
	is.InEpsilon(deleted, expire.UnixNano(), epsilon)

}
