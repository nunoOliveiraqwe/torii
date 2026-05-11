package api

import (
	"net/http"

	"github.com/nunoOliveiraqwe/torii/internal/app"
	cacheSub "github.com/nunoOliveiraqwe/torii/internal/subsystem/cache"
	"github.com/nunoOliveiraqwe/torii/middleware"
	"go.uber.org/zap"
)

func handleGetCacheSubsystem(service app.SystemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := middleware.GetRequestLoggerFromContext(r)
		logger.Debug("Fetching cache snapshots")
		cacheSubsystem := service.GetCacheSubsystem()
		if cacheSubsystem == nil {
			WriteResponseAsJSON([]cacheSub.SourceSnapshot{}, w)
			return
		}

		WriteResponseAsJSON(cacheSubsystem.Snapshots(), w)
	}
}

func handleDeleteCacheSubsystemEntry(service app.SystemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := middleware.GetRequestLoggerFromContext(r)
		logger.Debug("Deleting cache entry")
		cacheSubsystem := service.GetCacheSubsystem()
		if cacheSubsystem == nil {
			http.Error(w, "Cache subsystem not available", http.StatusInternalServerError)
			return
		}
		cacheName := r.PathValue("cacheID")
		if cacheName == "" {
			logger.Warn("Missing cache name in request")
			http.Error(w, "Missing cache name", http.StatusBadRequest)
			return
		}
		entryKey := r.PathValue("entryKey")
		if entryKey == "" {
			logger.Warn("Missing entry key in request")
			http.Error(w, "Missing entry key", http.StatusBadRequest)
			return
		}
		deleted := cacheSubsystem.DeleteEntryFromCache(cacheName, entryKey)
		if !deleted {
			logger.Warn("Cache entry not found for deletion", zap.String("cacheName", cacheName), zap.String("entryKey", entryKey))
			http.Error(w, "Cache entry not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
