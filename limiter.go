package limiter

// -----------------------------------------------------------------
// Store
// -----------------------------------------------------------------

// Store is the common interface for limiter stores.
type Store interface {
	Get(key string, rate Rate) (Context, error)
}

// -----------------------------------------------------------------
// Context
// -----------------------------------------------------------------

// Context is the limit context.
type Context struct {
	Limit     int64
	Remaining int64
	Used      int64
	Reset     int64
	Reached   bool
}

// -----------------------------------------------------------------
// Limiter
// -----------------------------------------------------------------

// Limiter is the limiter instance.
type Limiter struct {
	Store Store
	Rate  Rate
}

// NewLimiter returns an instance of ratelimit.
func NewLimiter(store Store, rate Rate) *Limiter {
	return &Limiter{
		Store: store,
		Rate:  rate,
	}
}

// Get returns the limit for the identifier.
func (l *Limiter) Get(key string) (Context, error) {
	return l.Store.Get(key, l.Rate)
}
