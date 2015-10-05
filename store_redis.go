package limiter

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

// RedisStore is the redis store.
type RedisStore struct {
	Prefix string
	Pool   *redis.Pool
}

// NewRedisStore returns an instance of redis store.
func NewRedisStore(pool *redis.Pool, prefix string) (*RedisStore, error) {
	if prefix == "" {
		prefix = "ratelimit"
	}

	store := &RedisStore{
		Pool:   pool,
		Prefix: prefix,
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

// Get returns the limit for the identifier.
func (s *RedisStore) Get(key string, rate Rate) (Context, error) {
	ctx := Context{}
	key = fmt.Sprintf("%s:%s", s.Prefix, key)

	c := s.Pool.Get()
	defer c.Close()
	if err := c.Err(); err != nil {
		return Context{}, err
	}

	exists, err := redis.Bool(c.Do("EXISTS", key))
	if err != nil {
		return ctx, err
	}

	ms := int64(time.Millisecond)
	if !exists {
		c.Do("SET", key, 1, "EX", rate.Period.Seconds())
		return Context{
			Limit:     rate.Limit,
			Remaining: rate.Limit - 1,
			Reset:     (time.Now().UnixNano()/ms + int64(rate.Period)/ms) / 1000,
			Reached:   false,
		}, nil
	}

	count, err := redis.Int64(c.Do("INCR", key))
	if err != nil {
		return ctx, nil
	}

	pttl, err := redis.Int64(c.Do("PTTL", key))
	if err != nil {
		return ctx, nil
	}

	remaining := int64(0)
	if count < rate.Limit {
		remaining = rate.Limit - count
	}

	return Context{
		Limit:     rate.Limit,
		Remaining: remaining,
		Reset:     time.Now().Add(time.Duration(pttl) * time.Millisecond).Unix(),
		Reached:   count > rate.Limit,
	}, nil
}
