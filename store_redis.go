package limiter

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

// RedisBucket is a redis bucket.
type RedisBucket struct {
	id    string
	value int64
}

// RedisStore is the redis store.
type RedisStore struct {
	Pool            *redis.Pool
	PrefixQuota     string
	PrefixRemaining string
	PrefixUsed      string
	PrefixReset     string
}

// NewRedisStore returns an instance of redis store.
func NewRedisStore(pool *redis.Pool, prefix string) (*RedisStore, error) {
	if prefix == "" {
		prefix = "ratelimit"
	}

	store := &RedisStore{
		Pool:            pool,
		PrefixQuota:     fmt.Sprintf("%s:quota:", prefix),
		PrefixRemaining: fmt.Sprintf("%s:remaining:", prefix),
		PrefixUsed:      fmt.Sprintf("%s:used:", prefix),
		PrefixReset:     fmt.Sprintf("%s:reset:", prefix),
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
func (s *RedisStore) Get(id string, rate Rate) (Context, error) {
	var (
		context Context
		err     error
		reply   []interface{}
	)

	conn := s.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return context, err
	}

	quota := &RedisBucket{id: s.PrefixQuota + id}
	remaining := &RedisBucket{id: s.PrefixRemaining + id}
	used := &RedisBucket{id: s.PrefixUsed + id}
	reset := &RedisBucket{id: s.PrefixReset + id}

	millisecond := int64(time.Millisecond)
	expiry := (time.Now().UnixNano()/millisecond + int64(rate.Period)/millisecond) / 1000

	conn.Send("WATCH", remaining)
	defer conn.Send("UNWATCH", remaining)

	reply, err = redis.Values(conn.Do("MGET", quota.id, remaining.id, used.id, reset.id))
	if err != nil {
		return context, err
	}

	if _, err = redis.Scan(reply, &quota.value, &remaining.value, &used.value, &reset.value); err != nil {
		return context, err
	}

	reached := false

	if quota.value == 0 {
		conn.Send("MULTI")
		conn.Send("SET", quota.id, rate.Limit, "EX", rate.Period.Seconds(), "NX")
		conn.Send("SET", remaining.id, rate.Limit-1, "EX", rate.Period.Seconds(), "NX")
		conn.Send("SET", used.id, 1, "EX", rate.Period.Seconds(), "NX")
		conn.Send("SET", reset.id, expiry, "EX", rate.Period.Seconds(), "NX")

		if reply, err = redis.Values(conn.Do("EXEC")); err != nil {
			return context, err
		}

		quota.value = rate.Limit
		remaining.value = rate.Limit - 1
		used.value = 1
		reset.value = expiry

	} else if remaining.value > 0 {
		conn.Do("DECR", remaining.id)
		remaining.value--

		conn.Do("INCR", used.id)
		used.value++

	} else if remaining.value == 0 {
		if used.value == quota.value {
			reached = true
		}
	}

	return Context{
		Limit:     quota.value,
		Remaining: remaining.value,
		Used:      used.value,
		Reset:     reset.value,
		Reached:   reached,
	}, nil
}
