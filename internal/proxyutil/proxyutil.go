package proxyutil

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// sharedTransport is a custom transport with sensible connection pool settings
// for reverse proxying. The stdlib default has MaxIdleConnsPerHost=2 which
// causes constant connection churn under any real concurrency.
var sharedTransport = &http.Transport{
	MaxIdleConns:        200,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:     90 * time.Second,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	TLSHandshakeTimeout: 10 * time.Second,
}

type ProxyOptions struct {
	DropPath          bool
	DropQuery         bool
	ReplaceHostHeader bool
	TLS               *ProxyTlsOptions
}

type ProxyTlsOptions struct {
	InsecureSkipVerify bool
	CaCert             string
	ClientCert         string
	ClientKey          string
}

func NewReverseProxy(backend string, opts ProxyOptions) (*httputil.ReverseProxy, error) {
	if strings.HasPrefix(backend, "http://") || strings.HasPrefix(backend, "https://") {
	} else {
		zap.S().Warnf("Proxy backend '%s' does not have a scheme, assuming http://", backend)
		backend = "http://" + backend
	}
	parsedUrl, err := url.Parse(backend)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: rewriteFunc(parsedUrl, opts),
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			zap.S().Warnf("proxy error: backend=%s path=%s err=%v", parsedUrl.Host, r.URL.Path, err)
			w.WriteHeader(http.StatusBadGateway)
		},
	}
	if opts.TLS != nil {
		transport, err := newSharedTransportWithTls(*opts.TLS)
		if err != nil {
			return nil, err
		}
		proxy.Transport = transport
	} else {
		proxy.Transport = sharedTransport
	}
	return proxy, nil
}

func newTlsConfig(opts ProxyTlsOptions) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: opts.InsecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
	}

	if opts.CaCert != "" {
		caCert, err := os.ReadFile(opts.CaCert)
		if err != nil {
			return nil, fmt.Errorf("failed to read proxy backend CA certificate %q: %w", opts.CaCert, err)
		}
		certPool, err := x509.SystemCertPool()
		if err != nil {
			certPool = x509.NewCertPool()
		}
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse proxy backend CA certificate %q", opts.CaCert)
		}
		tlsConfig.RootCAs = certPool
	}

	if opts.ClientCert != "" || opts.ClientKey != "" {
		if opts.ClientCert == "" || opts.ClientKey == "" {
			return nil, fmt.Errorf("proxy backend client certificate and key must both be configured")
		}
		clientCert, err := tls.LoadX509KeyPair(opts.ClientCert, opts.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load proxy backend client certificate/key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}

	return tlsConfig, nil
}

func newSharedTransportWithTls(opts ProxyTlsOptions) (*http.Transport, error) {
	tlsConfig, err := newTlsConfig(opts)
	if err != nil {
		return nil, err
	}

	transport := sharedTransport.Clone()
	transport.TLSClientConfig = tlsConfig
	return transport, nil
}

func rewriteFunc(proxyUrl *url.URL, opts ProxyOptions) func(r *httputil.ProxyRequest) {
	return func(r *httputil.ProxyRequest) {
		r.SetURL(proxyUrl)
		r.SetXForwarded()
		if !opts.ReplaceHostHeader {
			r.Out.Host = r.In.Host // preserve the original Host header, set url already re-writes it
		}
		r.Out.Header.Set("X-Origin-Host", proxyUrl.Host)

		if opts.DropPath {
			r.Out.URL.Path = proxyUrl.Path
			r.Out.URL.RawPath = ""
		}

		if opts.DropQuery {
			r.Out.URL.RawQuery = proxyUrl.RawQuery
		}
	}
}
