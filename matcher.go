package erl

import (
	"net/url"
	"strings"
)

// matchURL checks whether a request URL matches a resource's glob-style pattern.
// Matching is performed against host + path of the URL.
//
// Supported patterns:
//   - "api.stripe.com/*" matches any path on that host
//   - "api.openai.com/v1/chat/*" matches only chat endpoints
//   - "api.example.com/v1/specific" exact match
func matchURL(rawURL string, pattern string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	hostPath := parsed.Host + parsed.Path
	// Strip trailing slashes for consistency.
	hostPath = strings.TrimRight(hostPath, "/")
	pattern = strings.TrimRight(pattern, "/")

	return globMatch(pattern, hostPath)
}

// globMatch performs simple glob matching where "*" matches any sequence of
// characters within a single path segment and "**" or a trailing "/*" matches
// everything remaining.
func globMatch(pattern, value string) bool {
	// Fast path: exact match.
	if pattern == value {
		return true
	}

	// Trailing /* means match everything under that prefix.
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		if value == prefix || strings.HasPrefix(value, prefix+"/") {
			return true
		}
	}

	return wildcardMatch(pattern, value)
}

// wildcardMatch handles * as matching any non-empty sequence of characters.
func wildcardMatch(pattern, str string) bool {
	if pattern == "*" {
		return true
	}

	for len(pattern) > 0 {
		if pattern[0] == '*' {
			// Skip the star.
			pattern = pattern[1:]
			if len(pattern) == 0 {
				return true
			}
			// Try matching the rest of the pattern at every position.
			for i := 0; i <= len(str); i++ {
				if wildcardMatch(pattern, str[i:]) {
					return true
				}
			}
			return false
		}

		if len(str) == 0 || pattern[0] != str[0] {
			return false
		}

		pattern = pattern[1:]
		str = str[1:]
	}

	return len(str) == 0
}
