package store

import (
	"context"
	"time"
)

// Window mirrors erl.Window so the store package doesn't import the parent.
// Callers pass the window's duration and bucket key instead.
type Window struct {
	Duration   time.Duration
	BucketKey  string
	BucketStart time.Time
}

// Store defines the interface for rate limit counter backends.
type Store interface {
	// Increment atomically increments the counter for the given key in the
	// current window bucket and returns the new count.
	Increment(ctx context.Context, key string, w Window) (current int64, err error)

	// Get returns the current counter value for the key in the active window bucket.
	Get(ctx context.Context, key string, w Window) (current int64, err error)

	// Reset removes the counter for the given key.
	Reset(ctx context.Context, key string) error

	// Close releases any resources held by the store.
	Close() error
}
