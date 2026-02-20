# erl — External Rate Limiter for Go

A Go library that tracks and limits outgoing HTTP requests to external APIs. Prevent unexpected overage costs by enforcing configurable rate limits per resource.

## Install

```bash
go get github.com/ryhazerus/erl
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ryhazerus/erl"
	"github.com/ryhazerus/erl/store"
)

func main() {
	// Create a limiter
	limiter := erl.New(
		erl.WithStore(store.NewMemoryStore()),
		erl.WithOnLimitReached(func(r erl.Resource, current int64) {
			log.Printf("limit hit: %s (%d/%d)", r.Name, current, r.Limit)
		}),
	)
	defer limiter.Close()

	// Register resources to track
	limiter.Register(erl.Resource{
		Name:     "stripe",
		Pattern:  "api.stripe.com/*",
		Limit:    10000,
		Window:   erl.PerMonth,
		Strategy: erl.Block,
	})

	limiter.Register(erl.Resource{
		Name:     "openai-chat",
		Pattern:  "api.openai.com/v1/chat/*",
		Limit:    500,
		Window:   erl.PerDay,
		Strategy: erl.LogOnly,
	})

	// Wrap your HTTP client
	client := &http.Client{
		Transport: limiter.Transport(http.DefaultTransport),
	}

	// All requests through this client are now tracked
	resp, err := client.Get("https://api.stripe.com/v1/charges")
	if err != nil {
		fmt.Println("blocked:", err)
		return
	}
	resp.Body.Close()
	fmt.Println(resp.Status)
}
```

## Strategies

| Strategy | Behavior |
|---|---|
| `erl.Block` | Returns `erl.ErrLimitExceeded` immediately |
| `erl.BlockWithQueue` | Blocks, but the error exposes a `Wait(ctx)` method to wait for the window to reset |
| `erl.LogOnly` | Lets the request through, fires the `OnLimitReached` callback |

### BlockWithQueue example

```go
err := limiter.Check(ctx, "https://api.stripe.com/v1/charges")
if err != nil {
	var limErr *erl.LimitExceededError
	if errors.As(err, &limErr) {
		// Wait until the rate limit window resets
		limErr.Wait(ctx)
	}
}
```

## Windows

`erl.PerMinute` · `erl.PerHour` · `erl.PerDay` · `erl.PerMonth`

## Pattern Matching

Patterns match against the request URL's `host + path`:

```
api.stripe.com/*              — any path on that host
api.openai.com/v1/chat/*      — only chat endpoints
api.example.com/v1/specific   — exact match
```

## Storage Backends

### In-memory (default)

```go
limiter := erl.New(erl.WithStore(store.NewMemoryStore()))
```

Fast, no dependencies. Resets on restart.

### SQLite

```go
s, err := store.NewSQLiteStore("erl.db")
if err != nil {
	log.Fatal(err)
}
limiter := erl.New(erl.WithStore(s))
```

Persists across restarts. Uses [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGo).

## Check Usage

```go
count, err := limiter.GetUsage(ctx, "stripe")
fmt.Printf("stripe: %d/10000\n", count)
```

## Reset Usage

```go
limiter.ResetUsage(ctx, "stripe")
```
