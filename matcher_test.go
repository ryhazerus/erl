package erl

import "testing"

func TestMatchURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		pattern string
		want    bool
	}{
		{
			name:    "wildcard host match",
			url:     "https://api.stripe.com/v1/charges",
			pattern: "api.stripe.com/*",
			want:    true,
		},
		{
			name:    "wildcard host root",
			url:     "https://api.stripe.com/",
			pattern: "api.stripe.com/*",
			want:    true,
		},
		{
			name:    "wildcard sub-path",
			url:     "https://api.openai.com/v1/chat/completions",
			pattern: "api.openai.com/v1/chat/*",
			want:    true,
		},
		{
			name:    "wildcard does not match different path",
			url:     "https://api.openai.com/v1/embeddings",
			pattern: "api.openai.com/v1/chat/*",
			want:    false,
		},
		{
			name:    "exact match",
			url:     "https://api.example.com/v1/specific",
			pattern: "api.example.com/v1/specific",
			want:    true,
		},
		{
			name:    "exact no match",
			url:     "https://api.example.com/v1/other",
			pattern: "api.example.com/v1/specific",
			want:    false,
		},
		{
			name:    "different host",
			url:     "https://api.github.com/repos",
			pattern: "api.stripe.com/*",
			want:    false,
		},
		{
			name:    "with query params",
			url:     "https://api.stripe.com/v1/charges?limit=10",
			pattern: "api.stripe.com/*",
			want:    true,
		},
		{
			name:    "host only match",
			url:     "https://api.stripe.com",
			pattern: "api.stripe.com/*",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchURL(tt.url, tt.pattern)
			if got != tt.want {
				t.Errorf("matchURL(%q, %q) = %v, want %v", tt.url, tt.pattern, got, tt.want)
			}
		})
	}
}
