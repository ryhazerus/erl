package erl

// Resource defines a tracked external API endpoint with its rate limit configuration.
type Resource struct {
	Name     string   // unique identifier, e.g. "stripe-api"
	Pattern  string   // URL match pattern, e.g. "api.stripe.com/*"
	Limit    int64    // max calls allowed in the window
	Window   Window   // PerMinute, PerHour, PerDay, PerMonth
	Strategy Strategy // Block, BlockWithQueue, LogOnly
}
