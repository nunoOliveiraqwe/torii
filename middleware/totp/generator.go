package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"strconv"
	"strings"
)

const (
	AlgorithmSHA1   Algorithm = "SHA1"
	AlgorithmSHA256 Algorithm = "SHA256"
	AlgorithmSHA512 Algorithm = "SHA512"
)

type Algorithm string

func generateTOTP(secret string, timestamp int64) (string, error) {
	decoded, err := decodeSecret(secret)
	if err != nil {
		return "", err
	}
	return generateCode(decoded, timestamp/int64(defaultPeriod.Seconds()), defaultDigits, AlgorithmSHA1)
}

func validateTOTP(secret []byte, code string, timestamp int64) (bool, error) {
	step := timestamp / int64(defaultPeriod.Seconds())
	normalizedCode := strings.TrimSpace(code)
	for offset := -defaultCodeWindow; offset <= defaultCodeWindow; offset++ {
		candidate, err := generateCode(secret, step+int64(offset), defaultDigits, AlgorithmSHA1)
		if err != nil {
			return false, err
		}
		if subtle.ConstantTimeCompare([]byte(candidate), []byte(normalizedCode)) == 1 {
			return true, nil
		}
	}
	return false, nil
}

func decodeSecret(secret string) ([]byte, error) {
	normalized := strings.ToUpper(secret)
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.TrimRight(normalized, "=")
	if normalized == "" {
		return nil, fmt.Errorf("secret cannot be empty")
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(normalized)
}

func generateCode(secret []byte, counter int64, digits int, algorithm Algorithm) (string, error) {
	if counter < 0 {
		return "", fmt.Errorf("counter cannot be negative")
	}
	if digits <= 0 || digits > 9 {
		return "", fmt.Errorf("digits must be between 1 and 9")
	}

	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], uint64(counter))

	algorithm, err := normalizeAlgorithm(algorithm)
	if err != nil {
		return "", err
	}
	mac, err := newHMAC(algorithm, secret)
	if err != nil {
		return "", err
	}
	if _, err := mac.Write(msg[:]); err != nil {
		return "", err
	}
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	binCode := (uint32(sum[offset])&0x7f)<<24 |
		(uint32(sum[offset+1])&0xff)<<16 |
		(uint32(sum[offset+2])&0xff)<<8 |
		(uint32(sum[offset+3]) & 0xff)

	mod := uint32(math.Pow10(digits))
	code := strconv.Itoa(int(binCode % mod))
	if len(code) < digits {
		code = strings.Repeat("0", digits-len(code)) + code
	}
	return code, nil
}

func normalizeAlgorithm(algorithm Algorithm) (Algorithm, error) {
	if algorithm == "" {
		return AlgorithmSHA1, nil
	}
	switch Algorithm(strings.ToUpper(string(algorithm))) {
	case AlgorithmSHA1:
		return AlgorithmSHA1, nil
	case AlgorithmSHA256:
		return AlgorithmSHA256, nil
	case AlgorithmSHA512:
		return AlgorithmSHA512, nil
	default:
		return "", fmt.Errorf("unsupported TOTP algorithm %q", algorithm)
	}
}

func newHMAC(algorithm Algorithm, secret []byte) (hash.Hash, error) {
	switch algorithm {
	case AlgorithmSHA1:
		return hmac.New(sha1.New, secret), nil
	case AlgorithmSHA256:
		return hmac.New(sha256.New, secret), nil
	case AlgorithmSHA512:
		return hmac.New(sha512.New, secret), nil
	default:
		return nil, fmt.Errorf("unsupported TOTP algorithm %q", algorithm)
	}
}
