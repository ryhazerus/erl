package erl_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ryhazerus/erl"
)

func ExampleNew() {
	limiter := erl.New()
	limiter.Register(erl.Resource{
		Name:     "stripe",
		Pattern:  "api.stripe.com/*",
		Limit:    100,
		Window:   erl.PerMinute,
		Strategy: erl.Block,
	})

	fmt.Println("limiter created")
	// Output: limiter created
}

func ExampleLimiter_Check() {
	limiter := erl.New()
	limiter.Register(erl.Resource{
		Name:     "stripe",
		Pattern:  "api.stripe.com/*",
		Limit:    2,
		Window:   erl.PerMinute,
		Strategy: erl.Block,
	})

	ctx := context.Background()
	fmt.Println(limiter.Check(ctx, "https://api.stripe.com/v1/charges"))
	fmt.Println(limiter.Check(ctx, "https://api.stripe.com/v1/charges"))
	fmt.Println(limiter.Check(ctx, "https://api.stripe.com/v1/charges"))
	// Output:
	// <nil>
	// <nil>
	// erl: rate limit exceeded for stripe (3/2)
}

func ExampleLimiter_Transport() {
	limiter := erl.New()
	limiter.Register(erl.Resource{
		Name:     "example",
		Pattern:  "api.example.com/*",
		Limit:    10,
		Window:   erl.PerMinute,
		Strategy: erl.Block,
	})

	client := &http.Client{
		Transport: limiter.Transport(nil),
	}

	_ = client // use client to make requests
	fmt.Println("client configured")
	// Output: client configured
}
