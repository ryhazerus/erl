package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/ryhazerus/erl/store"
)

func newTestRedisStore(t *testing.T) *RedisStore {
	t.Helper()
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })
	return NewRedisStore(client)
}

func TestRedisStoreIncrement(t *testing.T) {
	s := newTestRedisStore(t)
	ctx := context.Background()
	w := store.Window{
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

func TestRedisStoreWindowRollover(t *testing.T) {
	s := newTestRedisStore(t)
	ctx := context.Background()

	w1 := store.Window{
		Duration:    time.Minute,
		BucketKey:   "2024-01-15T14:30",
		BucketStart: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
	}
	w2 := store.Window{
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

func TestRedisStoreGet(t *testing.T) {
	s := newTestRedisStore(t)
	ctx := context.Background()
	w := store.Window{
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

func TestRedisStoreReset(t *testing.T) {
	s := newTestRedisStore(t)
	ctx := context.Background()
	w := store.Window{
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
