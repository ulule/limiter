package limiter

// -----------------------------------------------------------------
// Context
// -----------------------------------------------------------------

// Context is the limit context.
type Context struct {
	Limit     int64
	Remaining int64
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

// NewLimiter returns an instance of Limiter.
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

// Peek returns the limit for identifier without impacting accounting
func (l *Limiter) Peek(key string) (Context, error) {
	return l.Store.Peek(key, l.Rate)
}
