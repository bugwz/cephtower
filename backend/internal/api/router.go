package api

import (
	"context"
	"net/http"

	"cephtower/backend/internal/api/v1"
	"cephtower/backend/internal/config"
	"cephtower/backend/internal/service/ceph"
)

type apiRoute struct {
	method  string
	path    string
	handler http.HandlerFunc
}

func (s *Server) registerAPIRouter(mux *http.ServeMux) {
	api := v1.NewAPI(s.ceph, v1.Dependencies{
		CurrentConfig:     s.currentConfig,
		Database:          s.database,
		ReplaceDatabase:   s.replaceDatabase,
		ClusterDiscoverer: s.clusterDiscoverer,
		ClusterRuntimeCleaner: func(ctx context.Context, clusterID uint) error {
			return ceph.DeleteCephClusterRuntimeFiles(config.ResolveRuntimeDir(s.currentConfig()), clusterID)
		},
	})

	for _, route := range apiRouterRoutes(api) {
		mux.HandleFunc(route.method+" "+v1.PathPrefix+route.path, route.handler)
	}
}

func apiRoutes(methods []string, path string, handler http.HandlerFunc) []apiRoute {
	routes := make([]apiRoute, 0, len(methods))
	for _, method := range methods {
		routes = append(routes, apiRoute{method, path, handler})
	}
	return routes
}

