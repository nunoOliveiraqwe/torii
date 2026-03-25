package proxy

import (
	"context"
	"net/http"

	"github.com/nunoOliveiraqwe/micro-proxy/config"
	"go.uber.org/zap"
)

type GlobalDispatcher struct {
	internalHandlers map[string]http.HandlerFunc
	next             http.Handler
}

func (d *GlobalDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, ok := d.internalHandlers[r.URL.Path]; ok {
		zap.S().Debugf("Matched internal handler for path %s", r.URL.Path)
		handler(w, r)
		return
	}
	d.next.ServeHTTP(w, r)
}

func buildGlobalDispatcher(ctx context.Context, global *config.GlobalConfig, next http.Handler) (http.Handler, error) {
	if global == nil {
		zap.S().Infof("No global dispatcher configuration provided. Skipping")
		return next, nil
	}

	zap.S().Infof("Building global dispatcher with %d internal handlers and %d middlewares",
		len(global.Handlers), len(global.Middlewares))

	d := &GlobalDispatcher{
		internalHandlers: make(map[string]http.HandlerFunc),
		next:             next,
	}

	for _, h := range global.Handlers {
		handler, err := resolveInternalHandler(h.Handler)
		if err != nil {
			zap.S().Errorf("Failed to resolve internal handler %s for path %s: %v", h.Handler, h.Path, err)
			return nil, err
		}
		d.internalHandlers[h.Path] = handler
		zap.S().Infof("Registered internal handler %s for path %s", h.Handler, h.Path)
	}

	if len(global.Middlewares) == 0 {
		return d, nil
	}

	wrapped, err := buildMiddlewareChain(ctx, d.ServeHTTP, global.Middlewares)
	if err != nil {
		return nil, err
	}
	return wrapped, nil
}

func resolveInternalHandler(name string) (http.HandlerFunc, error) {
	switch name {
	case "login":
		return loginHandler, nil
	case "logout":
		return logoutHandler, nil
	default:
		return nil, &unknownHandlerError{name: name}
	}
}

type unknownHandlerError struct {
	name string
}

func (e *unknownHandlerError) Error() string {
	return "unknown internal handler: " + e.name
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: render login page, validate credentials, issue session cookie
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: invalidate session cookie
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
