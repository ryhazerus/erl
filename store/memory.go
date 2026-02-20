package store

import (
	"context"
	"sync"
)

type bucket struct {
	count     int64
	bucketKey string
}

// Compile-time interface check.
var _ Store = (*MemoryStore)(nil)

// MemoryStore is an in-memory Store implementation.
// It is safe for concurrent use. Counters are lost on process restart.
type MemoryStore struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

// NewMemoryStore creates a new in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		buckets: make(map[string]*bucket),
	}
}

// Increment atomically adds one to the counter for key in the current window bucket.
func (m *MemoryStore) Increment(_ context.Context, key string, w Window) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, ok := m.buckets[key]
	if !ok || b.bucketKey != w.BucketKey {
		b = &bucket{bucketKey: w.BucketKey}
		m.buckets[key] = b
	}

	b.count++
	return b.count, nil
}

// Get returns the current counter value for key in the active window bucket.
func (m *MemoryStore) Get(_ context.Context, key string, w Window) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, ok := m.buckets[key]
	if !ok || b.bucketKey != w.BucketKey {
		return 0, nil
	}
	return b.count, nil
}

// Reset removes the counter for the given key.
func (m *MemoryStore) Reset(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.buckets, key)
	return nil
}

// Close is a no-op for the in-memory store.
func (m *MemoryStore) Close() error {
	return nil
}
