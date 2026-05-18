package proxyutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewReverseProxyWithCustomCA(t *testing.T) {
	backend := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("proxied"))
	}))
	defer backend.Close()

	caPath := writePEMFile(t, "ca.pem", "CERTIFICATE", backend.Certificate().Raw)
	proxy, err := NewReverseProxy(backend.URL, ProxyOptions{
		TLS: &ProxyTlsOptions{CaCert: caPath},
	})
	if err != nil {
		t.Fatalf("NewReverseProxy() error = %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://proxy.test/", nil)
	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("proxy status = %d, want %d; body=%q", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if rec.Body.String() != "proxied" {
		t.Fatalf("proxy body = %q, want %q", rec.Body.String(), "proxied")
	}
}

func TestNewReverseProxyWithInsecureSkipVerify(t *testing.T) {
	backend := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer backend.Close()

	proxy, err := NewReverseProxy(backend.URL, ProxyOptions{
		TLS: &ProxyTlsOptions{InsecureSkipVerify: true},
	})
	if err != nil {
		t.Fatalf("NewReverseProxy() error = %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://proxy.test/", nil)
	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("proxy status = %d, want %d; body=%q", rec.Code, http.StatusNoContent, rec.Body.String())
	}
}

func TestNewTlsConfigRequiresClientCertAndKeyTogether(t *testing.T) {
	_, err := newTlsConfig(ProxyTlsOptions{ClientCert: "client.crt"})
	if err == nil {
		t.Fatal("newTlsConfig() error = nil, want missing key error")
	}

	_, err = newTlsConfig(ProxyTlsOptions{ClientKey: "client.key"})
	if err == nil {
		t.Fatal("newTlsConfig() error = nil, want missing cert error")
	}
}

func TestNewTlsConfigLoadsClientCertificate(t *testing.T) {
	certPEM, keyPEM := generateSelfSignedCert(t)
	certPath := writeRawFile(t, "client.crt", certPEM)
	keyPath := writeRawFile(t, "client.key", keyPEM)

	tlsConfig, err := newTlsConfig(ProxyTlsOptions{
		ClientCert: certPath,
		ClientKey:  keyPath,
	})
	if err != nil {
		t.Fatalf("newTlsConfig() error = %v", err)
	}

	if len(tlsConfig.Certificates) != 1 {
		t.Fatalf("len(Certificates) = %d, want 1", len(tlsConfig.Certificates))
	}
	if tlsConfig.MinVersion != tls.VersionTLS12 {
		t.Fatalf("MinVersion = %d, want %d", tlsConfig.MinVersion, tls.VersionTLS12)
	}
}

func writePEMFile(t *testing.T, name, typ string, der []byte) string {
	t.Helper()
	return writeRawFile(t, name, pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der}))
}

func writeRawFile(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
	return path
}

func generateSelfSignedCert(t *testing.T) ([]byte, []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey() error = %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "proxyutil-test-client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("x509.CreateCertificate() error = %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return certPEM, keyPEM
}
