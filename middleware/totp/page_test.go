package totp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerifyActionPreservesRequestPathAndAddsReservedMarker(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.test/app?x=1", nil)

	action := verifyAction(req.URL)

	require.Equal(t, "/app?__torii_totp=verify&x=1", action)
}

func TestCleanVerificationURLRemovesReservedMarker(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "https://example.test/app?__torii_totp=verify&x=1", nil)

	clean := cleanVerificationURL(req.URL)

	require.Equal(t, "/app?x=1", clean)
}

func TestRenderChallengeIncludesFixedKeyboardScript(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.test/app", nil)
	rec := httptest.NewRecorder()

	RenderChallenge(rec, req, 6, "")

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Body.String(), `action="/app?__torii_totp=verify"`)
	require.Contains(t, rec.Body.String(), `id="keypad"`)
	require.Contains(t, rec.Body.String(), `Torii Proxy`)
	require.Contains(t, rec.Body.String(), `['1', '2', '3', '4', '5', '6', '7', '8', '9'].forEach`)
	require.Contains(t, rec.Body.String(), `button('0', function () { addDigit('0'); });`)
	require.NotContains(t, rec.Body.String(), `shuffle(`)
}
