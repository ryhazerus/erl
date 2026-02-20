package erl

import "net/http"

// transport implements http.RoundTripper and checks rate limits before
// forwarding requests to the underlying transport.
type transport struct {
	limiter *Limiter
	base    http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.limiter.Check(req.Context(), req.URL.String()); err != nil {
		return nil, err
	}
	return t.base.RoundTrip(req)
}
