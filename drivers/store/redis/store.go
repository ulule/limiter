package redis

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
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
	// MaxRetry is the maximum number of retry under race conditions.
	// Deprecated: this option is no longer required since all operations are atomic now.
	MaxRetry int
	// client used to communicate with redis server.
	client Client
	// luaMutex is a mutex used to avoid concurrent access on luaIncrSHA and luaPeekSHA.
	luaMutex sync.RWMutex
	// luaLoaded is used for CAS and reduce pressure on luaMutex.
	luaLoaded uint32
	// luaIncrSHA is the SHA of increase and expire key script.
	luaIncrSHA string
	// luaPeekSHA is the SHA of peek and expire key script.
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

	err := store.preloadLuaScripts(context.Background())
	if err != nil {
		return nil, err
	}

	return store, nil
}

// Increment increments the limit by given count & gives back the new limit for given identifier
func (store *Store) Increment(ctx context.Context, key string, count int64, rate limiter.Rate) (limiter.Context, error) {
	key = fmt.Sprintf("%s:%s", store.Prefix, key)
	cmd := store.evalSHA(ctx, store.getLuaIncrSHA, []string{key}, count, rate.Period.Milliseconds())
	return currentContext(cmd, rate)
}

// Get returns the limit for given identifier.
func (store *Store) Get(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	key = fmt.Sprintf("%s:%s", store.Prefix, key)
	cmd := store.evalSHA(ctx, store.getLuaIncrSHA, []string{key}, 1, rate.Period.Milliseconds())
	return currentContext(cmd, rate)
}

// Peek returns the limit for given identifier, without modification on current values.
func (store *Store) Peek(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	key = fmt.Sprintf("%s:%s", store.Prefix, key)
	cmd := store.evalSHA(ctx, store.getLuaPeekSHA, []string{key})
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
	_, err := store.client.Del(ctx, key).Result()
	if err != nil {
		return limiter.Context{}, err
	}

	count := int64(0)
	now := time.Now()
	expiration := now.Add(rate.Period)

	return common.GetContextFromState(now, rate, expiration, count), nil
}

// preloadLuaScripts preloads the "incr" and "peek" lua scripts.
func (store *Store) preloadLuaScripts(ctx context.Context) error {
	// Verify if we need to load lua scripts.
	// Inspired by sync.Once.
	if atomic.LoadUint32(&store.luaLoaded) == 0 {
		return store.loadLuaScripts(ctx)
	}
	return nil
}

// reloadLuaScripts forces a reload of "incr" and "peek" lua scripts.
func (store *Store) reloadLuaScripts(ctx context.Context) error {
	// Reset lua scripts loaded state.
	// Inspired by sync.Once.
	atomic.StoreUint32(&store.luaLoaded, 0)
	return store.loadLuaScripts(ctx)
}

// loadLuaScripts load "incr" and "peek" lua scripts.
// WARNING: Please use preloadLuaScripts or reloadLuaScripts, instead of this one.
func (store *Store) loadLuaScripts(ctx context.Context) error {
	store.luaMutex.Lock()
	defer store.luaMutex.Unlock()

	// Check if scripts are already loaded.
	if atomic.LoadUint32(&store.luaLoaded) != 0 {
		return nil
	}

	luaIncrSHA, err := store.client.ScriptLoad(ctx, luaIncrScript).Result()
	if err != nil {
		return errors.Wrap(err, `failed to load "incr" lua script`)
	}

	luaPeekSHA, err := store.client.ScriptLoad(ctx, luaPeekScript).Result()
	if err != nil {
		return errors.Wrap(err, `failed to load "peek" lua script`)
	}

	store.luaIncrSHA = luaIncrSHA
	store.luaPeekSHA = luaPeekSHA

	atomic.StoreUint32(&store.luaLoaded, 1)

	return nil
}

// getLuaIncrSHA returns a "thread-safe" value for luaIncrSHA.
func (store *Store) getLuaIncrSHA() string {
	store.luaMutex.RLock()
	defer store.luaMutex.RUnlock()
	return store.luaIncrSHA
}

// getLuaPeekSHA returns a "thread-safe" value for luaPeekSHA.
func (store *Store) getLuaPeekSHA() string {
	store.luaMutex.RLock()
	defer store.luaMutex.RUnlock()
	return store.luaPeekSHA
}

// evalSHA eval the redis lua sha and load the scripts if missing.
func (store *Store) evalSHA(ctx context.Context, getSha func() string,
	keys []string, args ...interface{}) *libredis.Cmd {

	cmd := store.client.EvalSha(ctx, getSha(), keys, args...)
	err := cmd.Err()
	if err == nil || !isLuaScriptGone(err) {
		return cmd
	}

	err = store.reloadLuaScripts(ctx)
	if err != nil {
		cmd = libredis.NewCmd(ctx)
		cmd.SetErr(err)
		return cmd
	}

	return store.client.EvalSha(ctx, getSha(), keys, args...)
}

// isLuaScriptGone returns if the error is a missing lua script from redis server.
func isLuaScriptGone(err error) bool {
	return strings.HasPrefix(err.Error(), "NOSCRIPT")
}

// parseCountAndTTL parse count and ttl from lua script output.
func parseCountAndTTL(cmd *libredis.Cmd) (int64, int64, error) {
	result, err := cmd.Result()
	if err != nil {
		return 0, 0, errors.Wrap(err, "an error has occurred with redis command")
	}

	fields, ok := result.([]interface{})
	if !ok || len(fields) != 2 {
		return 0, 0, errors.New("two elements in result were expected")
	}

	count, ok1 := fields[0].(int64)
	ttl, ok2 := fields[1].(int64)
	if !ok1 || !ok2 {
		return 0, 0, errors.New("type of the count and/or ttl should be number")
	}

	return count, ttl, nil
}

func currentContext(cmd *libredis.Cmd, rate limiter.Rate) (limiter.Context, error) {
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
