package api

import (
	"net/http"

	"github.com/nunoOliveiraqwe/micro-proxy/internal/app"
)

func handleHealthCheck(_ app.SystemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func handleGetFtsStatus(systemService app.SystemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func handleCompleteFts(systemService app.SystemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func handleLogin(systemService app.SystemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
