package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ryhazerus/erl"
	"github.com/ryhazerus/erl/store"
)

func main() {
	path := "erl_demo.db"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	s, err := store.NewSQLiteStore(path)
	if err != nil {
		log.Fatal(err)
	}

	limiter := erl.New(erl.WithStore(s))
	defer limiter.Close()

	limiter.Register(erl.Resource{
		Name:     "stripe-api",
		Pattern:  "api.stripe.com/*",
		Limit:    100,
		Window:   erl.PerMinute,
		Strategy: erl.Block,
	})
	limiter.Register(erl.Resource{
		Name:     "github-api",
		Pattern:  "api.github.com/*",
		Limit:    5000,
		Window:   erl.PerHour,
		Strategy: erl.Block,
	})
	limiter.Register(erl.Resource{
		Name:     "openai-api",
		Pattern:  "api.openai.com/*",
		Limit:    60,
		Window:   erl.PerMinute,
		Strategy: erl.BlockWithQueue,
	})
	limiter.Register(erl.Resource{
		Name:     "weather-api",
		Pattern:  "api.weather.gov/*",
		Limit:    1000,
		Window:   erl.PerDay,
		Strategy: erl.LogOnly,
	})

	ctx := context.Background()

	// Simulate some traffic.
	for i := 0; i < 42; i++ {
		limiter.Check(ctx, "https://api.stripe.com/v1/charges")
	}
	for i := 0; i < 137; i++ {
		limiter.Check(ctx, "https://api.github.com/repos")
	}
	for i := 0; i < 58; i++ {
		limiter.Check(ctx, "https://api.openai.com/v1/chat/completions")
	}
	for i := 0; i < 312; i++ {
		limiter.Check(ctx, "https://api.weather.gov/points/39,-77")
	}

	fmt.Printf("Seeded %s with sample counters.\n", path)
	fmt.Println("Run: erl-dashboard --sqlite", path)
}
