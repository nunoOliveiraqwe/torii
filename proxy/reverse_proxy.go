package proxy

import (
	"net/http"
	"net/http/httputil"

	"github.com/nunoOliveiraqwe/torii/config"
	"github.com/nunoOliveiraqwe/torii/internal/proxyutil"
	"github.com/nunoOliveiraqwe/torii/middleware"
	"go.uber.org/zap"
)

func buildHttpRevProxy(backend string, opts proxyutil.ProxyOptions) (*httputil.ReverseProxy, error) {
	zap.S().Infof("Building proxy for HTTP server with target URL: %s", backend)
	return proxyutil.NewReverseProxy(backend, opts)
}

func backendTLSOptions(tlsConf *config.BackendTlsConfig) *proxyutil.ProxyTlsOptions {
	if tlsConf == nil {
		return nil
	}
	return &proxyutil.ProxyTlsOptions{
		InsecureSkipVerify: tlsConf.InsecureSkipVerify,
		CaCert:             tlsConf.CaCert,
		ClientCert:         tlsConf.ClientCert,
		ClientKey:          tlsConf.ClientKey,
	}
}

func buildDefaultHttpHandler(proxy *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := middleware.GetRequestLoggerFromContext(r)
		logger.Debug("Proxying request to target", zap.String("target", r.URL.String()), zap.String("x-forwarded-for", r.Header.Get("X-Forwarded-For")))
		proxy.ServeHTTP(w, r)
	}
}
