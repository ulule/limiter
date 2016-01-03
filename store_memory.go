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

	if !found || item.Expired() {
		s.Cache.Set(key, int64(1), rate.Period)

		return Context{
			Limit:     rate.Limit,
			Remaining: rate.Limit - 1,
			Reset:     (time.Now().UnixNano()/ms + int64(rate.Period)/ms) / 1000,
			Reached:   false,
		}, nil
	}

	err := s.Cache.Increment(key, int64(1))

	if err != nil {
		return ctx, nil
	}

	remaining := int64(0)

	count := item.Object.(int64)

	if count < rate.Limit {
		remaining = rate.Limit - count
	}

	return Context{
		Limit:     rate.Limit,
		Remaining: remaining,
		Reset:     time.Now().Add(time.Duration(item.Expiration) * time.Second).Unix(),
		Reached:   count > rate.Limit,
	}, nil
}
