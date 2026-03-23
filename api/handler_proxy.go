package api

import (
	"net/http"

	"github.com/nunoOliveiraqwe/micro-proxy/internal/app"
	"github.com/nunoOliveiraqwe/micro-proxy/metrics"
	"go.uber.org/zap"
)

func handleGetProxies(systemService app.SystemService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		zap.S().Infof("Fetching configured proxy servers")
		proxies := systemService.GetConfiguredProxyServers()
		if proxies == nil {
			zap.S().Errorf("Failed to retrieve configured proxy servers")
			http.Error(writer, "Failed to retrieve configured proxy servers", http.StatusInternalServerError)
			return
		}
		zap.S().Infof("Retrieved %d configured proxy servers", len(proxies))
		WriteResponseAsJSON(proxies, writer)
	}
}

func handleGetGlobalMetrics(_ app.SystemService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		zap.S().Infof("Fetching global proxy metrics")
		globalMetrics := metrics.GlobalMetricsManager.GetGlobalMetrics()
		WriteResponseAsJSON(globalMetrics, writer)
	}
}

func handleGetMetricForConnection(_ app.SystemService) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		connectionId := request.PathValue("connectionId")
		zap.S().Infof("Fetching metric for connection %s", connectionId)
		metric := metrics.GlobalMetricsManager.GetMetricForConnection(connectionId)
		if metric == nil {
			metric = &metrics.Metric{}
		}
		WriteResponseAsJSON(metric, writer)
	}
}
