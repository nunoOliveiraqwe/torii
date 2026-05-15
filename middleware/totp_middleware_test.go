package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTOTPMiddlewareRateLimitReturnsTooManyRequests(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := TOTPMiddleware(BuildContext{RuntimeContext: context.Background()}, next, Config{
		Type: "TOTP",
		Options: map[string]interface{}{
			"seed": "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
			"limiter-req": map[string]interface{}{
				"rate-per-second": 0.001,
				"burst":           1,
			},
			"rate-limit-cache-ttl":        "1h",
			"rate-limit-cleanup-interval": "1h",
		},
	})

	rec := httptest.NewRecorder()
	req := newTOTPVerifyRequest("10.0.0.1:1234")
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	rec = httptest.NewRecorder()
	req = newTOTPVerifyRequest("10.0.0.1:1234")
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusTooManyRequests, rec.Code)
	require.Equal(t, "1000", rec.Header().Get("Retry-After"))
}

func newTOTPVerifyRequest(remoteAddr string) *http.Request {
	req := httptest.NewRequest(
		http.MethodPost,
		"https://example.test/app?__torii_totp=verify",
		strings.NewReader("code=1"),
	)
	req.RemoteAddr = remoteAddr
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}
