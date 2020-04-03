package redis_test

import (
	"os"
	"testing"
	"time"

	libredis "github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/require"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/redis"
	"github.com/ulule/limiter/v3/drivers/store/tests"
)

func TestRedisStoreSequentialAccess(t *testing.T) {
	is := require.New(t)

	client, err := newRedisClient()
	is.NoError(err)
	is.NotNil(client)

	store, err := redis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "limiter:redis:sequential",
		MaxRetry: 3,
	})
	is.NoError(err)
	is.NotNil(store)

	tests.TestStoreSequentialAccess(t, store)
}

func TestRedisStoreConcurrentAccess(t *testing.T) {
	is := require.New(t)

	client, err := newRedisClient()
	is.NoError(err)
	is.NotNil(client)

	store, err := redis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "limiter:redis:concurrent",
		MaxRetry: 7,
	})
	is.NoError(err)
	is.NotNil(store)

	tests.TestStoreConcurrentAccess(t, store)
}

func TestRedisClientExpiration(t *testing.T) {
	is := require.New(t)

	client, err := newRedisClient()
	is.NoError(err)
	is.NotNil(client)

	key := "foobar"
	value := 642
	keyNoExpiration := -1 * time.Nanosecond
	keyNotExist := -2 * time.Nanosecond

	delCmd := client.Del(key)
	_, err = delCmd.Result()
	is.NoError(err)

	expCmd := client.PTTL(key)
	ttl, err := expCmd.Result()
	is.NoError(err)
	is.Equal(keyNotExist, ttl)

	setCmd := client.Set(key, value, 0)
	_, err = setCmd.Result()
	is.NoError(err)

	expCmd = client.PTTL(key)
	ttl, err = expCmd.Result()
	is.NoError(err)
	is.Equal(keyNoExpiration, ttl)

	setCmd = client.Set(key, value, time.Second)
	_, err = setCmd.Result()
	is.NoError(err)

	time.Sleep(100 * time.Millisecond)

	expCmd = client.PTTL(key)
	ttl, err = expCmd.Result()
	is.NoError(err)

	expected := int64(0)
	actual := int64(ttl)
	is.Greater(actual, expected)
}

func newRedisClient() (*libredis.Client, error) {
	uri := "redis://localhost:6379/0"
	if os.Getenv("REDIS_URI") != "" {
		uri = os.Getenv("REDIS_URI")
	}

	opt, err := libredis.ParseURL(uri)
	if err != nil {
		return nil, err
	}

	client := libredis.NewClient(opt)
	return client, nil
}
