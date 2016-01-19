package limiter

import (
	"fmt"
	"time"

	cache "github.com/pmylund/go-cache"
)

// MemoryStore is the in-memory store.
type MemoryStore struct {
	Cache  *cache.Cache
	Prefix string
}

// NewMemoryStore creates a new instance of memory store with defaults.
func NewMemoryStore() Store {
	return NewMemoryStoreWithOptions(StoreOptions{
		Prefix:          DefaultPrefix,
		CleanUpInterval: DefaultCleanUpInterval,
	})
}

// NewMemoryStoreWithOptions creates a new instance of memory store with options.
func NewMemoryStoreWithOptions(options StoreOptions) Store {
	return &MemoryStore{
		Prefix: options.Prefix,
		Cache:  cache.New(cache.NoExpiration, options.CleanUpInterval),
	}
}

// Get implement Store.Get() method.
func (s *MemoryStore) Get(key string, rate Rate) (Context, error) {
	ctx := Context{}
	key = fmt.Sprintf("%s:%s", s.Prefix, key)
	item, found := s.Cache.Items()[key]
	ms := int64(time.Millisecond)
	now := time.Now()

	if !found || item.Expired() {
		s.Cache.Set(key, int64(1), rate.Period)

		return Context{
			Limit:     rate.Limit,
			Remaining: rate.Limit - 1,
			Reset:     (now.UnixNano()/ms + int64(rate.Period)/ms) / 1000,
			Reached:   false,
		}, nil
	}

	count, err := s.Cache.IncrementInt64(key, 1)
	if err != nil {
		return ctx, err
	}

	remaining := int64(0)
	if count < rate.Limit {
		remaining = rate.Limit - count
	}

	expire := time.Unix(0, item.Expiration)
	reset := expire.Add(time.Duration(expire.Sub(now).Seconds()) * time.Second).Unix()

	return Context{
		Limit:     rate.Limit,
		Remaining: remaining,
		Reset:     reset,
		Reached:   count > rate.Limit,
	}, nil
}
