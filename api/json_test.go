package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// DecodeJSONBody
// ---------------------------------------------------------------------------

func TestDecodeJSONBody_ValidJSON(t *testing.T) {
	body := LoginRequest{Username: "admin", Password: "secret"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))

	result, err := DecodeJSONBody[LoginRequest](req)

	require.NoError(t, err)
	assert.Equal(t, "admin", result.Username)
	assert.Equal(t, "secret", result.Password)
}

func TestDecodeJSONBody_NilBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Body = nil

	_, err := DecodeJSONBody[LoginRequest](req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestDecodeJSONBody_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))

	_, err := DecodeJSONBody[LoginRequest](req)

	assert.Error(t, err)
}

func TestDecodeJSONBody_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{not valid"))

	_, err := DecodeJSONBody[LoginRequest](req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode JSON")
}

func TestDecodeJSONBody_ExtraFields(t *testing.T) {
	body := `{"username":"admin","password":"secret","extra":"field"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	result, err := DecodeJSONBody[LoginRequest](req)

	require.NoError(t, err)
	assert.Equal(t, "admin", result.Username)
}

// ---------------------------------------------------------------------------
// ParseJSONWithLimit
// ---------------------------------------------------------------------------

func TestParseJSONWithLimit_WithinLimit(t *testing.T) {
	body := `{"username":"test"}`
	result, err := ParseJSONWithLimit[LoginRequest](strings.NewReader(body), 1024)

	require.NoError(t, err)
	assert.Equal(t, "test", result.Username)
}

func TestParseJSONWithLimit_ExceedsLimit(t *testing.T) {
	// Create a body that is larger than the limit
	longVal := strings.Repeat("a", 100)
	body := `{"username":"` + longVal + `"}`
	_, err := ParseJSONWithLimit[LoginRequest](strings.NewReader(body), 10)

	assert.Error(t, err)
}

func TestParseJSONWithLimit_EmptyReader(t *testing.T) {
	_, err := ParseJSONWithLimit[LoginRequest](strings.NewReader(""), 1024)

	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// WriteResponseAsJSON
// ---------------------------------------------------------------------------

func TestWriteResponseAsJSON_Success(t *testing.T) {
	data := FtsStatusResponse{IsFtsCompleted: true}
	rec := httptest.NewRecorder()

	WriteResponseAsJSON(data, rec)
	resp := rec.Result()

	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var decoded FtsStatusResponse
	b, _ := io.ReadAll(resp.Body)
	require.NoError(t, json.Unmarshal(b, &decoded))
	assert.True(t, decoded.IsFtsCompleted)
}

func TestWriteResponseAsJSON_NilData(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteResponseAsJSON(nil, rec)

	resp := rec.Result()
	b, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "null", string(b))
}

// ---------------------------------------------------------------------------
// EncodeToJson
// ---------------------------------------------------------------------------

func TestEncodeToJson_Success(t *testing.T) {
	data := LoginRequest{Username: "admin", Password: "pw"}
	b, err := EncodeToJson(data)

	require.NoError(t, err)

	var decoded LoginRequest
	require.NoError(t, json.Unmarshal(b, &decoded))
	assert.Equal(t, "admin", decoded.Username)
}

func TestEncodeToJson_UnsupportedType(t *testing.T) {
	// Channels cannot be marshalled.
	ch := make(chan int)
	_, err := EncodeToJson(ch)
	assert.Error(t, err)
}
