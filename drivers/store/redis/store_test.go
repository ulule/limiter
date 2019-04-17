package redis_test

import (
	"os"
	"testing"

	libredis "github.com/go-redis/redis"
	"github.com/stretchr/testify/require"

	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/store/redis"
	"github.com/ulule/limiter/drivers/store/tests"
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
