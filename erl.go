package erl

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ryhazerus/erl/store"
)

// ErrLimitExceeded is returned when a request is blocked due to rate limiting.
var ErrLimitExceeded = errors.New("erl: rate limit exceeded")

// LimitExceededError provides details about which resource hit its limit
// and supports waiting for the window to reset (BlockWithQueue strategy).
type LimitExceededError struct {
	Resource Resource
	Current  int64
	resetAt  time.Time
}

func (e *LimitExceededError) Error() string {
	return fmt.Sprintf("erl: rate limit exceeded for %s (%d/%d)", e.Resource.Name, e.Current, e.Resource.Limit)
}

func (e *LimitExceededError) Unwrap() error {
	return ErrLimitExceeded
}

// Wait blocks until the current window resets or the context is cancelled.
// This is intended for use with the BlockWithQueue strategy.
func (e *LimitExceededError) Wait(ctx context.Context) error {
	delay := time.Until(e.resetAt)
	if delay <= 0 {
		return nil
	}
	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// Limiter is the main entry point for the erl library. It tracks outgoing HTTP
// requests against registered resources and enforces configurable rate limits.
type Limiter struct {
	mu             sync.RWMutex
	resources      []Resource
	store          store.Store
	onLimitReached func(Resource, int64)
}

// New creates a new Limiter with the given options.
// If no store is provided, an in-memory store is used.
func New(opts ...Option) *Limiter {
	l := &Limiter{}
	for _, o := range opts {
		o(l)
	}
	if l.store == nil {
		l.store = store.NewMemoryStore()
	}
	return l
}

// Register adds a resource to be tracked by the limiter.
func (l *Limiter) Register(r Resource) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.resources = append(l.resources, r)
}

// Check tests whether a request to the given URL is allowed.
// It increments the counter and enforces the resource's strategy.
// Returns nil if the request is allowed, or an error if it should be blocked.
func (l *Limiter) Check(ctx context.Context, rawURL string) error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, r := range l.resources {
		if !matchURL(rawURL, r.Pattern) {
			continue
		}

		now := time.Now()
		w := store.Window{
			Duration:    r.Window.Duration(),
			BucketKey:   r.Window.BucketKey(now),
			BucketStart: r.Window.BucketStart(now),
		}

		current, err := l.store.Increment(ctx, r.Name, w)
		if err != nil {
			return fmt.Errorf("erl: store error: %w", err)
		}

		if current > r.Limit {
			if l.onLimitReached != nil {
				l.onLimitReached(r, current)
			}

			switch r.Strategy {
			case Block:
				return &LimitExceededError{
					Resource: r,
					Current:  current,
					resetAt:  w.BucketStart.Add(w.Duration),
				}
			case BlockWithQueue:
				return &LimitExceededError{
					Resource: r,
					Current:  current,
					resetAt:  w.BucketStart.Add(w.Duration),
				}
			case LogOnly:
				// Allow the request through.
				return nil
			}
		}

		// Matched a resource and under the limit; allow.
		return nil
	}

	// No matching resource; allow.
	return nil
}

// GetUsage returns the current counter for a resource in the active window.
func (l *Limiter) GetUsage(ctx context.Context, name string) (int64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, r := range l.resources {
		if r.Name == name {
			now := time.Now()
			w := store.Window{
				Duration:    r.Window.Duration(),
				BucketKey:   r.Window.BucketKey(now),
				BucketStart: r.Window.BucketStart(now),
			}
			return l.store.Get(ctx, r.Name, w)
		}
	}

	return 0, fmt.Errorf("erl: resource %q not found", name)
}

// ResetUsage resets the counter for a resource.
func (l *Limiter) ResetUsage(ctx context.Context, name string) error {
	return l.store.Reset(ctx, name)
}

// Transport wraps an http.RoundTripper so that all requests made through it
// are automatically checked against registered resources.
func (l *Limiter) Transport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &transport{limiter: l, base: base}
}

// Resources returns a copy of all registered resources.
func (l *Limiter) Resources() []Resource {
	l.mu.RLock()
	defer l.mu.RUnlock()

	out := make([]Resource, len(l.resources))
	copy(out, l.resources)
	return out
}

// ResourceStatus holds a point-in-time counter for a single resource.
type ResourceStatus struct {
	Resource Resource
	Current  int64
}

// Snapshot returns the current counter for every registered resource
// in its active window bucket.
func (l *Limiter) Snapshot(ctx context.Context) ([]ResourceStatus, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	out := make([]ResourceStatus, 0, len(l.resources))
	now := time.Now()

	for _, r := range l.resources {
		w := store.Window{
			Duration:    r.Window.Duration(),
			BucketKey:   r.Window.BucketKey(now),
			BucketStart: r.Window.BucketStart(now),
		}

		current, err := l.store.Get(ctx, r.Name, w)
		if err != nil {
			return nil, fmt.Errorf("erl: snapshot %s: %w", r.Name, err)
		}

		out = append(out, ResourceStatus{Resource: r, Current: current})
	}

	return out, nil
}

// Close releases resources held by the limiter's store.
func (l *Limiter) Close() error {
	return l.store.Close()
}
