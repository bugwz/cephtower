package router

import (
	"net/http"

	"cephtower/backend/internal/api/v1/handler"
)

const pathPrefix = "/api/v1"

type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}

// Register binds the v1 route table to the versioned HTTP handlers.
func Register(mux *http.ServeMux, h *handler.Handler) {
	registerRoutes(mux, healthRoutes(h))
	registerRoutes(mux, authRoutes(h))
	registerRoutes(mux, setupRoutes(h))
	registerRoutes(mux, systemConfigRoutes(h))
	registerRoutes(mux, clusterRoutes(h))
	registerRoutes(mux, hostRoutes(h))
	registerRoutes(mux, osdRoutes(h))
	registerRoutes(mux, daemonRoutes(h))
	registerRoutes(mux, proxyRoutes(h))
	registerRoutes(mux, settingRoutes(h))
	registerRoutes(mux, grafanaRoutes(h))
	registerRoutes(mux, prometheusRoutes(h))
	registerRoutes(mux, rgwRoutes(h))
	registerRoutes(mux, iscsiRoutes(h))
	registerRoutes(mux, nfsRoutes(h))
	registerRoutes(mux, dashboardOperationsRoutes(h))
	registerRoutes(mux, configurationRoutes(h))
	registerRoutes(mux, nvmeofRoutes(h))
	registerRoutes(mux, smbRoutes(h))
	registerRoutes(mux, dashboardIdentityRoutes(h))
}

func registerRoutes(mux *http.ServeMux, routes []Route) {
	for _, route := range routes {
		mux.HandleFunc(route.Method+" "+pathPrefix+route.Path, route.Handler)
	}
}
