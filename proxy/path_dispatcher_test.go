package proxy

import "testing"

func TestNormalizePattern(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Trailing wildcard → catch-all
		{"/api/v1/users/*", "/api/v1/users/{path...}"},
		// Mid-path wildcard → single-segment named wildcard
		{"/users/*/start", "/users/{_seg1}/start"},
		// Mixed: mid-path + trailing
		{"/users/*/jobs/*", "/users/{_seg1}/jobs/{path...}"},
		// Multiple mid-path wildcards
		{"/a/*/b/*/c", "/a/{_seg1}/b/{_seg2}/c"},
		// Prefix match (trailing slash, no wildcard)
		{"/api/v1/users/", "/api/v1/users/"},
		// Exact match
		{"/health", "/health"},
		// Concrete path (no wildcards)
		{"/users/whatever/stop", "/users/whatever/stop"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizePattern(tt.input)
			if got != tt.expected {
				t.Errorf("normalizePattern(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
