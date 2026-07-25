package handler

import "net/http"

// Cluster configuration and mgr module endpoints.
func (h *Handler) ListConfiguration(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/cluster_conf")
}

func (h *Handler) CreateConfiguration(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/cluster_conf")
}

func (h *Handler) UpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/cluster_conf")
}

func (h *Handler) ListConfigurationFilter(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/cluster_conf/filter")
}

func (h *Handler) GetConfiguration(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/cluster_conf/{name}")
}

func (h *Handler) DeleteConfiguration(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/cluster_conf/{name}")
}

func (h *Handler) ListMgrModules(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/mgr/module")
}

func (h *Handler) GetMgrModule(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/mgr/module/{name}")
}

func (h *Handler) UpdateMgrModule(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/mgr/module/{name}")
}

func (h *Handler) EnableMgrModule(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/mgr/module/{name}/enable")
}

func (h *Handler) DisableMgrModule(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/mgr/module/{name}/disable")
}

func (h *Handler) GetMgrModuleOptions(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/mgr/module/{name}/options")
}

type ConfigurationRequest = RawJSONRequest
type ConfigurationResponse = RawJSONResponse
type MgrModuleRequest = RawJSONRequest
type MgrModuleResponse = RawJSONResponse
