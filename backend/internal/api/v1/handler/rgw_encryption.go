package handler

import (
	"cephtower/backend/internal/service/mutation"
	operationservice "cephtower/backend/internal/service/operation"
	"net/http"
)

func (h *Handler) UpdateRGWEncryptionConfiguration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	var body map[string]any
	if !DecodeStrict(w, r, &body) {
		return
	}
	if err := ValidateMutationRequest("rgw_encryption.update", body); err != nil {
		WriteError(w, r, 400, "invalid_request", err.Error(), false, nil)
		return
	}
	id, ok := requiredUintBody(w, r, body, "cluster_id")
	if !ok {
		return
	}
	delete(body, "cluster_id")
	if err := mutation.ValidateRGWEncryptionUpdate(body); err != nil {
		writeActionError(w, r, err)
		return
	}
	if _, err := h.Clusters.Get(r.Context(), id); err != nil {
		clusterError(w, r, err)
		return
	}
	entity := body["entity"].(string)
	key := "rgw/encryption/" + entity
	annotateAudit(r, "rgw_encryption.update", "rgw_configuration", key, "high", &id)
	op, err := h.enqueueOperation(r, operationservice.EnqueueRequest{
		ClusterID: id, Action: "rgw_encryption.update", ResourceKind: "rgw_configuration",
		ResourceKey: key, LockKey: key, Risk: "high", Parameters: body,
	})
	if err != nil {
		writeOperationEnqueueError(w, r, err)
		return
	}
	WriteSuccess(w, http.StatusAccepted, "accepted", toOperationDTO(op))
}

func (h *Handler) GetRGWEncryptionConfiguration(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID      uint64 `json:"cluster_id"`
		Entity         string `json:"entity"`
		EncryptionType string `json:"encryption_type"`
		Provider       string `json:"provider"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "rgw.encryption.read", "rgw_configuration", request.Entity, "", &request.ClusterID)
	w.Header().Set("Cache-Control", "no-store")
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.RGWEncryptionConfiguration(r.Context(), request.ClusterID, request.Entity, request.EncryptionType, request.Provider)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	WriteSuccess(w, 200, "success", result)
}
