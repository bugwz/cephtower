package router

import (
	"net/http"
	"sort"

	"cephtower/backend/internal/api/v1/handler"
)

const pathPrefix = "/api/v1"

type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}

func Register(mux *http.ServeMux, h *handler.Handler) {
	groups := [][]Route{
		healthRoutes(h),
		authRoutes(h),
		rbacRoutes(h),
		clusterRoutes(h),
		resourceRoutes(h),
		credentialRoutes(h),
		endpointRoutes(h),
		overviewRoutes(h),
		hostRoutes(h),
		serviceRoutes(h),
		daemonRoutes(h),
		upgradeRoutes(h),
		monitorRoutes(h),
		managerRoutes(h),
		managerModuleRoutes(h),
		deviceRoutes(h),
		osdRoutes(h),
		crushRoutes(h),
		erasureCodeProfileRoutes(h),
		poolRoutes(h),
		rbdRoutes(h),
		cephfsRoutes(h),
		rgwRoutes(h),
		nfsRoutes(h),
		smbRoutes(h),
		nvmeofRoutes(h),
		iscsiRoutes(h),
		configurationRoutes(h),
		logsRoutes(h),
		metricRoutes(h),
		alertRoutes(h),
		grafanaRoutes(h),
		auditRoutes(h),
	}
	for _, routes := range groups {
		registerRoutes(mux, h, routes)
	}
}

func registerRoutes(mux *http.ServeMux, h *handler.Handler, routes []Route) {
	for _, route := range routes {
		mux.Handle(route.Method+" "+pathPrefix+route.Path, h.PrepareRoute(route.Handler))
	}
}

func ReadRoutesForContract(h *handler.Handler) []Route {
	var all []Route
	groups := [][]Route{
		healthRoutes(h),
		authRoutes(h),
		rbacRoutes(h),
		clusterRoutes(h),
		resourceRoutes(h),
		credentialRoutes(h),
		endpointRoutes(h),
		overviewRoutes(h),
		hostRoutes(h),
		serviceRoutes(h),
		daemonRoutes(h),
		upgradeRoutes(h),
		monitorRoutes(h),
		managerRoutes(h),
		managerModuleRoutes(h),
		deviceRoutes(h),
		osdRoutes(h),
		crushRoutes(h),
		erasureCodeProfileRoutes(h),
		poolRoutes(h),
		rbdRoutes(h),
		cephfsRoutes(h),
		rgwRoutes(h),
		nfsRoutes(h),
		smbRoutes(h),
		nvmeofRoutes(h),
		iscsiRoutes(h),
		configurationRoutes(h),
		logsRoutes(h),
		metricRoutes(h),
		alertRoutes(h),
		grafanaRoutes(h),
		auditRoutes(h),
	}
	for _, group := range groups {
		all = append(all, group...)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Path+all[i].Method < all[j].Path+all[j].Method })
	return all
}
