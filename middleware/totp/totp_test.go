package totp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateCodeRFC6238Vectors(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
		algorithm Algorithm
		secret    string
		want      string
	}{
		//from https://datatracker.ietf.org/doc/html/rfc6238#appendix-B
		//https://www.rfc-editor.org/errata/eid2866 - errata here so no one has to wonder why the RFC gives one secret but the test vectors use a different one
		{name: "59 SHA1", timestamp: 59, algorithm: AlgorithmSHA1, secret: "12345678901234567890", want: "94287082"},
		{name: "59 SHA256", timestamp: 59, algorithm: AlgorithmSHA256, secret: "12345678901234567890123456789012", want: "46119246"},
		{name: "59 SHA512", timestamp: 59, algorithm: AlgorithmSHA512, secret: "1234567890123456789012345678901234567890123456789012345678901234", want: "90693936"},
		{name: "1111111109 SHA1", timestamp: 1111111109, algorithm: AlgorithmSHA1, secret: "12345678901234567890", want: "07081804"},
		{name: "1111111109 SHA256", timestamp: 1111111109, algorithm: AlgorithmSHA256, secret: "12345678901234567890123456789012", want: "68084774"},
		{name: "1111111109 SHA512", timestamp: 1111111109, algorithm: AlgorithmSHA512, secret: "1234567890123456789012345678901234567890123456789012345678901234", want: "25091201"},
		{name: "1111111111 SHA1", timestamp: 1111111111, algorithm: AlgorithmSHA1, secret: "12345678901234567890", want: "14050471"},
		{name: "1111111111 SHA256", timestamp: 1111111111, algorithm: AlgorithmSHA256, secret: "12345678901234567890123456789012", want: "67062674"},
		{name: "1111111111 SHA512", timestamp: 1111111111, algorithm: AlgorithmSHA512, secret: "1234567890123456789012345678901234567890123456789012345678901234", want: "99943326"},
		{name: "1234567890 SHA1", timestamp: 1234567890, algorithm: AlgorithmSHA1, secret: "12345678901234567890", want: "89005924"},
		{name: "1234567890 SHA256", timestamp: 1234567890, algorithm: AlgorithmSHA256, secret: "12345678901234567890123456789012", want: "91819424"},
		{name: "1234567890 SHA512", timestamp: 1234567890, algorithm: AlgorithmSHA512, secret: "1234567890123456789012345678901234567890123456789012345678901234", want: "93441116"},
		{name: "2000000000 SHA1", timestamp: 2000000000, algorithm: AlgorithmSHA1, secret: "12345678901234567890", want: "69279037"},
		{name: "2000000000 SHA256", timestamp: 2000000000, algorithm: AlgorithmSHA256, secret: "12345678901234567890123456789012", want: "90698825"},
		{name: "2000000000 SHA512", timestamp: 2000000000, algorithm: AlgorithmSHA512, secret: "1234567890123456789012345678901234567890123456789012345678901234", want: "38618901"},
		{name: "20000000000 SHA1", timestamp: 20000000000, algorithm: AlgorithmSHA1, secret: "12345678901234567890", want: "65353130"},
		{name: "20000000000 SHA256", timestamp: 20000000000, algorithm: AlgorithmSHA256, secret: "12345678901234567890123456789012", want: "77737706"},
		{name: "20000000000 SHA512", timestamp: 20000000000, algorithm: AlgorithmSHA512, secret: "1234567890123456789012345678901234567890123456789012345678901234", want: "47863826"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := generateCode([]byte(tt.secret), tt.timestamp/30, 8, tt.algorithm)

			require.NoError(t, err)
			require.Equal(t, tt.want, code)
		})
	}
}
