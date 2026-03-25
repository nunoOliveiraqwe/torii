package middleware

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

var logEntryContextKey = "logger"

type zapLogFormatter struct {
	logger *zap.Logger
}

func newZapLogFormatter() *zapLogFormatter {
	return &zapLogFormatter{
		logger: zap.L(),
	}
}

func (z *zapLogFormatter) LogRequest(r *http.Request) {
	ctx := r.Context()
	reqId := ""
	if ctx != nil {
		reqId = GetRequestIDFromContext(ctx)
	} else {
		ctx = context.Background()
	}
	log := z.logger.With(
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("request_id", reqId),
		zap.String("user_agent", r.UserAgent()),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("host", r.Host),
	)
	log.Info("Incoming request")
	ctx = context.WithValue(ctx, logEntryContextKey, log)
	*r = *r.WithContext(ctx)
}

func GetRequestLoggerFromContext(r *http.Request) *zap.Logger {
	ctx := r.Context()
	if ctx == nil || ctx.Value(logEntryContextKey) == nil {
		return zap.L()
	}
	log := ctx.Value(logEntryContextKey)
	if logEntry, ok := log.(*zap.Logger); ok {
		return logEntry
	}
	return zap.L()
}

func RequestLoggerMiddleware(_ context.Context, next http.HandlerFunc, _ Config) http.HandlerFunc {
	newZapLogFormatter := newZapLogFormatter()
	return func(w http.ResponseWriter, r *http.Request) {
		newZapLogFormatter.LogRequest(r)
		next.ServeHTTP(w, r)
	}
}
