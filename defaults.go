package limiter

const (
	// DefaultPrefix is the prefix to use for the store key.
	DefaultPrefix = "limiter"

	// DefaultMaxRetry is the maximum number of key retries under race condition.
	DefaultMaxRetry = 3
)
