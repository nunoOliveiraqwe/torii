package proxyutil

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

type ProxyOptions struct {
	DropPath  bool
	DropQuery bool
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
	}
	return proxy, nil
}

func rewriteFunc(proxyUrl *url.URL, opts ProxyOptions) func(r *httputil.ProxyRequest) {
	return func(r *httputil.ProxyRequest) {
		r.SetURL(proxyUrl)
		r.SetXForwarded()
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
