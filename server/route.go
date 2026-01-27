package server

import (
	"net/http"

	"github.com/nunoOliveiraqwe/micro-proxy/internal/manager"
)

var APPLICATION_ROUTE_BASE_PATH = "/api/v1"

type ApplicationHandlerFunc func(manager manager.SystemManager) http.HandlerFunc

type ApplicationRoute struct {
	Name               string
	Description        string
	Method             string
	Pattern            string
	IsAllowedAfterFTS  bool
	IsAllowedBeforeFTS bool
	IsSecure           bool
	HandlerFunc        ApplicationHandlerFunc
}

var externalRoutes = []ApplicationRoute{
	{
		Name:               "Healthcheck",
		Description:        "Healthcheck endpoint",
		Method:             "GET",
		Pattern:            "/healthcheck",
		IsAllowedAfterFTS:  true,
		IsAllowedBeforeFTS: true,
		IsSecure:           false,
		HandlerFunc:        handleHealthCheck,
	},
	{
		Name:               "Startup ID",
		Description:        "Startup ID endpoint. Returns the startup ID of the application. Unique per container instance and lifecycle.",
		Method:             "GET",
		Pattern:            "/startup-id",
		IsAllowedAfterFTS:  true,
		IsAllowedBeforeFTS: true,
		IsSecure:           false,
	},
	{
		Name:               "Login",
		Description:        "Login endpoint",
		Method:             "POST",
		Pattern:            "auth/login",
		IsAllowedAfterFTS:  true,
		IsAllowedBeforeFTS: false,
		IsSecure:           false,
	},
}
