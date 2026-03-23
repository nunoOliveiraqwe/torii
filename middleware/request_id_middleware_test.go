package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// RequestIDMiddleware
// ---------------------------------------------------------------------------

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		// The middleware should have set a request ID in the context.
		reqID := GetRequestIDFromContext(r.Context())
		assert.NotEmpty(t, reqID)
	})

	mw := RequestIDMiddleware(next, Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	assert.True(t, called)
}

func TestRequestIDMiddleware_UsesExistingID(t *testing.T) {
	existingID := "existing-request-id"

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		// Since we set a request ID in the context, the middleware should
		// pass it through unchanged.
		reqID := GetRequestIDFromContext(r.Context())
		assert.Equal(t, existingID, reqID)
	})

	mw := RequestIDMiddleware(next, Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), requestIdContextKey, existingID)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	assert.True(t, called)
}

func TestRequestIDMiddleware_WithPrefix(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		reqID := GetRequestIDFromContext(r.Context())
		assert.Contains(t, reqID, "testprefix")
	})

	conf := Config{
		Options: map[string]interface{}{
			"prefix": "testprefix",
		},
	}
	mw := RequestIDMiddleware(next, conf)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	assert.True(t, called)
}

// ---------------------------------------------------------------------------
// GetRequestIDFromContext
// ---------------------------------------------------------------------------

func TestGetRequestIDFromContext_NilContext(t *testing.T) {
	result := GetRequestIDFromContext(nil)
	assert.Empty(t, result)
}

func TestGetRequestIDFromContext_NoValue(t *testing.T) {
	ctx := context.Background()
	result := GetRequestIDFromContext(ctx)
	assert.Empty(t, result)
}

func TestGetRequestIDFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), requestIdContextKey, 12345)
	result := GetRequestIDFromContext(ctx)
	assert.Empty(t, result)
}
