package limiter

const (
	// RedisDefaultPrefix is the prefix to use for the key in Redis store.
	RedisDefaultPrefix = "limiter"

	// RedisDefaultMaxRetry is the maximum number of retries under race condition
	// for Redis store.
	RedisDefaultMaxRetry = 3
)
