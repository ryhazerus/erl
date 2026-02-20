// Package erl (External Rate Limiter) provides client-side rate limiting for
// outgoing HTTP requests. It lets you define rate limit budgets for external
// APIs and enforces them before requests leave your application.
//
// # Key Concepts
//
//   - [Resource] describes a tracked API endpoint: a URL pattern, a call limit,
//     a time [Window], and an enforcement [Strategy].
//   - [Window] sets the duration of a rate limit bucket (per-minute, per-hour,
//     per-day, or per-month).
//   - [Strategy] controls what happens when the limit is exceeded: block the
//     request, block with the option to wait, or log only.
//   - [store.Store] is the counter backend. An in-memory store is used by
//     default; a SQLite-backed store is available for persistence across
//     restarts.
//
// # Quick Start
//
//	limiter := erl.New()
//	limiter.Register(erl.Resource{
//		Name:     "stripe",
//		Pattern:  "api.stripe.com/*",
//		Limit:    100,
//		Window:   erl.PerMinute,
//		Strategy: erl.Block,
//	})
//
//	// Wrap an http.Client to enforce limits automatically.
//	client := &http.Client{
//		Transport: limiter.Transport(nil),
//	}
//
// See the [Limiter] documentation for the full API.
package erl
