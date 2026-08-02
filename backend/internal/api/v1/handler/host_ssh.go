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

type hostSSHSaveRequest struct {
	ClusterID        uint64  `json:"cluster_id"`
	Hostname         string  `json:"hostname,omitempty"`
	Host             string  `json:"host,omitempty"`
	SSHAddress       string  `json:"ssh_address"`
	SSHPort          uint16  `json:"ssh_port,omitempty"`
	SSHUser          string  `json:"ssh_user"`
	SSHAuthMethod    string  `json:"ssh_auth_method"`
	SSHPassword      *string `json:"ssh_password,omitempty"`
	SSHPrivateKey    *string `json:"ssh_private_key,omitempty"`
	SSHKeyPassphrase *string `json:"ssh_key_passphrase,omitempty"`
	Notes            *string `json:"notes,omitempty"`
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
		ClusterID:        request.ClusterID,
		Hostname:         hostname,
		SSHAddress:       request.SSHAddress,
		SSHPort:          request.SSHPort,
		SSHUser:          request.SSHUser,
		SSHAuthMethod:    request.SSHAuthMethod,
		SSHPassword:      request.SSHPassword,
		SSHPrivateKey:    request.SSHPrivateKey,
		SSHKeyPassphrase: request.SSHKeyPassphrase,
		Notes:            request.Notes,
	})
	if err != nil {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", err.Error(), false, nil)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", view)
}
