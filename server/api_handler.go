package server

import (
	"net/http"

	"github.com/nunoOliveiraqwe/micro-proxy/internal/manager"
)

func handleHealthCheck(_ manager.SystemManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func handleGetFtsStatus(systemManager manager.SystemManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func handleCompleteFts(systemManager manager.SystemManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func handleLogin(systemManager manager.SystemManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
