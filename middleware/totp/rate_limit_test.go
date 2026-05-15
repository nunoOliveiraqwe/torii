package totp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nunoOliveiraqwe/torii/internal/session_store"
	"github.com/nunoOliveiraqwe/torii/internal/util"
	"github.com/stretchr/testify/require"
)

const testTOTPSeed = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

func TestValidateAndStartSessionRateLimitsVerificationAttempts(t *testing.T) {
	manager, err := NewTOTPManager(Config{
		Seed: testTOTPSeed,
		RateLimit: &RateLimitConfig{
			RatePerSecond: 0.001,
			Burst:         1,
			CacheOptions:  testRateLimitCacheOptions(),
		},
	}, testSessionConfig())
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "https://example.test/app", nil)
	req.RemoteAddr = "10.0.0.1:1234"

	ok, err := manager.ValidateAndStartSession(req, "1", "scope")
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = manager.ValidateAndStartSession(req, "1", "scope")
	require.False(t, ok)
	var rateLimitErr *RateLimitError
	require.True(t, errors.As(err, &rateLimitErr))
	require.Equal(t, "1000", rateLimitErr.RetryAfter)
}

func TestValidateAndStartSessionCanDisableRateLimit(t *testing.T) {
	manager, err := NewTOTPManager(Config{
		Seed:      testTOTPSeed,
		RateLimit: &RateLimitConfig{Disabled: true},
	}, testSessionConfig())
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "https://example.test/app", nil)
	req.RemoteAddr = "10.0.0.1:1234"

	for i := 0; i < 3; i++ {
		ok, err := manager.ValidateAndStartSession(req, "1", "scope")
		require.NoError(t, err)
		require.False(t, ok)
	}
}

func testRateLimitCacheOptions() *util.CacheOptions {
	return &util.CacheOptions{
		MaxEntries:      100,
		TTL:             time.Hour,
		CleanupInterval: time.Hour,
	}
}

func testSessionConfig() session_store.Config {
	return session_store.Config{
		Lifetime:        time.Hour,
		IdleTimeout:     time.Hour,
		CleanupInterval: time.Hour,
		CookieSameSite:  "lax",
	}
}
