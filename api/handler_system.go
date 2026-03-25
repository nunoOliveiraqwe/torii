package api

import (
	"net/http"

	"github.com/nunoOliveiraqwe/micro-proxy/internal/app"
	"github.com/nunoOliveiraqwe/micro-proxy/middleware"
)

func handleGetSystemHealth(svc app.SystemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := middleware.GetRequestLoggerFromContext(r)
		logger.Debug("Fetching system health")
		WriteResponseAsJSON(svc.GetSystemHealth(), w)
	}
}

func handleGetRecentErrors(svc app.SystemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := middleware.GetRequestLoggerFromContext(r)
		logger.Debug("Fetching recent error logs")
		WriteResponseAsJSON(svc.GetRecentErrors(50), w)
	}
}

func handleGetRecentRequests(svc app.SystemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := middleware.GetRequestLoggerFromContext(r)
		logger.Debug("Fetching recent request logs")
		WriteResponseAsJSON(svc.GetRecentRequests(200), w)
	}
}
