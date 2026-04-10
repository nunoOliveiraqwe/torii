package proxy

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/nunoOliveiraqwe/torii/config"
	"github.com/nunoOliveiraqwe/torii/internal/proxyutil"
	"go.uber.org/zap"
)

type PathDispatcher struct {
	mux *http.ServeMux
}

func (d *PathDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.mux.ServeHTTP(w, r)
}

func buildPathDispatcher(ctx context.Context, defaultHandler http.HandlerFunc, pathRules []config.PathRule) (http.Handler, []string, []string, error) {
	mux := http.NewServeMux()

	mwNames := make([]string, 0)
	backends := make([]string, 0)
	for _, rule := range pathRules {
		pathBaseHandler := defaultHandler
		if rule.Backend != "" {
			zap.S().Infof("Building backend handler for path rule %q with backend %q", rule.Pattern, rule.Backend)
			opts := proxyutil.ProxyOptions{
				DropQuery: rule.DropQuery != nil && *rule.DropQuery,
			}
			proxy, err := buildHttpRevProxy(rule.Backend, opts)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to build reverse proxy for path rule %q: %w", rule.Pattern, err)
			}
			pathBaseHandler = buildDefaultHttpHandler(proxy)

			// Strip the path-rule prefix before forwarding to the backend so
			// that the backend sees the request at its own root.  For example,
			// a rule with pattern "/jellyfino" proxying to 192.168.1.27:2884
			// will forward "/" instead of "/jellyfino".
			if prefix := pathRulePrefix(rule.Pattern); prefix != "" {
				zap.S().Infof("Stripping prefix %q for path rule %q", prefix, rule.Pattern)
				pathBaseHandler = http.StripPrefix(prefix, pathBaseHandler).ServeHTTP
			}

			backends = append(backends, rule.Backend)
		}

		pattern := normalizePattern(rule.Pattern)
		ctx2 := context.WithValue(ctx, "path", rule.Pattern)
		handler, err := buildMiddlewareChain(ctx2, pathBaseHandler, rule.Middlewares)
		if err != nil {
			return nil, nil, nil, err
		}

		mux.HandleFunc(pattern, handler)
		mwNames = append(mwNames, middlewareNames(rule.Middlewares)...)
		zap.S().Infof("Registered path rule %q with %d middlewares", pattern, len(rule.Middlewares))
	}

	// Default catch-all: the route-level middleware chain wrapping the backend.
	mux.HandleFunc("/", defaultHandler)

	return &PathDispatcher{mux: mux}, mwNames, backends, nil
}

// pathRulePrefix extracts the static prefix from a path-rule pattern so it
// can be stripped before forwarding to the backend.
//
// Examples:
//
//	/jellyfino    → /jellyfino
//	/jellyfino/   → /jellyfino
//	/jellyfino/*  → /jellyfino
//	/api/v1/      → /api/v1
//	/             → ""          (root – nothing to strip)
func pathRulePrefix(pattern string) string {
	p := strings.TrimSuffix(pattern, "*")
	p = strings.TrimRight(p, "/")
	if p == "" {
		return ""
	}
	return p
}

// normalizePattern converts user-friendly glob patterns into Go ServeMux
// patterns. A trailing /* becomes a catch-all wildcard, and any mid-path *
// becomes a single-segment wildcard.
//
// Examples:
//
//	/api/v1/users/*       → /api/v1/users/{path...}   (catch-all)
//	/users/*/start        → /users/{_seg1}/start       (single-segment wildcard)
//	/users/*/jobs/*       → /users/{_seg1}/jobs/{path...}
//	/api/v1/users/        → /api/v1/users/             (prefix match, unchanged)
//	/health               → /health                    (exact match, unchanged)
//	/users/whatever/stop  → /users/whatever/stop        (concrete, unchanged)
//
// Precedence in Go's ServeMux: concrete segments always beat wildcards, so
// /users/whatever/stop wins over /users/{_seg1}/stop for that specific path.
func normalizePattern(pattern string) string {
	segments := strings.Split(pattern, "/")
	wildIdx := 0
	for i, seg := range segments {
		if seg != "*" {
			continue
		}
		if i == len(segments)-1 {
			// Trailing /* → catch-all
			segments[i] = "{path...}"
		} else {
			// Mid-path * → single-segment named wildcard
			wildIdx++
			segments[i] = fmt.Sprintf("{_seg%d}", wildIdx)
		}
	}
	return strings.Join(segments, "/")
}
