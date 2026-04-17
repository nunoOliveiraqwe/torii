package proxy

import (
	"net/http"
	"testing"
)

var dummyHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

func newTestTrie() *HostTrie {
	h := NewHostTrie()
	hosts := []string{"jelly.mydomain.com", "admin.mydomain.com", "login.mydomain.com", "test.pt", "sub.test.pt", "*.test.pt"}
	for _, host := range hosts {
		h.InsertHost(host, dummyHandler)
	}
	return h
}

func TestGetAllHosts(t *testing.T) {
	h := newTestTrie()
	allHosts := h.GetAllHosts()
	expectedHosts := []string{"jelly.mydomain.com", "admin.mydomain.com", "login.mydomain.com", "test.pt", "sub.test.pt", "*.test.pt"}
	if len(allHosts) != len(expectedHosts) {
		t.Fatalf("Expected %d hosts, got %d", len(expectedHosts), len(allHosts))
	}
	hostMap := make(map[string]bool)
	for _, host := range allHosts {
		hostMap[host] = true
	}
	for _, expected := range expectedHosts {
		if !hostMap[expected] {
			t.Errorf("Expected host %s not found in result", expected)
		}
	}
}

func TestMatches(t *testing.T) {
	h := newTestTrie()

	tests := []struct {
		host    string
		matched bool
	}{
		{"jelly.mydomain.com", true},
		{"admin.mydomain.com", true},
		{"login.mydomain.com", true},
		{"test.pt", true},
		{"sub.test.pt", true},
		{"other.test.pt", true},     // wildcard *.test.pt
		{"unknown.com", false},      // no match
		{"mydomain.com", false},     // no match — no bare domain registered
		{"deep.sub.test.pt", false}, // wildcard is single-level only
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			result := h.Contains(tt.host)
			if tt.matched && result == nil {
				t.Errorf("Expected %s to match, but got nil", tt.host)
			}
			if !tt.matched && result != nil {
				t.Errorf("Expected %s to NOT match, but got a handler", tt.host)
			}
		})
	}
}

func TestWildcardDoesNotMatchExact(t *testing.T) {
	h := NewHostTrie()
	h.InsertHost("*.example.com", dummyHandler)

	if h.Contains("example.com") != nil {
		t.Error("Wildcard *.example.com should not match bare example.com")
	}

	if h.Contains("foo.example.com") == nil {
		t.Error("Wildcard *.example.com should match foo.example.com")
	}

	if h.Contains("a.b.example.com") != nil {
		t.Error("Wildcard *.example.com should not match a.b.example.com")
	}
}

func TestHostWithPort(t *testing.T) {
	h := NewHostTrie()
	h.InsertHost("app.example.com", dummyHandler)

	if h.Contains("app.example.com:8080") == nil {
		t.Error("Should match host with port stripped")
	}
}

func TestCaseInsensitive(t *testing.T) {
	h := NewHostTrie()
	h.InsertHost("App.Example.COM", dummyHandler)

	if h.Contains("app.example.com") == nil {
		t.Error("Lookup should be case-insensitive")
	}
}

func TestHasAnyEntry(t *testing.T) {
	h := NewHostTrie()
	if h.HasAnyEntry() {
		t.Error("Empty trie should have no entries")
	}
	h.InsertHost("example.com", dummyHandler)
	if !h.HasAnyEntry() {
		t.Error("Trie with an insert should have entries")
	}
}
