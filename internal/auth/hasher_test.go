package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHasher() *Argon2Hasher {
	// Use small params for fast tests
	return NewArgon2Hasher(64*1024, 1, 1, 32)
}

func TestHasher_HashAndCompare_Success(t *testing.T) {
	h := newTestHasher()
	salt := []byte("1234567890123456")

	hashed, err := h.Hash("my-password", salt)
	require.NoError(t, err)
	assert.NotEmpty(t, hashed)

	err = h.CompareHashAndPassword(hashed, "my-password")
	assert.NoError(t, err)
}

func TestHasher_CompareHashAndPassword_WrongPassword(t *testing.T) {
	h := newTestHasher()
	salt := []byte("1234567890123456")

	hashed, err := h.Hash("correct-password", salt)
	require.NoError(t, err)

	err = h.CompareHashAndPassword(hashed, "wrong-password")
	assert.Error(t, err)
}

func TestHasher_CompareHashAndPassword_InvalidHash(t *testing.T) {
	h := newTestHasher()

	err := h.CompareHashAndPassword("not-a-valid-hash", "password")
	assert.Error(t, err)
}

func TestHasher_DifferentSalts_DifferentHashes(t *testing.T) {
	h := newTestHasher()

	hash1, _ := h.Hash("password", []byte("salt-aaaaaaaaaaaa"))
	hash2, _ := h.Hash("password", []byte("salt-bbbbbbbbbbbb"))

	assert.NotEqual(t, hash1, hash2)
}

func TestHasher_GetName(t *testing.T) {
	h := newTestHasher()
	assert.Equal(t, ARGON2id, h.GetName())
}

func TestEncoder_EncryptAndMatches(t *testing.T) {
	enc := NewDefaultEncoder()

	hashed, err := enc.Encrypt("test-password")
	require.NoError(t, err)
	assert.NotEmpty(t, hashed)

	assert.NoError(t, enc.Matches("test-password", hashed))
	assert.Error(t, enc.Matches("wrong-password", hashed))
}

func TestEncoder_EncryptProducesDifferentHashes(t *testing.T) {
	enc := NewDefaultEncoder()

	hash1, err := enc.Encrypt("same-password")
	require.NoError(t, err)

	hash2, err := enc.Encrypt("same-password")
	require.NoError(t, err)

	// Different salts should produce different hashes
	assert.NotEqual(t, hash1, hash2)
}
