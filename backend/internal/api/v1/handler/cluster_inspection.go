package handler

import "net/http"

func (h *Handler) GetDaemonPerf(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		Name      string `json:"name"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "daemon.perf", "daemon", request.Name, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.DaemonPerf(r.Context(), request.ClusterID, request.Name)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) GetServiceDaemons(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		Name      string `json:"name"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "service.daemons", "service", request.Name, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.ServiceDaemons(r.Context(), request.ClusterID, request.Name)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) SetSubvolumeSnapshotVisibility(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("subvolume", "subvolume.snapshot_visibility", "medium")(w, r)
}

func (h *Handler) GetSubvolumeSnapshotVisibility(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		FS        string `json:"fs"`
		Subvolume string `json:"subvolume"`
		Group     string `json:"group,omitempty"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "subvolume.snapshot_visibility.get", "subvolume", request.Subvolume, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.SubvolumeSnapshotVisibility(r.Context(), request.ClusterID, request.FS, request.Subvolume, request.Group)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) GetCephFSMDS(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		FS        string `json:"fs"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "filesystem.mds", "filesystem", request.FS, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.CephFSMDS(r.Context(), request.ClusterID, request.FS)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) GetCephFSPools(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		FS        string `json:"fs"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "filesystem.pools", "filesystem", request.FS, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.CephFSPools(r.Context(), request.ClusterID, request.FS)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) GetCephFSPerformance(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		FS        string `json:"fs"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "filesystem.performance", "filesystem", request.FS, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.CephFSCounters(r.Context(), request.ClusterID, request.FS)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) ListLogs(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		Channel   string `json:"channel,omitempty"`
		Level     string `json:"level,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "log.list", "log", request.Channel, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.Logs(r.Context(), request.ClusterID, request.Channel, request.Level, request.Limit)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}
func (h *Handler) GetConfigurationOption(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		Name      string `json:"name"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "config_option.get", "config_option", request.Name, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.ConfigurationOption(r.Context(), request.ClusterID, request.Name)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) GetTelemetryStatus(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "telemetry.get", "telemetry", "status", "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.TelemetryStatus(r.Context(), request.ClusterID)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) GetTelemetryReport(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		Mode      string `json:"mode"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "telemetry.report", "telemetry", request.Mode, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.TelemetryReport(r.Context(), request.ClusterID, request.Mode)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) UpdateTelemetry(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("mgr_module", "telemetry.update", "high")(w, r)
}

func (h *Handler) UpdateTelemetryChannel(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("mgr_module", "telemetry.channel.update", "high")(w, r)
}

func (h *Handler) GetOSDInspection(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		OSDID     string `json:"osd_id"`
		Section   string `json:"section"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "osd.inspect", "osd", request.OSDID, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.OSDInspection(r.Context(), request.ClusterID, request.OSDID, request.Section)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) GetSnapshotScheduleStatus(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		FS        string `json:"fs"`
		Path      string `json:"path"`
		Subvol    string `json:"subvol,omitempty"`
		Group     string `json:"group,omitempty"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "snapshot_schedule.status", "snapshot_schedule", request.Path, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	rows, err := h.Inspection.SnapshotSchedules(r.Context(), request.ClusterID, request.FS, request.Path, request.Subvol, request.Group)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", map[string]any{"items": rows})
}

func (h *Handler) ListCephFSEntries(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		FS        string `json:"fs"`
		Path      string `json:"path,omitempty"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "cephfs_entry.list", "cephfs_entry", request.Path, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	if !h.ensureResourceCapability(w, r, request.ClusterID, "cephfs_entry") {
		return
	}
	result, err := h.Inspection.CephFSDirectories(r.Context(), request.ClusterID, request.FS, request.Path)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) ListCephFSEntrySnapshots(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		FS        string `json:"fs"`
		Path      string `json:"path"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "cephfs_entry_snapshot.list", "cephfs_entry", request.Path, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	if !h.ensureResourceCapability(w, r, request.ClusterID, "cephfs_entry") {
		return
	}
	result, err := h.Inspection.CephFSSnapshots(r.Context(), request.ClusterID, request.FS, request.Path)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}
