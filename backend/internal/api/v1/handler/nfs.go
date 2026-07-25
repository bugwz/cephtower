package handler

import "net/http"

// NFS endpoints.
func (h *Handler) GetNFSSettings(w http.ResponseWriter, r *http.Request) {
	h.GetSettingGroup(w, r)
}

func (h *Handler) UpdateNFSSettings(w http.ResponseWriter, r *http.Request) {
	h.UpdateSettingGroup(w, r)
}

func (h *Handler) ListNFSClusters(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nfs-ganesha/cluster")
}

func (h *Handler) ListNFSExports(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nfs-ganesha/export")
}

func (h *Handler) CreateNFSExport(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nfs-ganesha/export")
}

func (h *Handler) DeleteNFSExport(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nfs-ganesha/export/{cluster_id}/{id}")
}

type NFSSettingsRequest = RawJSONRequest
type NFSSettingsResponse = RawJSONResponse
type NFSClusterResponse = RawJSONResponse
type NFSExportRequest = RawJSONRequest
type NFSExportResponse = RawJSONResponse
