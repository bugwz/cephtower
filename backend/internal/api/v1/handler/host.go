package handler

import (
	"errors"
	"net/http"

	hostprofileservice "cephtower/backend/internal/service/hostprofile"
	"cephtower/backend/internal/store"
)

type hostSSHRequest struct {
	ClusterID uint64 `json:"cluster_id"`
	Hostname  string `json:"hostname,omitempty"`
	Host      string `json:"host,omitempty"`
}

func (h *Handler) GetHostDevices(w http.ResponseWriter, r *http.Request) {
	request, ok := decodeHostDetailRequest(w, r)
	if !ok {
		return
	}
	if h.HostDetails == nil {
		WriteError(w, r, http.StatusNotImplemented, "capability_unavailable", "host device details are unavailable", false, nil)
		return
	}
	devices, err := h.HostDetails.Devices(r.Context(), request.ClusterID, request.Hostname)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", devices)
}

func (h *Handler) GetHostSMART(w http.ResponseWriter, r *http.Request) {
	request, ok := decodeHostDetailRequest(w, r)
	if !ok {
		return
	}
	if h.HostDetails == nil {
		WriteError(w, r, http.StatusNotImplemented, "capability_unavailable", "host SMART data is unavailable", false, nil)
		return
	}
	payload, err := h.HostDetails.SMART(r.Context(), request.ClusterID, request.Hostname)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", payload)
}

func decodeHostDetailRequest(w http.ResponseWriter, r *http.Request) (hostSSHRequest, bool) {
	var request hostSSHRequest
	if !DecodeStrict(w, r, &request) {
		return request, false
	}
	if request.Hostname == "" {
		request.Hostname = request.Host
	}
	if request.ClusterID == 0 || request.Hostname == "" {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", "cluster_id and hostname are required", false, nil)
		return request, false
	}
	return request, true
}

type hostSSHSaveRequest struct {
	ClusterID     uint64   `json:"cluster_id"`
	Hostname      string   `json:"hostname,omitempty"`
	Host          string   `json:"host,omitempty"`
	SSHAddress    string   `json:"ssh_address"`
	SSHPort       uint16   `json:"ssh_port,omitempty"`
	SSHUser       string   `json:"ssh_user"`
	SSHPassword   *string  `json:"ssh_password,omitempty"`
	SyncHostnames []string `json:"sync_hostnames,omitempty"`
}

func (h *Handler) GetHostSSH(w http.ResponseWriter, r *http.Request) {
	if h.HostProfiles == nil {
		WriteError(w, r, http.StatusNotImplemented, "capability_unavailable", "host settings are unavailable", false, nil)
		return
	}
	var request hostSSHRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	hostname := request.Hostname
	if hostname == "" {
		hostname = request.Host
	}
	if request.ClusterID == 0 || hostname == "" {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", "cluster_id and hostname are required", false, nil)
		return
	}
	auditClusterID := request.ClusterID
	annotateAudit(r, "host_ssh.get", "host", "host/"+hostname+"/ssh", "", &auditClusterID)
	view, err := h.HostProfiles.Get(r.Context(), request.ClusterID, hostname)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			WriteSuccess(w, http.StatusOK, "success", map[string]any{"hostname": hostname})
			return
		}
		WriteError(w, r, http.StatusInternalServerError, "store_error", err.Error(), false, nil)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", view)
}

func (h *Handler) SaveHostSSH(w http.ResponseWriter, r *http.Request) {
	if h.HostProfiles == nil {
		WriteError(w, r, http.StatusNotImplemented, "capability_unavailable", "host settings are unavailable", false, nil)
		return
	}
	var request hostSSHSaveRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	hostname := request.Hostname
	if hostname == "" {
		hostname = request.Host
	}
	auditClusterID := request.ClusterID
	annotateAudit(r, "host_ssh.save", "host", "host/"+hostname+"/ssh", "medium", &auditClusterID)
	view, err := h.HostProfiles.Save(r.Context(), hostprofileservice.SaveInput{
		ClusterID:     request.ClusterID,
		Hostname:      hostname,
		SSHAddress:    request.SSHAddress,
		SSHPort:       request.SSHPort,
		SSHUser:       request.SSHUser,
		SSHPassword:   request.SSHPassword,
		SyncHostnames: request.SyncHostnames,
	})
	if err != nil {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", err.Error(), false, nil)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", view)
}
