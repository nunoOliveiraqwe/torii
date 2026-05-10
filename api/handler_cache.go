package api

import (
	"net/http"

	"github.com/nunoOliveiraqwe/torii/internal/app"
	cacheSub "github.com/nunoOliveiraqwe/torii/internal/subsystem/cache"
	"github.com/nunoOliveiraqwe/torii/middleware"
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