func apiRouterRoutes(api *v1.API) []apiRoute {
	routes := []apiRoute{
		{"GET", "/healthz", v1.Healthz},

		{"POST", "/auth/login", api.Login},
		{"POST", "/auth/password-reset/request", api.RequestPasswordReset},
		{"POST", "/auth/password-reset/confirm", api.ConfirmPasswordReset},
		{"GET", "/auth/me", api.Me},
		{"GET", "/user", api.ListUsers},
		{"POST", "/user", api.CreateUser},
		{"PATCH", "/user/{id}", api.UpdateUser},

		{"GET", "/setup/status", api.SetupStatus},
		{"POST", "/setup/initialize", api.InitializeSetup},

		{"GET", "/system/config/setting", api.ListSystemSettings},
		{"PUT", "/system/config/setting/{key}", api.UpdateSystemSetting},
		{"POST", "/system/config/default/reset", api.ResetSystemConfigDefaults},
		{"POST", "/system/config/data-fetch/{module}/run", api.RunDataFetchModuleNow},
		{"GET", "/system/config/data-fetch/run", api.ListDataFetchRuns},

		{"GET", "/cluster", api.ListClusters},
		{"POST", "/cluster", api.CreateCluster},
		{"GET", "/cluster/{id}", api.GetCluster},
		{"PUT", "/cluster/{id}", api.UpdateCluster},
		{"DELETE", "/cluster/{id}", api.DeleteCluster},
		{"GET", "/cluster/summary", api.ClusterSummary},
		{"GET", "/cluster/version", api.ClusterVersion},
		{"GET", "/cluster/health", api.ClusterHealthMinimal},
		{"GET", "/cluster/health/full", api.ClusterHealthFull},

		{"GET", "/host", api.ListHosts},
		{"POST", "/host", api.CreateHost},
		{"GET", "/host/{hostname}", api.HostDetails},
		{"PUT", "/host/{hostname}", api.UpdateHost},
		{"DELETE", "/host/{hostname}", api.DeleteHost},
		{"GET", "/host/{hostname}/daemon", api.HostDaemons},
		{"GET", "/host/{hostname}/device", api.HostDevices},
		{"GET", "/host/{hostname}/inventory", api.HostInventory},

		{"GET", "/osd", api.ListOSDs},
		{"GET", "/osd/flag", api.OSDFlags},
		{"GET", "/osd/{id}", api.OSDDetails},
		{"GET", "/osd/{id}/device", api.ProxyCephPath},
		{"GET", "/osd/{id}/histogram", api.ProxyCephPath},
		{"PUT", "/osd/{id}/mark", api.ProxyCephPath},
		{"POST", "/osd/{id}/reweight", api.ProxyCephPath},
		{"POST", "/osd/{id}/scrub", api.ProxyCephPath},
		{"GET", "/osd/{id}/smart", api.ProxyCephPath},

		{"GET", "/monitor", api.ProxyCephPath},
		{"GET", "/mgr/module", api.ProxyCephPath},
		{"POST", "/mgr/module/{name}/enable", api.ProxyCephPath},
		{"POST", "/mgr/module/{name}/disable", api.ProxyCephPath},

		{"GET", "/daemon", api.ListDaemons},
		{"PUT", "/daemon/{name}/action", api.ApplyDaemonAction},

		{"GET", "/service", api.ProxyCephPath},
		{"POST", "/service", api.ProxyCephPath},
		{"GET", "/service/known-type", api.ProxyCephPath},
		{"GET", "/service/{name}", api.ProxyCephPath},
		{"PUT", "/service/{name}", api.ProxyCephPath},
		{"DELETE", "/service/{name}", api.ProxyCephPath},
		{"GET", "/service/{name}/daemon", api.ProxyCephPath},

		{"GET", "/pool", api.ProxyCephPath},
		{"POST", "/pool", api.ProxyCephPath},
		{"GET", "/pool/{name}", api.ProxyCephPath},
		{"PUT", "/pool/{name}", api.ProxyCephPath},
		{"DELETE", "/pool/{name}", api.ProxyCephPath},
		{"GET", "/pool/{name}/configuration", api.ProxyCephPath},

		{"GET", "/block/image", api.ProxyCephPath},
		{"POST", "/block/image", api.ProxyCephPath},
		{"GET", "/block/image/default-feature", api.ProxyCephPath},
		{"GET", "/block/image/clone-format-version", api.ProxyCephPath},
		{"GET", "/block/image/trash", api.ProxyCephPath},
		{"POST", "/block/image/trash/purge", api.ProxyCephPath},
		{"DELETE", "/block/image/trash/{image}", api.ProxyCephPath},
		{"POST", "/block/image/trash/{image}/restore", api.ProxyCephPath},
		{"GET", "/block/image/{image}", api.ProxyCephPath},
		{"PUT", "/block/image/{image}", api.ProxyCephPath},
		{"DELETE", "/block/image/{image}", api.ProxyCephPath},
		{"POST", "/block/image/{image}/copy", api.ProxyCephPath},
		{"POST", "/block/image/{image}/flatten", api.ProxyCephPath},
		{"POST", "/block/image/{image}/snapshot", api.ProxyCephPath},
		{"DELETE", "/block/image/{image}/snapshot/{snapshot}", api.ProxyCephPath},
		{"PUT", "/block/image/{image}/snapshot/{snapshot}", api.ProxyCephPath},
		{"POST", "/block/image/{image}/snapshot/{snapshot}/clone", api.ProxyCephPath},
		{"POST", "/block/image/{image}/snapshot/{snapshot}/rollback", api.ProxyCephPath},
		{"GET", "/block/mirroring/summary", api.ProxyCephPath},

		{"GET", "/filesystem", api.ProxyCephPath},
		{"POST", "/filesystem", api.ProxyCephPath},
		{"GET", "/filesystem/{id}", api.ProxyCephPath},
		{"GET", "/filesystem/{id}/client", api.ProxyCephPath},
		{"GET", "/filesystem/{id}/root", api.ProxyCephPath},
		{"GET", "/filesystem/{id}/directory", api.ProxyCephPath},
		{"GET", "/filesystem/{id}/quota", api.ProxyCephPath},
		{"PUT", "/filesystem/{id}/quota", api.ProxyCephPath},
		{"GET", "/filesystem/{id}/statfs", api.ProxyCephPath},

		{"GET", "/object/gateway", api.ProxyCephPath},
		{"GET", "/object/gateway/{id}", api.ProxyCephPath},
		{"GET", "/object/user", api.ProxyCephPath},
		{"POST", "/object/user", api.ProxyCephPath},
		{"GET", "/object/user/{uid}", api.ProxyCephPath},
		{"PUT", "/object/user/{uid}", api.ProxyCephPath},
		{"DELETE", "/object/user/{uid}", api.ProxyCephPath},
		{"GET", "/object/bucket", api.ProxyCephPath},
		{"POST", "/object/bucket", api.ProxyCephPath},
		{"GET", "/object/bucket/{bucket}", api.ProxyCephPath},
		{"PUT", "/object/bucket/{bucket}", api.ProxyCephPath},
		{"DELETE", "/object/bucket/{bucket}", api.ProxyCephPath},
		{"GET", "/object/account", api.ProxyCephPath},
		{"POST", "/object/account", api.ProxyCephPath},
		{"GET", "/object/account/{id}", api.ProxyCephPath},
		{"PUT", "/object/account/{id}", api.ProxyCephPath},
		{"DELETE", "/object/account/{id}", api.ProxyCephPath},

		{"GET", "/configuration/filter", api.ProxyCephPath},
	}
	routes = append(routes, apiRoutes([]string{"GET", "POST", "PUT"}, "/configuration", api.ProxyCephPath)...)
	routes = append(routes, apiRoutes([]string{"GET", "DELETE"}, "/configuration/{name}", api.ProxyCephPath)...)
	routes = append(routes, apiRoute{"GET", "/log", api.ProxyCephPath})
	return routes
}
