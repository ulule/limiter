package limiter

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

// RedisStoreFunc is a redis store function.
type RedisStoreFunc func(c redis.Conn, key string, rate Rate) ([]int, error)

// RedisStoreOptions are options for Redis store.
type RedisStoreOptions struct {
	// The prefix to use for the key.
	Prefix string

	// The maximum number of retry under race conditions.
	MaxRetry int
}

// RedisStore is the redis store.
type RedisStore struct {
	// The prefix to use for the key.
	Prefix string

	// github.com/garyburd/redigo Pool instance.
	Pool *redis.Pool

	// The maximum number of retry under race conditions.
	MaxRetry int
}

// NewRedisStore returns an instance of redis store.
func NewRedisStore(pool *redis.Pool) (Store, error) {
	store := &RedisStore{
		Pool:     pool,
		Prefix:   DefaultPrefix,
		MaxRetry: DefaultMaxRetry,
	}

	if _, err := store.ping(); err != nil {
		return nil, err
	}

	return store, nil
}

// NewRedisStoreWithOptions returns an instance of redis store with custom options.
func NewRedisStoreWithOptions(pool *redis.Pool, options RedisStoreOptions) (Store, error) {
	if options.Prefix == "" {
		options.Prefix = DefaultPrefix
	}

	if options.MaxRetry == 0 {
		options.MaxRetry = DefaultMaxRetry
	}

	store := &RedisStore{
		Pool:     pool,
		Prefix:   options.Prefix,
		MaxRetry: options.MaxRetry,
	}

	if _, err := store.ping(); err != nil {
		return nil, err
	}

	return store, nil
}

// ping checks if redis is alive.
func (s *RedisStore) ping() (bool, error) {
	conn := s.Pool.Get()
	defer conn.Close()

	data, err := conn.Do("PING")
	if err != nil || data == nil {
		return false, err
	}

	return (data == "PONG"), nil
}

func (s RedisStore) do(f RedisStoreFunc, c redis.Conn, key string, rate Rate) ([]int, error) {
	for i := 1; i <= s.MaxRetry; i++ {
		values, err := f(c, key, rate)
		if err == nil && len(values) != 0 {
			return values, nil
		}
	}
	return nil, fmt.Errorf("retry limit exceeded")
}

func (s RedisStore) setRate(c redis.Conn, key string, rate Rate) ([]int, error) {
	c.Send("MULTI")
	c.Send("SETNX", key, 1)
	c.Send("EXPIRE", key, rate.Period.Seconds())
	return redis.Ints(c.Do("EXEC"))
}

func (s RedisStore) updateRate(c redis.Conn, key string, rate Rate) ([]int, error) {
	c.Send("MULTI")
	c.Send("INCR", key)
	c.Send("TTL", key)
	return redis.Ints(c.Do("EXEC"))
}

// Get returns the limit for the identifier.
func (s RedisStore) Get(key string, rate Rate) (Context, error) {
	var (
		err    error
		values []int
	)

	ctx := Context{}
	key = fmt.Sprintf("%s:%s", s.Prefix, key)

	c := s.Pool.Get()
	defer c.Close()
	if err := c.Err(); err != nil {
		return Context{}, err
	}

	c.Do("WATCH", key)
	defer c.Do("UNWATCH", key)

	values, err = s.do(s.setRate, c, key, rate)
	if err != nil {
		return ctx, err
	}

	created := (values[0] == 1)
	ms := int64(time.Millisecond)

	if created {
		return Context{
			Limit:     rate.Limit,
			Remaining: rate.Limit - 1,
			Reset:     (time.Now().UnixNano()/ms + int64(rate.Period)/ms) / 1000,
			Reached:   false,
		}, nil
	}

	values, err = s.do(s.updateRate, c, key, rate)
	if err != nil {
		return ctx, err
	}

	count := int64(values[0])
	ttl := int64(values[1])
	remaining := int64(0)

	if count < rate.Limit {
		remaining = rate.Limit - count
	}

	return Context{
		Limit:     rate.Limit,
		Remaining: remaining,
		Reset:     time.Now().Add(time.Duration(ttl) * time.Second).Unix(),
		Reached:   count > rate.Limit,
	}, nil
}
