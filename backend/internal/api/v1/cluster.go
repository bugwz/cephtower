package v1

import (
	"net/http"
)

func registerClusterRoutes(mux *http.ServeMux, api *API) {
	mux.HandleFunc("GET /api/v1/cluster/summary", api.clusterSummary)
	mux.HandleFunc("GET /api/v1/ceph/cluster/summary", api.clusterSummary)
	mux.HandleFunc("GET /api/v1/ceph/cluster/version", api.clusterVersion)
	mux.HandleFunc("GET /api/v1/ceph/cluster/health", api.clusterHealthMinimal)
	mux.HandleFunc("GET /api/v1/ceph/cluster/health/full", api.clusterHealthFull)
}

func (api *API) clusterSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := api.ceph.ClusterSummary(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (api *API) clusterVersion(w http.ResponseWriter, r *http.Request) {
	version, err := api.ceph.Version(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"version": version})
}

func (api *API) clusterHealthMinimal(w http.ResponseWriter, r *http.Request) {
	health, err := api.ceph.HealthMinimal(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, health)
}

func (api *API) clusterHealthFull(w http.ResponseWriter, r *http.Request) {
	health, err := api.ceph.HealthFull(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, health)
}
