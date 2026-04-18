package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromArgon2String_Valid(t *testing.T) {
	input := "$argon2id$v=19$t=3,m=65536,p=4$c29tZXNhbHQ$c29tZWhhc2g"
	h, err := FromArgon2String(input)
	require.NoError(t, err)

	assert.Equal(t, "argon2id", h.AlgorithmName)
	assert.Equal(t, 19, h.Version)
	assert.Equal(t, 3, h.Iterations)
	assert.Equal(t, 65536, h.Memory)
	assert.Equal(t, 4, h.Parallelism)
	assert.Equal(t, []byte("somesalt"), h.Salt)
	assert.Equal(t, []byte("somehash"), h.HashedData)
}

func TestFromArgon2String_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"no dollar prefix", "argon2id$v=19$t=3,m=65536,p=4$salt$hash"},
		{"too few parts", "$argon2id$v=19$t=3"},
		{"unsupported algorithm", "$bcrypt$v=19$t=3,m=65536,p=4$salt$hash"},
		{"missing version prefix", "$argon2id$19$t=3,m=65536,p=4$salt$hash"},
		{"invalid version", "$argon2id$v=abc$t=3,m=65536,p=4$salt$hash"},
		{"missing memory", "$argon2id$v=19$t=3,p=4$c29tZXNhbHQ$c29tZWhhc2g"},
		{"missing iterations", "$argon2id$v=19$m=65536,p=4$c29tZXNhbHQ$c29tZWhhc2g"},
		{"missing parallelism", "$argon2id$v=19$t=3,m=65536$c29tZXNhbHQ$c29tZWhhc2g"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FromArgon2String(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestToArgon2String_RoundTrip(t *testing.T) {
	input := "$argon2id$v=19$t=3,m=65536,p=4$c29tZXNhbHQ$c29tZWhhc2g"
	h, err := FromArgon2String(input)
	require.NoError(t, err)

	output := h.ToArgon2String()
	assert.Equal(t, input, output)
}

func TestArgon2Hash_Equals(t *testing.T) {
	base := &Argon2Hash{
		AlgorithmName: ARGON2id,
		Version:       19,
		HashedData:    []byte("hash"),
		Salt:          []byte("salt"),
		Iterations:    3,
		Parallelism:   4,
		Memory:        65536,
	}

	identical := &Argon2Hash{
		AlgorithmName: ARGON2id,
		Version:       19,
		HashedData:    []byte("hash"),
		Salt:          []byte("salt"),
		Iterations:    3,
		Parallelism:   4,
		Memory:        65536,
	}

	assert.True(t, base.Equals(identical))
	assert.False(t, base.Equals(nil))

	different := *identical
	different.HashedData = []byte("diff")
	assert.False(t, base.Equals(&different))

	differentSalt := *identical
	differentSalt.Salt = []byte("xxxx")
	assert.False(t, base.Equals(&differentSalt))

	differentVersion := *identical
	differentVersion.Version = 18
	assert.False(t, base.Equals(&differentVersion))

	differentAlgo := *identical
	differentAlgo.AlgorithmName = ARGON2i
	assert.False(t, base.Equals(&differentAlgo))

	differentIter := *identical
	differentIter.Iterations = 99
	assert.False(t, base.Equals(&differentIter))

	differentPar := *identical
	differentPar.Parallelism = 99
	assert.False(t, base.Equals(&differentPar))

	differentMem := *identical
	differentMem.Memory = 99
	assert.False(t, base.Equals(&differentMem))
}
