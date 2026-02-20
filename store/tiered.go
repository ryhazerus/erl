package store

import "context"

// Compile-time interface check.
var _ Store = (*TieredStore)(nil)

// TieredStore wraps an in-memory store (fast path) with a persistent backend
// (durable path). Writes go to both stores (write-through); reads check memory
// first and fall back to the persistent store on a miss.
type TieredStore struct {
	memory     *MemoryStore
	persistent Store
}

// NewTieredStore creates a TieredStore backed by the given persistent store.
// An internal MemoryStore is created automatically.
func NewTieredStore(persistent Store) *TieredStore {
	return &TieredStore{
		memory:     NewMemoryStore(),
		persistent: persistent,
	}
}

// Increment writes through to both memory and the persistent backend.
// The persistent store is the source of truth for the returned count.
func (t *TieredStore) Increment(ctx context.Context, key string, w Window) (int64, error) {
	count, err := t.persistent.Increment(ctx, key, w)
	if err != nil {
		return 0, err
	}

	// Keep memory in sync. We ignore the memory return value because the
	// persistent store is authoritative.
	t.memory.Increment(ctx, key, w)

	return count, nil
}

// Get reads from memory first. On a miss (zero value), it falls back to the
// persistent store and backfills memory.
func (t *TieredStore) Get(ctx context.Context, key string, w Window) (int64, error) {
	count, err := t.memory.Get(ctx, key, w)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return count, nil
	}

	// Memory miss — read from persistent backend.
	count, err = t.persistent.Get(ctx, key, w)
	if err != nil {
		return 0, err
	}

	// Backfill memory so subsequent reads are fast. We approximate by
	// resetting and re-incrementing to the persistent count. This is safe
	// because the tiered store serialises through the MemoryStore lock.
	if count > 0 {
		t.memory.mu.Lock()
		t.memory.buckets[key] = &bucket{count: count, bucketKey: w.BucketKey}
		t.memory.mu.Unlock()
	}

	return count, nil
}

// Reset removes the counter from both stores.
func (t *TieredStore) Reset(ctx context.Context, key string) error {
	t.memory.Reset(ctx, key)
	return t.persistent.Reset(ctx, key)
}

// Close closes the persistent backend. The in-memory store needs no cleanup.
func (t *TieredStore) Close() error {
	return t.persistent.Close()
}
