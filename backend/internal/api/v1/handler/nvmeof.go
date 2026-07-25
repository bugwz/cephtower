package handler

import "net/http"

func (h *Handler) ListNVMeOFGateways(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/gateway")
}

func (h *Handler) ListNVMeOFGatewayGroups(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/gateway/group")
}

func (h *Handler) GetNVMeOFListenerInfo(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/gateway/listener_info/{nqn}")
}

func (h *Handler) GetNVMeOFGatewayLogLevel(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/gateway/log_level")
}

func (h *Handler) UpdateNVMeOFGatewayLogLevel(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/gateway/log_level")
}

func (h *Handler) GetNVMeOFGatewayStats(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/gateway/stats")
}

func (h *Handler) GetNVMeOFGatewayVersion(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/gateway/version")
}

func (h *Handler) GetNVMeOFSPDKLogLevel(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/spdk/log_level")
}

func (h *Handler) UpdateNVMeOFSPDKLogLevel(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/spdk/log_level")
}

func (h *Handler) DisableNVMeOFSPDKLogLevel(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/spdk/log_level/disable")
}

func (h *Handler) ListNVMeOFSubsystems(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem")
}

func (h *Handler) CreateNVMeOFSubsystem(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem")
}

func (h *Handler) GetNVMeOFSubsystem(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}")
}

func (h *Handler) DeleteNVMeOFSubsystem(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}")
}

func (h *Handler) ListNVMeOFSubsystemConnections(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/connection")
}

func (h *Handler) ListNVMeOFSubsystemHosts(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/host")
}

func (h *Handler) AddNVMeOFSubsystemHost(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/host")
}

func (h *Handler) RemoveNVMeOFSubsystemHost(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/host/{host_nqn}")
}

func (h *Handler) ChangeNVMeOFHostControllerKey(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/host/{host_nqn}/change_controller_key")
}

func (h *Handler) ChangeNVMeOFHostKey(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/host/{host_nqn}/change_key")
}

func (h *Handler) DeleteNVMeOFHostControllerKey(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/host/{host_nqn}/del_controller_key")
}

func (h *Handler) DeleteNVMeOFHostKey(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/host/{host_nqn}/del_key")
}

func (h *Handler) ListNVMeOFListeners(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/listener")
}

func (h *Handler) CreateNVMeOFListener(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/listener")
}

func (h *Handler) DeleteNVMeOFListener(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/listener/{host_name}/{traddr}/{trsvcid}")
}

func (h *Handler) ListNVMeOFNamespaces(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/namespace")
}

func (h *Handler) CreateNVMeOFNamespace(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/namespace")
}

func (h *Handler) GetNVMeOFNamespace(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/namespace/{nsid}")
}

func (h *Handler) UpdateNVMeOFNamespace(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/namespace/{nsid}")
}

func (h *Handler) DeleteNVMeOFNamespace(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/namespace/{nsid}")
}

func (h *Handler) AddNVMeOFNamespaceHost(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/namespace/{nsid}/add_host")
}

func (h *Handler) ChangeNVMeOFNamespaceLoadBalancingGroup(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/namespace/{nsid}/change_load_balancing_group")
}

func (h *Handler) ChangeNVMeOFNamespaceVisibility(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/namespace/{nsid}/change_visibility")
}

func (h *Handler) DeleteNVMeOFNamespaceHost(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/namespace/{nsid}/del_host")
}

func (h *Handler) GetNVMeOFNamespaceIOStats(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/namespace/{nsid}/io_stats")
}

func (h *Handler) RefreshNVMeOFNamespaceSize(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/namespace/{nsid}/refresh_size")
}

func (h *Handler) ResizeNVMeOFNamespace(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/namespace/{nsid}/resize")
}

func (h *Handler) SetNVMeOFNamespaceAutoResize(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/namespace/{nsid}/set_auto_resize")
}

func (h *Handler) SetNVMeOFNamespaceQoS(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/namespace/{nsid}/set_qos")
}

func (h *Handler) SetNVMeOFNamespaceRBDTrashImage(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/nvmeof/subsystem/{nqn}/namespace/{nsid}/set_rbd_trash_image")
}

type NVMeOFGatewayResponse = RawJSONResponse
type NVMeOFSubsystemRequest = RawJSONRequest
type NVMeOFSubsystemResponse = RawJSONResponse
type NVMeOFHostRequest = RawJSONRequest
type NVMeOFHostResponse = RawJSONResponse
type NVMeOFListenerRequest = RawJSONRequest
type NVMeOFListenerResponse = RawJSONResponse
type NVMeOFNamespaceRequest = RawJSONRequest
type NVMeOFNamespaceResponse = RawJSONResponse
type NVMeOFLogLevelRequest = RawJSONRequest
type NVMeOFLogLevelResponse = RawJSONResponse
