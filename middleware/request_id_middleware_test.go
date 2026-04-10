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

	mw := RequestIDMiddleware(context.Background(), next, Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	assert.True(t, called)
}

func TestRequestIDMiddleware_UsesExistingID(t *testing.T) {
	prefix := "my-prefix"

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		reqID := GetRequestIDFromContext(r.Context())
		assert.Equal(t, prefix+"-42", reqID)
	})

	conf := Config{Options: map[string]interface{}{"prefix": prefix}}
	mw := RequestIDMiddleware(context.Background(), next, conf)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	existingID := prefix + "-42"
	ctx := context.WithValue(req.Context(), requestIdContextKey, existingID)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	assert.True(t, called)
}

func TestRequestIDMiddleware_CompositeID_DifferentPrefix(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		reqID := GetRequestIDFromContext(r.Context())
		assert.Contains(t, reqID, "existing-request-id")
	})

	mw := RequestIDMiddleware(context.Background(), next, Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), requestIdContextKey, "existing-request-id")
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
	mw := RequestIDMiddleware(context.Background(), next, conf)

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
