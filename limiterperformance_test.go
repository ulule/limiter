package limiter_test

import (
	"context"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stackimpact/stackimpact-go"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

type rateLimiter struct {
	g *limiter.Limiter
	s *limiter.Limiter
	a *limiter.Limiter
}

const (
	max     = math.MaxInt64
	maxRand = max / 4
)

var (
	agentKey = flag.String("agentkey", "", "Stack impact agent key")
)

// Usage: go test -run TestMemoryStorePerformance -agentkey=stackimpact_hash -timeout 24h
// remember to increase your fd
func TestMemoryStorePerformance(t *testing.T) {
	_, err := LiftRLimits()
	PanicOnError(err)

	flag.Parse()
	stackimpact.Start(stackimpact.Options{
		AgentKey: *agentKey,
		AppName:  "limiter",
	})

	r := &rateLimiter{}
	r.g = newLimiter("1-S")
	r.a = newLimiter("3-S")
	r.s = newLimiter("15-S")

	var (
		wg sync.WaitGroup
		i  int64
	)
	for i = 0; i < max; i++ {
		rd := RandomNumber(1, maxRand)
		go func(x int64) {
			wg.Add(1)
			time.Sleep(time.Duration(RandomNumber(1, 120)) * time.Second)
			r.rateLimit(x, fmt.Sprintf("%d", rd), context.Background())
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func (r *rateLimiter) rateLimit(i int64, k string, ctx context.Context) {
	var (
		limiterCtx limiter.Context
		err        error
	)
	fmt.Printf("%d requests, key: %s \n", i, k)
	if i%5 == 0 {
		limiterCtx, err = r.s.Get(ctx, k)
	} else if i%3 == 0 {
		limiterCtx, err = r.a.Get(ctx, k)
	} else {
		limiterCtx, err = r.g.Get(ctx, k)
	}

	PanicOnError(err)

	if limiterCtx.Reached {
		return
	}

}

func newLimiter(rate string) *limiter.Limiter {
	r, err := limiter.NewRateFromFormatted(rate)
	PanicOnError(err)
	return limiter.New(memory.NewStoreWithOptions(limiter.StoreOptions{
		Prefix:          limiter.DefaultPrefix,
		CleanUpInterval: 5 * time.Second,
	}), r)
}

func RandomNumber(min int64, max int64) int64 {
	if min == max {
		return min
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min) + min
}

func LiftRLimits() (rLimit syscall.Rlimit, err error) {
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	rLimit.Cur = rLimit.Max
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &syscall.Rlimit{Cur: 1048576, Max: rLimit.Max})
		if err != nil {
			return
		}
	}
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	return
}

func PanicOnError(e error) {
	if e != nil {
		panic(e)
	}
}
