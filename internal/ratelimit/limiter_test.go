package ratelimit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nunoOliveiraqwe/torii/internal/util"
	"github.com/stretchr/testify/require"
)

func TestSeparateLimitersDoNotShareClientBuckets(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	first := newTestLimiter(t, ctx)
	second := newTestLimiter(t, ctx)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"

	require.True(t, first.Allow(req).Allowed)
	require.False(t, first.Allow(req).Allowed)
	require.True(t, second.Allow(req).Allowed)
}

func newTestLimiter(t *testing.T, ctx context.Context) *Limiter {
	t.Helper()
	limiter, err := New(Config{
		Mode:          ModePerClient,
		RatePerSecond: 0.001,
		Burst:         1,
		CacheOptions: &util.CacheOptions{
			Ctx:             ctx,
			MaxEntries:      100,
			TTL:             time.Hour,
			CleanupInterval: time.Hour,
		},
	})
	require.NoError(t, err)
	return limiter
}
