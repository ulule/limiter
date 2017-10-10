package limiter

import (
	"context"
)

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

// New returns an instance of Limiter.
func New(store Store, rate Rate) *Limiter {
	return &Limiter{
		Store: store,
		Rate:  rate,
	}
}

// Get returns the limit for given identifier.
func (limiter *Limiter) Get(ctx context.Context, key string) (Context, error) {
	return limiter.Store.Get(ctx, key, limiter.Rate)
}

// Peek returns the limit for given identifier, without modification on current values.
func (limiter *Limiter) Peek(ctx context.Context, key string) (Context, error) {
	return limiter.Store.Peek(ctx, key, limiter.Rate)
}
