package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/ulule/limiter"
)

// Store is the in-memory store.
type Store struct {
	// Prefix used for the key.
	Prefix string
	// cache used to store values in-memory.
	cache *CacheWrapper
}

// NewStore creates a new instance of memory store with defaults.
func NewStore() limiter.Store {
	return NewStoreWithOptions(limiter.StoreOptions{
		Prefix:          limiter.DefaultPrefix,
		CleanUpInterval: limiter.DefaultCleanUpInterval,
	})
}

// NewStoreWithOptions creates a new instance of memory store with options.
func NewStoreWithOptions(options limiter.StoreOptions) limiter.Store {
	return &Store{
		Prefix: options.Prefix,
		cache:  NewCache(options.CleanUpInterval),
	}
}

// Get returns the limit for given identifier.
func (store *Store) Get(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	key = fmt.Sprintf("%s:%s", store.Prefix, key)
	now := time.Now()

	count, expiration := store.cache.Increment(key, 1, rate.Period)

	lctx := store.getContextFromState(now, rate, expiration, count)
	return lctx, nil
}

// Peek returns the limit for given identifier, without modification on current values.
func (store *Store) Peek(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	key = fmt.Sprintf("%s:%s", store.Prefix, key)
	now := time.Now()

	count, expiration := store.cache.Get(key, rate.Period)

	lctx := store.getContextFromState(now, rate, expiration, count)
	return lctx, nil
}

func (store *Store) getContextFromState(now time.Time, rate limiter.Rate,
	expiration time.Time, count int64) limiter.Context {

	limit := rate.Limit
	remaining := int64(0)
	reached := true

	if count <= limit {
		remaining = limit - count
		reached = false
	}

	reset := expiration.Add(time.Duration(expiration.Sub(now).Seconds()) * time.Second).Unix()

	return limiter.Context{
		Limit:     limit,
		Remaining: remaining,
		Reset:     reset,
		Reached:   reached,
	}
}
