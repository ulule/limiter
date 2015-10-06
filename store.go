package limiter

// Store is the common interface for limiter stores.
type Store interface {
	Get(key string, rate Rate) (Context, error)
}

// StoreOptions are options for store.
type StoreOptions struct {
	// The prefix to use for the key.
	Prefix string

	// The maximum number of retry under race conditions.
	MaxRetry int
}
