package proxy

import (
	"context"
	"net/http"

	"github.com/nunoOliveiraqwe/torii/config"
	"github.com/nunoOliveiraqwe/torii/middleware"
	"go.uber.org/zap"
)

type globalDispatcherPortKey struct{}

type GlobalDispatcher struct {
	globalConfig       *config.GlobalConfig
	globalMwNames      []string
	registeredHandlers map[int]http.HandlerFunc
	globalChain        http.HandlerFunc
}

func (d *GlobalDispatcher) registerHandler(port int, next http.HandlerFunc) http.HandlerFunc {
	if d == nil || d.globalConfig == nil {
		return next
	} else if d.globalChain == nil {
		return next
	}
	if d.registeredHandlers == nil {
		d.registeredHandlers = make(map[int]http.HandlerFunc)
	}
	d.registeredHandlers[port] = next
	return func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), globalDispatcherPortKey{}, port))
		d.globalChain(w, r)
	}
}

func (d *GlobalDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	port, _ := r.Context().Value(globalDispatcherPortKey{}).(int)
	if handler, exists := d.registeredHandlers[port]; exists {
		handler(w, r)
	} else {
		zap.S().Errorf("No handler registered for port %d", port)
		http.Error(w, "", http.StatusNotFound)
	}
}

func initGlobalDispatcher(ctx middleware.BuildContext, global *config.GlobalConfig) (*GlobalDispatcher, error) {
	if global == nil {
		zap.S().Infof("No global dispatcher configuration provided. Skipping")
		return &GlobalDispatcher{}, nil
	}

	zap.S().Infof("Initializing global dispatcher with %d middlewares", len(global.Middlewares))

	if len(global.Middlewares) == 0 && global.TrustedProxies == nil {
		return &GlobalDispatcher{}, nil
	}

	d := &GlobalDispatcher{
		globalConfig:       global,
		registeredHandlers: make(map[int]http.HandlerFunc),
	}
	handler := d.ServeHTTP

	if len(global.Middlewares) > 0 {
		ctx = ctx.WithServerID("global").WithOverrideMetricsName("global")

		wrapped, appliedMw, err := buildMiddlewareChain(ctx, handler, global.Middlewares, global.DisableDefaults)
		if err != nil {
			return nil, err
		}
		handler = wrapped
		d.globalMwNames = middlewareNames(appliedMw)
	}

	handler = wrapTrustedProxies(ctx.Context(), handler, global.TrustedProxies)
	d.globalChain = handler
	return d, nil
}
