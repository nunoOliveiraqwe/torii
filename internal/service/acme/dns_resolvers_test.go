package acme

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeDNSResolvers(t *testing.T) {
	got := NormalizeDNSResolvers([]string{
		" 1.1.1.1:53 ",
		"",
		"8.8.8.8",
		"1.1.1.1:53",
		"GOOGLE-PUBLIC-DNS-A.GOOGLE.COM:53",
		"google-public-dns-a.google.com:53",
	})

	assert.Equal(t, []string{
		"1.1.1.1:53",
		"8.8.8.8",
		"GOOGLE-PUBLIC-DNS-A.GOOGLE.COM:53",
	}, got)
}
