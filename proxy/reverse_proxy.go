package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"go.uber.org/zap"
)

func buildHttpRevProxy(backend string) (*httputil.ReverseProxy, error) {
	zap.S().Infof("Building proxy for HTTP server with target URL: %s", backend)
	parsedUrl, err := url.Parse(backend)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: rewriteProxyRequest(parsedUrl),
	}
	return proxy, nil
}

func rewriteProxyRequest(proxyUrl *url.URL) func(r *httputil.ProxyRequest) {
	return func(r *httputil.ProxyRequest) {
		r.SetURL(proxyUrl)
		r.SetXForwarded()
		r.Out.Header.Set("X-Origin-Host", proxyUrl.Host)
		zap.S().Debugf("Rewriting request to target: %s, X-Forwarded-For: %s", proxyUrl.String(), r.Out.Header.Get("X-Forwarded-For"))
	}
}

func buildDefaultHttpHandler(proxy *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zap.S().Debugf("Proxying request to target: %s, X-Forwarded-For: %s",
			r.URL.String(), r.Header.Get("X-Forwarded-For"))
		proxy.ServeHTTP(w, r)
	}
}
