package erl

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTransportAllowsRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	l := New()
	l.Register(Resource{
		Name:     "test-server",
		Pattern:  "*",
		Limit:    10,
		Window:   PerMinute,
		Strategy: Block,
	})

	client := &http.Client{
		Transport: l.Transport(nil),
	}

	resp, err := client.Get(srv.URL + "/hello")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestTransportBlocksWhenLimitExceeded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	l := New()
	l.Register(Resource{
		Name:     "limited-server",
		Pattern:  "*",
		Limit:    2,
		Window:   PerMinute,
		Strategy: Block,
	})

	client := &http.Client{
		Transport: l.Transport(nil),
	}

	// First 2 should succeed.
	for i := 0; i < 2; i++ {
		resp, err := client.Get(srv.URL + "/test")
		if err != nil {
			t.Fatalf("request %d: %v", i+1, err)
		}
		resp.Body.Close()
	}

	// 3rd should fail.
	_, err := client.Get(srv.URL + "/test")
	if err == nil {
		t.Fatal("expected error on 3rd request")
	}
	if !errors.Is(err, ErrLimitExceeded) {
		// http.Client wraps the error; unwrap it.
		var urlErr interface{ Unwrap() error }
		if errors.As(err, &urlErr) {
			inner := urlErr.Unwrap()
			if !errors.Is(inner, ErrLimitExceeded) {
				t.Fatalf("expected ErrLimitExceeded in chain, got: %v", err)
			}
		}
	}
}
