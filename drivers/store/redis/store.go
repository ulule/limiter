package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	libredis "github.com/go-redis/redis/v8"
	"github.com/pkg/errors"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/common"
)

const (
	luaIncrScript = `
local key = KEYS[1]
local count = tonumber(ARGV[1])
local ttl = tonumber(ARGV[2])
local ret = redis.call("incrby", key, ARGV[1])
if ret == count then
	if ttl > 0 then
		redis.call("pexpire", key, ARGV[2])
	end
	return {ret, ttl}
end
ttl = redis.call("pttl", key)
return {ret, ttl}
`
	luaPeekScript = `
local key = KEYS[1]
local v = redis.call("get", key)
if v == false then
	return {0, 0}
end
local ttl = redis.call("pttl", key)
return {tonumber(v), ttl}
`
)

// Client is an interface thats allows to use a redis cluster or a redis single client seamlessly.
type Client interface {
	Get(ctx context.Context, key string) *libredis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *libredis.StatusCmd
	Watch(ctx context.Context, handler func(*libredis.Tx) error, keys ...string) error
	Del(ctx context.Context, keys ...string) *libredis.IntCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *libredis.BoolCmd
	EvalSha(ctx context.Context, sha string, keys []string, args ...interface{}) *libredis.Cmd
	ScriptLoad(ctx context.Context, script string) *libredis.StringCmd
}

// Store is the redis store.
type Store struct {
	// Prefix used for the key.
	Prefix string
	// deprecated, this option make no sense when all operations were atomic
	// MaxRetry is the maximum number of retry under race conditions.
	MaxRetry int
	// client used to communicate with redis server.
	client Client
	// luaIncrSHA is the SHA of increase and expire key script
	luaIncrSHA string
	// luaPeekSHA is the SHA of peek and expire key script
	luaPeekSHA string
}

// NewStore returns an instance of redis store with defaults.
func NewStore(client Client) (limiter.Store, error) {
	return NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:          limiter.DefaultPrefix,
		CleanUpInterval: limiter.DefaultCleanUpInterval,
		MaxRetry:        limiter.DefaultMaxRetry,
	})
}

// NewStoreWithOptions returns an instance of redis store with options.
func NewStoreWithOptions(client Client, options limiter.StoreOptions) (limiter.Store, error) {
	store := &Store{
		client:   client,
		Prefix:   options.Prefix,
		MaxRetry: options.MaxRetry,
	}

	if store.MaxRetry <= 0 {
		store.MaxRetry = 1
	}
	if err := store.preloadLuaScripts(context.Background()); err != nil {
		return nil, err
	}
	return store, nil
}

// preloadLuaScripts would preload the  incr and peek lua script
func (store *Store) preloadLuaScripts(ctx context.Context) error {
	incrLuaSHA, err := store.client.ScriptLoad(ctx, luaIncrScript).Result()
	if err != nil {
		return errors.Wrap(err, "failed to load incr lua script")
	}
	peekLuaSHA, err := store.client.ScriptLoad(ctx, luaPeekScript).Result()
	if err != nil {
		return errors.Wrap(err, "failed to load peek lua script")
	}
	store.luaIncrSHA = incrLuaSHA
	store.luaPeekSHA = peekLuaSHA
	return nil
}

// Get returns the limit for given identifier.
func (store *Store) Get(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	key = fmt.Sprintf("%s:%s", store.Prefix, key)
	cmd := store.evalSHA(ctx, store.luaIncrSHA, []string{key}, 1, rate.Period.Milliseconds())
	count, ttl, err := parseCountAndTTL(cmd)
	if err != nil {
		return limiter.Context{}, err
	}
	now := time.Now()
	expiration := now.Add(rate.Period)
	if ttl > 0 {
		expiration = now.Add(time.Duration(ttl) * time.Millisecond)
	}
	return common.GetContextFromState(now, rate, expiration, count), nil
}

// Peek returns the limit for given identifier, without modification on current values.
func (store *Store) Peek(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	key = fmt.Sprintf("%s:%s", store.Prefix, key)
	cmd := store.evalSHA(ctx, store.luaPeekSHA, []string{key})
	count, ttl, err := parseCountAndTTL(cmd)
	if err != nil {
		return limiter.Context{}, err
	}
	now := time.Now()
	expiration := now.Add(rate.Period)
	if ttl > 0 {
		expiration = now.Add(time.Duration(ttl) * time.Millisecond)
	}
	return common.GetContextFromState(now, rate, expiration, count), nil
}

// Reset returns the limit for given identifier which is set to zero.
func (store *Store) Reset(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	key = fmt.Sprintf("%s:%s", store.Prefix, key)
	if _, err := store.client.Del(ctx, key).Result(); err != nil {
		return limiter.Context{}, err
	}
	count := int64(0)
	now := time.Now()
	expiration := now.Add(rate.Period)
	return common.GetContextFromState(now, rate, expiration, count), nil
}

// evalSHA eval the redis lua sha and load the script if missing
func (store *Store) evalSHA(ctx context.Context, sha string, keys []string, args ...interface{}) *libredis.Cmd {
	cmd := store.client.EvalSha(ctx, sha, keys, args...)
	if err := cmd.Err(); err != nil {
		if !isLuaScriptGone(err) {
			return cmd
		}
		if err := store.preloadLuaScripts(ctx); err != nil {
			cmd = libredis.NewCmd(ctx)
			cmd.SetErr(err)
			return cmd
		}
		cmd = store.client.EvalSha(ctx, sha, keys)
	}
	return cmd
}

// isLuaScriptGone check whether the error was no script or no
func isLuaScriptGone(err error) bool {
	return strings.HasPrefix(err.Error(), "NOSCRIPT")
}

// parseCountAndTTL parse count and ttl from lua script output
func parseCountAndTTL(cmd *libredis.Cmd) (int64, int64, error) {
	ret, err := cmd.Result()
	if err != nil {
		return 0, 0, err
	}
	if fields, ok := ret.([]interface{}); !ok || len(fields) != 2 {
		return 0, 0, errors.New("two elements in array was expected")
	}
	fields := ret.([]interface{})
	count, ok1 := fields[0].(int64)
	ttl, ok2 := fields[1].(int64)
	if !ok1 || !ok2 {
		return 0, 0, errors.New("type of the count and ttl should be number")
	}
	return count, ttl, nil
}
