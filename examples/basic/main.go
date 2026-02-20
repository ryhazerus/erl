package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/ryhazerus/erl"
	"github.com/ryhazerus/erl/store"
)

func main() {
	// Start a test server to simulate an external API.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK from %s", r.URL.Path)
	}))
	defer srv.Close()

	// Create a limiter with an in-memory store and a callback.
	limiter := erl.New(
		erl.WithStore(store.NewMemoryStore()),
		erl.WithOnLimitReached(func(r erl.Resource, current int64) {
			log.Printf("LIMIT REACHED: %s (%d/%d)", r.Name, current, r.Limit)
		}),
	)
	defer limiter.Close()

	// Register a resource with a low limit for demonstration.
	limiter.Register(erl.Resource{
		Name:     "test-api",
		Pattern:  "*", // matches everything for this demo
		Limit:    3,
		Window:   erl.PerMinute,
		Strategy: erl.Block,
	})

	// Wrap the default transport.
	client := &http.Client{
		Transport: limiter.Transport(nil),
	}

	// Make requests.
	for i := 1; i <= 5; i++ {
		resp, err := client.Get(srv.URL + fmt.Sprintf("/request/%d", i))
		if err != nil {
			fmt.Printf("Request %d: BLOCKED — %v\n", i, err)
			continue
		}
		resp.Body.Close()
		fmt.Printf("Request %d: %s\n", i, resp.Status)
	}

	// Show usage.
	usage, _ := limiter.GetUsage(context.Background(), "test-api")
	fmt.Printf("\nCurrent usage for test-api: %d/3\n", usage)
}
