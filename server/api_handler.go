package server

import (
	"net/http"

	"github.com/nunoOliveiraqwe/micro-proxy/internal/manager"
)

func handleHealthCheck(manager manager.SystemManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}
