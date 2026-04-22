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
		// Glued trailing star
		{"/api*", "/api/{path...}"},
		{"/jellyfin*", "/jellyfin/{path...}"},
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

func TestEnsureSubtree(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Bare path → catch-all appended
		{"/api", "/api/{path...}"},
		{"/jellyfin", "/jellyfin/{path...}"},
		// Already a trailing slash (prefix match) → unchanged
		{"/api/", "/api/"},
		// Already a catch-all → unchanged
		{"/api/{path...}", "/api/{path...}"},
		{"/api/v1/users/{path...}", "/api/v1/users/{path...}"},
		// Root → unchanged (trailing slash)
		{"/", "/"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ensureSubtree(tt.input)
			if got != tt.expected {
				t.Errorf("ensureSubtree(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestUserPatternVariants verifies that the combination of normalizePattern +
// ensureSubtree produces the expected catch-all for every way a user might
// write a path rule with a backend.
func TestUserPatternVariants(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// /api  → normalize keeps it, ensureSubtree adds catch-all
		{"/api", "/api/{path...}"},
		// /api/ → normalize keeps it, ensureSubtree leaves it (prefix match)
		{"/api/", "/api/"},
		// /api/* → normalize converts to catch-all, ensureSubtree leaves it
		{"/api/*", "/api/{path...}"},
		// /api* → normalize converts to catch-all, ensureSubtree leaves it
		{"/api*", "/api/{path...}"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ensureSubtree(normalizePattern(tt.input))
			if got != tt.expected {
				t.Errorf("ensureSubtree(normalizePattern(%q)) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPathRulePrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/jellyfino", "/jellyfino"},
		{"/jellyfino/", "/jellyfino"},
		{"/jellyfino/*", "/jellyfino"},
		{"/api/v1/", "/api/v1"},
		{"/", ""},
		{"/*", ""},
		{"/a/b/c", "/a/b/c"},
		{"/a/b/c/", "/a/b/c"},
		{"/a/b/c/*", "/a/b/c"},
		// Mid-path wildcards — cannot be used with StripPrefix
		{"/users/*/profile", ""},
		{"/users/*/profile/*", ""},
		{"/a/*/b/*/c", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := pathRulePrefix(tt.input)
			if got != tt.expected {
				t.Errorf("pathRulePrefix(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
