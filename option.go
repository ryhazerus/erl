package erl

import "github.com/ryhazerus/erl/store"

// Option configures the Limiter.
type Option func(*Limiter)

// WithStore sets the backing store for rate limit counters.
// If not provided, an in-memory store is used by default.
func WithStore(s store.Store) Option {
	return func(l *Limiter) {
		l.store = s
	}
}

// WithOnLimitReached sets a callback that fires when a resource's limit is reached.
// This is always called regardless of strategy, but is the primary mechanism for
// LogOnly resources.
func WithOnLimitReached(fn func(Resource, int64)) Option {
	return func(l *Limiter) {
		l.onLimitReached = fn
	}
}
