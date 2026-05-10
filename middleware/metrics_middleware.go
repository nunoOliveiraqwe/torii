package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/nunoOliveiraqwe/torii/internal/requestctx"
	"github.com/nunoOliveiraqwe/torii/metrics"
	"go.uber.org/zap"
)

type responseWriterWithMetrics struct {
	http.ResponseWriter
	reqMetrics  *metrics.RequestMetric
	wroteHeader bool
}

func (w *responseWriterWithMetrics) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.reqMetrics.StatusCode = statusCode
	w.reqMetrics.Is2xxResponse = statusCode >= 200 && statusCode < 300
	w.reqMetrics.Is3xxResponse = statusCode >= 300 && statusCode < 400
	w.reqMetrics.Is4xxResponse = statusCode >= 400 && statusCode < 500
	w.reqMetrics.Is5xxResponse = statusCode >= 500 && statusCode < 600
	w.ResponseWriter.WriteHeader(statusCode)
	w.wroteHeader = true
}

func (w *responseWriterWithMetrics) Write(b []byte) (int, error) {
	w.reqMetrics.BytesSent += int64(len(b))
	//htto/server.go writes the header when calling write
	// if I write twice, i get a superfluous response.WriteHeader call log which annoys me to no end
	n, err := w.ResponseWriter.Write(b)
	if err == nil {
		w.wroteHeader = true
		if w.reqMetrics.StatusCode == 0 {
			w.reqMetrics.StatusCode = 200
			w.reqMetrics.Is2xxResponse = true
		}
	}
	return n, err
}

func (w *responseWriterWithMetrics) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *responseWriterWithMetrics) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func MetricsMiddleware(contx BuildContext, next http.HandlerFunc, _ Config) http.HandlerFunc {
	reportFunc := resolveReportFunc(contx)
	return func(w http.ResponseWriter, r *http.Request) {
		if reportFunc == nil {
			next.ServeHTTP(w, r)
			return
		}
		metric := initializeRequestMetrics(r)
		responseWriter := &responseWriterWithMetrics{ResponseWriter: w,
			reqMetrics: metric}
		startTime := time.Now()
		next.ServeHTTP(responseWriter, r)
		elapsedTime := time.Since(startTime)
		if err := r.Context().Err(); errors.Is(err, context.DeadlineExceeded) {
			metric.IsTimedOut = true
		}
		metric.LatencyMs = elapsedTime.Milliseconds()
		ctxStruct := requestctx.GetContextStruct(r)
		metric.Country = ctxStruct.CountryCode
		if ctxStruct.BlockInfo != nil {
			metric.IsMiddlewareBlockedRequest = true
			metric.BlockingMiddleware = ctxStruct.BlockInfo.Middleware
			metric.BlockReason = ctxStruct.BlockInfo.Reason
		}
		reportFunc(metric)
	}
}

func resolveReportFunc(ctx BuildContext) metrics.MetricsReportFunc {
	conName := ctx.OverrideMetricsName

	if conName == "" {
		var err error
		conName, err = ctx.ScopedName("metric")
		if err != nil {
			zap.S().Warnf("Failed to build connection name for metrics resolution: %v", err)
			return nil
		}
	}

	if ctx.MetricsManager == nil {
		zap.S().Warnf("Mgr not found in middleware options for metrics resolution")
		return nil
	}
	if ctx.ServerID == "" {
		zap.S().Warnf("ServerId not found in middleware options for metrics resolution")
		return nil
	}
	return ctx.MetricsManager.TrackMetricsForConnection(ctx.ServerID, conName)
}

func initializeRequestMetrics(r *http.Request) *metrics.RequestMetric {
	return &metrics.RequestMetric{
		RemoteAddress: r.RemoteAddr,
		Host:          r.Host,
		BytesReceived: r.ContentLength,
		BytesSent:     0,
		IsTimedOut:    false,
		Path:          r.URL.Path,
		Method:        r.Method,
	}
}
