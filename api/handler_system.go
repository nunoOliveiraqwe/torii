package api

import (
	"net/http"
	"strconv"

	"github.com/nunoOliveiraqwe/torii/internal/app"
	"github.com/nunoOliveiraqwe/torii/middleware"
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
		limit, offset := resolveRecentLogPage(r, svc.GetSystemHealth().ErrorLogCapacity)
		WriteResponseAsJSON(sliceRecentLogs(svc.GetRecentErrors(limit+offset), offset, limit), w)
	}
}

func handleGetRecentRequests(svc app.SystemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := middleware.GetRequestLoggerFromContext(r)
		logger.Debug("Fetching recent request logs")
		limit, offset := resolveRecentLogPage(r, svc.GetSystemHealth().RequestLogCapacity)
		WriteResponseAsJSON(sliceRecentLogs(svc.GetRecentRequests(limit+offset), offset, limit), w)
	}
}

func handleGetRecentBlocked(svc app.SystemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := middleware.GetRequestLoggerFromContext(r)
		logger.Debug("Fetching recent blocked entries")
		limit, offset := resolveRecentLogPage(r, svc.GetSystemHealth().BlockedLogCapacity)
		WriteResponseAsJSON(sliceRecentLogs(svc.GetRecentBlockedEntries(limit+offset), offset, limit), w)
	}
}

func resolveRecentLogPage(r *http.Request, capacity int) (int, int) {
	limit := capacity
	if limit <= 0 {
		limit = 1000
	}
	rawLimit := r.URL.Query().Get("limit")
	if rawLimit != "" {
		requestedLimit, err := strconv.Atoi(rawLimit)
		if err == nil && requestedLimit > 0 && requestedLimit < limit {
			limit = requestedLimit
		}
	}

	offset := 0
	rawOffset := r.URL.Query().Get("offset")
	if rawOffset != "" {
		requestedOffset, err := strconv.Atoi(rawOffset)
		if err == nil && requestedOffset > 0 {
			offset = requestedOffset
		}
	}
	if offset > capacity && capacity > 0 {
		offset = capacity
	}
	if capacity > 0 && limit > capacity-offset {
		limit = capacity - offset
	}
	return limit, offset
}

func resolveRecentLogLimit(r *http.Request, capacity int) int {
	limit, _ := resolveRecentLogPage(r, capacity)
	return limit
}

func sliceRecentLogs[T any](entries []T, offset, limit int) []T {
	if offset >= len(entries) || limit <= 0 {
		return []T{}
	}
	end := offset + limit
	if end > len(entries) {
		end = len(entries)
	}
	return entries[offset:end]
}
