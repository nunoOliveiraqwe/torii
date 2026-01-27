package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/nunoOliveiraqwe/micro-proxy/configuration"
	"github.com/nunoOliveiraqwe/micro-proxy/internal/manager"
	"github.com/nunoOliveiraqwe/micro-proxy/middleware"
	"go.uber.org/zap"
)

func StartServer(serverFlags *Flags, systemManager manager.SystemManager) *http.Server {
	zap.S().Infof("starting local server at if %s, port %d", serverFlags.Host, serverFlags.Port)

	if serverFlags.Port <= 0 && serverFlags.Port > 65535 {
		zap.S().Fatalf("invalid port number %d", serverFlags.Port)
		return nil
	}

	httpServer := &http.Server{
		Addr: fmt.Sprintf("%s:%d", serverFlags.Host, serverFlags.Port),
	}
	httpServer.IdleTimeout = time.Duration(serverFlags.IdleTimeoutSecs) * time.Second
	httpServer.ReadTimeout = time.Duration(serverFlags.ReadTimeoutSecs) * time.Second
	httpServer.WriteTimeout = time.Duration(serverFlags.WriteTimeoutSecs) * time.Second

	httpServer.Handler = buildMux(systemManager)
	return httpServer

}

func buildMux(manager manager.SystemManager) *http.ServeMux {
	zap.S().Debugf("building http mux")
	mux := http.NewServeMux()
	for _, route := range externalRoutes {
		zap.S().Debugf("Initializing route named %s with path %s", route.Name, route.Pattern)
		fullPathWithMethod := fmt.Sprintf("%s %s%s", route.Method, APPLICATION_ROUTE_BASE_PATH, route.Pattern)
		zap.S().Debugf("full path for route %s is %s", route.Name, fullPathWithMethod)
		routeHandlerFunc := route.HandlerFunc(manager)
		routeHandlerFunc = middleware.RequestIDMiddleware(routeHandlerFunc, configuration.Middleware{})
		routeHandlerFunc = middleware.RequestLoggerMiddleware(routeHandlerFunc, configuration.Middleware{})

		mux.HandleFunc(fullPathWithMethod, routeHandlerFunc)
		zap.S().Debugf("route %s initialized", route.Name)
	}
	return mux
}
