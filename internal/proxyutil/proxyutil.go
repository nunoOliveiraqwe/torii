package proxyutil

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
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
		Rewrite:   rewriteFunc(parsedUrl, opts),
		Transport: sharedTransport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			zap.S().Warnf("proxy error: backend=%s path=%s err=%v", parsedUrl.Host, r.URL.Path, err)
			w.WriteHeader(http.StatusBadGateway)
		},
	}
	return proxy, nil
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
