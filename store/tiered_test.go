package store

import (
	"context"
	"testing"
	"time"
)

func newTestTieredStore(t *testing.T) *TieredStore {
	t.Helper()
	persistent, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	ts := NewTieredStore(persistent)
	t.Cleanup(func() { ts.Close() })
	return ts
}

func TestTieredStoreIncrement(t *testing.T) {
	s := newTestTieredStore(t)
	ctx := context.Background()
	w := Window{
		Duration:    time.Minute,
		BucketKey:   "2024-01-15T14:30",
		BucketStart: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
	}

	for i := int64(1); i <= 5; i++ {
		got, err := s.Increment(ctx, "test", w)
		if err != nil {
			t.Fatal(err)
		}
		if got != i {
			t.Errorf("increment %d: got %d, want %d", i, got, i)
		}
	}
}

func TestTieredStoreWindowRollover(t *testing.T) {
	s := newTestTieredStore(t)
	ctx := context.Background()

	w1 := Window{
		Duration:    time.Minute,
		BucketKey:   "2024-01-15T14:30",
		BucketStart: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
	}
	w2 := Window{
		Duration:    time.Minute,
		BucketKey:   "2024-01-15T14:31",
		BucketStart: time.Date(2024, 1, 15, 14, 31, 0, 0, time.UTC),
	}

	s.Increment(ctx, "key", w1)
	s.Increment(ctx, "key", w1)

	got, _ := s.Increment(ctx, "key", w2)
	if got != 1 {
		t.Errorf("after rollover: got %d, want 1", got)
	}
}

func TestTieredStoreGet(t *testing.T) {
	s := newTestTieredStore(t)
	ctx := context.Background()
	w := Window{
		Duration:    time.Minute,
		BucketKey:   "2024-01-15T14:30",
		BucketStart: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
	}

	got, _ := s.Get(ctx, "key", w)
	if got != 0 {
		t.Errorf("initial get: got %d, want 0", got)
	}

	s.Increment(ctx, "key", w)
	s.Increment(ctx, "key", w)

	got, _ = s.Get(ctx, "key", w)
	if got != 2 {
		t.Errorf("after 2 increments: got %d, want 2", got)
	}
}

func TestTieredStoreReset(t *testing.T) {
	s := newTestTieredStore(t)
	ctx := context.Background()
	w := Window{
		Duration:    time.Minute,
		BucketKey:   "2024-01-15T14:30",
		BucketStart: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
	}

	s.Increment(ctx, "key", w)
	s.Reset(ctx, "key")

	got, _ := s.Get(ctx, "key", w)
	if got != 0 {
		t.Errorf("after reset: got %d, want 0", got)
	}
}

func TestTieredStorePersistentFallback(t *testing.T) {
	persistent, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer persistent.Close()

	ctx := context.Background()
	w := Window{
		Duration:    time.Minute,
		BucketKey:   "2024-01-15T14:30",
		BucketStart: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
	}

	// Write data through a tiered store.
	ts1 := NewTieredStore(persistent)
	ts1.Increment(ctx, "key", w)
	ts1.Increment(ctx, "key", w)
	ts1.Increment(ctx, "key", w)

	// Simulate memory loss by creating a new tiered store with the same
	// persistent backend but a fresh MemoryStore.
	ts2 := NewTieredStore(persistent)

	got, err := ts2.Get(ctx, "key", w)
	if err != nil {
		t.Fatal(err)
	}
	if got != 3 {
		t.Errorf("persistent fallback: got %d, want 3", got)
	}
}
