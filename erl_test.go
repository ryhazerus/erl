package erl

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestLimiterBlockStrategy(t *testing.T) {
	l := New()
	l.Register(Resource{
		Name:     "test-api",
		Pattern:  "api.test.com/*",
		Limit:    3,
		Window:   PerMinute,
		Strategy: Block,
	})

	ctx := context.Background()
	url := "https://api.test.com/v1/foo"

	// First 3 requests should succeed.
	for i := 0; i < 3; i++ {
		if err := l.Check(ctx, url); err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}
	}

	// 4th request should be blocked.
	err := l.Check(ctx, url)
	if err == nil {
		t.Fatal("expected error on 4th request, got nil")
	}
	if !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("expected ErrLimitExceeded, got: %v", err)
	}

	var limErr *LimitExceededError
	if !errors.As(err, &limErr) {
		t.Fatalf("expected *LimitExceededError, got %T", err)
	}
	if limErr.Resource.Name != "test-api" {
		t.Errorf("resource name = %q, want %q", limErr.Resource.Name, "test-api")
	}
}

func TestLimiterLogOnlyStrategy(t *testing.T) {
	var callbackCalled bool
	l := New(
		WithOnLimitReached(func(r Resource, current int64) {
			callbackCalled = true
		}),
	)
	l.Register(Resource{
		Name:     "log-api",
		Pattern:  "api.logged.com/*",
		Limit:    1,
		Window:   PerMinute,
		Strategy: LogOnly,
	})

	ctx := context.Background()
	url := "https://api.logged.com/endpoint"

	// First request.
	if err := l.Check(ctx, url); err != nil {
		t.Fatalf("request 1: %v", err)
	}

	// Second request exceeds limit but should still pass.
	if err := l.Check(ctx, url); err != nil {
		t.Fatalf("request 2 (LogOnly) should not error, got: %v", err)
	}
	if !callbackCalled {
		t.Error("expected OnLimitReached callback to be called")
	}
}

func TestLimiterUnmatchedURLPassesThrough(t *testing.T) {
	l := New()
	l.Register(Resource{
		Name:    "only-stripe",
		Pattern: "api.stripe.com/*",
		Limit:   1,
		Window:  PerMinute,
	})

	ctx := context.Background()
	// URL that doesn't match any resource should always pass.
	for i := 0; i < 10; i++ {
		if err := l.Check(ctx, "https://api.github.com/repos"); err != nil {
			t.Fatalf("unmatched request %d: %v", i, err)
		}
	}
}

func TestLimiterConcurrent(t *testing.T) {
	l := New()
	l.Register(Resource{
		Name:     "concurrent-api",
		Pattern:  "api.concurrent.com/*",
		Limit:    100,
		Window:   PerMinute,
		Strategy: Block,
	})

	ctx := context.Background()
	url := "https://api.concurrent.com/v1/test"

	var wg sync.WaitGroup
	errs := make(chan error, 200)

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- l.Check(ctx, url)
		}()
	}

	wg.Wait()
	close(errs)

	var allowed, blocked int
	for err := range errs {
		if err == nil {
			allowed++
		} else if errors.Is(err, ErrLimitExceeded) {
			blocked++
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if allowed != 100 {
		t.Errorf("allowed = %d, want 100", allowed)
	}
	if blocked != 100 {
		t.Errorf("blocked = %d, want 100", blocked)
	}
}

func TestLimiterGetUsage(t *testing.T) {
	l := New()
	l.Register(Resource{
		Name:    "usage-api",
		Pattern: "api.usage.com/*",
		Limit:   100,
		Window:  PerMinute,
	})

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		l.Check(ctx, "https://api.usage.com/test")
	}

	count, err := l.GetUsage(ctx, "usage-api")
	if err != nil {
		t.Fatal(err)
	}
	if count != 5 {
		t.Errorf("usage = %d, want 5", count)
	}
}

func TestLimiterResetUsage(t *testing.T) {
	l := New()
	l.Register(Resource{
		Name:    "reset-api",
		Pattern: "api.reset.com/*",
		Limit:   3,
		Window:  PerMinute,
		Strategy: Block,
	})

	ctx := context.Background()
	url := "https://api.reset.com/test"

	for i := 0; i < 3; i++ {
		l.Check(ctx, url)
	}

	// Should be blocked now.
	if err := l.Check(ctx, url); !errors.Is(err, ErrLimitExceeded) {
		t.Fatal("expected limit exceeded before reset")
	}

	// Reset and try again.
	if err := l.ResetUsage(ctx, "reset-api"); err != nil {
		t.Fatal(err)
	}

	if err := l.Check(ctx, url); err != nil {
		t.Fatalf("after reset, expected nil error, got: %v", err)
	}
}
