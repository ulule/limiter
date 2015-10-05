package limiter

import (
	"fmt"
	cache "github.com/pmylund/go-cache"
	"time"
)

type MemoryStore struct {
	Cache  *cache.Cache
	Prefix string
}

func NewMemoryStore(prefix string, cleanupInterval time.Duration) Store {

	cache := cache.New(cache.NoExpiration, cleanupInterval)

	return &MemoryStore{
		Prefix: prefix,
		Cache:  cache,
	}
}

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
		Reset:     time.Now().Add(time.Duration(item.Expiration.Second()) * time.Second).Unix(),
		Reached:   count > rate.Limit,
	}, nil
}
