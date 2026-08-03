package handler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/security"
	externalservice "cephtower/backend/internal/service/external"
	mutationservice "cephtower/backend/internal/service/mutation"
	"cephtower/backend/internal/store"
)

type resourceDTO struct {
	Kind            string    `json:"kind"`
	NaturalKey      string    `json:"natural_key"`
	Name            *string   `json:"name"`
	Status          *string   `json:"status"`
	ResourceVersion uint64    `json:"resource_version"`
	Source          string    `json:"source"`
	ObservedAt      time.Time `json:"observed_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Stale           bool      `json:"stale"`
	Data            any       `json:"data"`
}

func (h *Handler) ReadResource(kind string, item bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, id, ok := h.scopedBody(w, r)
		if !ok {
			return
		}
		if !h.ensureResourceCapability(w, r, id, kind) {
			return
		}
		auditClusterID := id
		if item {
			key := readResourceKey(kind, body)
			annotateAudit(r, kind+".get", kind, key, "", &auditClusterID)
			row, err := h.Database().FindResource(r.Context(), id, kind, key)
			if err != nil {
				WriteError(w, r, 404, "resource_not_found", "resource was not found", false, nil)
				return
			}
			w.Header().Set("ETag", strconv.FormatUint(row.ResourceVersion, 10))
			items := []resourceDTO{toResourceDTO(row)}
			if err := h.decorateResourceDTOs(r.Context(), id, kind, items); err != nil {
				WriteError(w, r, http.StatusInternalServerError, "store_error", err.Error(), false, nil)
				return
			}
			WriteSuccess(w, 200, "success", items[0])
			return
		}
		annotateAudit(r, kind+".list", kind, "", "", &auditClusterID)
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		after := decodeCursor(r.URL.Query().Get("cursor"))
		filter := storeResourceFilter(kind, limit, after, r, body)
		if r.URL.Query().Get("filter_options") == "1" {
			fields := requestedFilterOptionFields(r)
			options, err := h.Database().ResourceFilterOptions(r.Context(), id, filter, fields)
			if err != nil {
				WriteError(w, r, 500, "store_error", err.Error(), false, nil)
				return
			}
			WriteSuccess(w, 200, "success", map[string]any{"filter_options": options})
			return
		}
		filter.FieldValues = resourceFieldFilters(r)
		rows, err := h.Database().ListResources(r.Context(), id, filter)
		if err != nil {
			WriteError(w, r, 500, "store_error", err.Error(), false, nil)
			return
		}
		next := any(nil)
		if limit <= 0 {
			limit = 50
		}
		if len(rows) > limit {
			next = base64.RawURLEncoding.EncodeToString([]byte(strconv.FormatUint(rows[limit-1].ID, 10)))
			rows = rows[:limit]
		}
		items := make([]resourceDTO, 0, len(rows))
		var observed *time.Time
		stale := false
		for _, row := range rows {
			items = append(items, toResourceDTO(row))
			if observed == nil || row.ObservedAt.Before(*observed) {
				value := row.ObservedAt
				observed = &value
			}
			stale = stale || row.StaleAt != nil
		}
		if err := h.decorateResourceDTOs(r.Context(), id, kind, items); err != nil {
			WriteError(w, r, http.StatusInternalServerError, "store_error", err.Error(), false, nil)
			return
		}
		var staleReason any
		if stale {
			staleReason = "collection_failed_or_resource_missing"
		}
		WriteSuccess(w, 200, "success", map[string]any{"items": items, "pagination": map[string]any{"next_cursor": next}, "meta": map[string]any{"request_id": RequestID(r), "observed_at": observed, "stale": stale, "stale_reason": staleReason}})
	}
}

func (h *Handler) ensureResourceCapability(w http.ResponseWriter, r *http.Request, clusterID uint64, kind string) bool {
	endpointKind := ""
	switch kind {
	case "metric", "alert_rule":
		endpointKind = "prometheus"
	case "alert", "silence":
		endpointKind = "alertmanager"
	case "grafana":
		endpointKind = "grafana"
	case "iscsi_gateway", "iscsi_target":
		endpointKind = "iscsi"
	case "nvmeof_gateway", "nvmeof_subsystem", "nvmeof_namespace", "nvmeof_listener", "nvmeof_host", "nvmeof_connection":
		endpointKind = "nvmeof"
	case "rgw_bucket_policy":
		endpointKind = "s3"
	}
	if endpointKind != "" {
		if _, err := h.Endpoints.Endpoint(r.Context(), clusterID, endpointKind); err != nil {
			WriteError(w, r, http.StatusNotImplemented, "capability_unavailable", endpointKind+" endpoint is not configured", false, map[string]any{"capability": endpointKind})
			return false
		}
		return true
	}
	capability := ""
	switch {
	case strings.HasPrefix(kind, "rbd_"):
		capability = "rbd"
	case strings.HasPrefix(kind, "rgw_"):
		capability = "rgw_admin"
	case kind == "nfs_cluster" || kind == "nfs_export":
		capability = "nfs"
	case kind == "smb_cluster" || kind == "smb_share":
		capability = "smb"
	case kind == "cephfs_entry":
		capability = "cephfs_data_access"
	case kind == "filesystem" || kind == "subvolume_group" || kind == "subvolume" || kind == "cephfs_snapshot" || kind == "snapshot_schedule" || kind == "cephfs_authorization" || kind == "cephfs_client":
		capability = "cephfs_volume"
	}
	if capability == "" {
		return true
	}
	rows, err := h.Database().ListCapabilities(r.Context(), clusterID)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "store_error", "capability lookup failed", false, nil)
		return false
	}
	for _, row := range rows {
		if row.Name == capability && row.Supported {
			return true
		}
		if row.Name == capability {
			message := capability + " capability is unavailable"
			if row.Reason != nil && *row.Reason != "" {
				message += ": " + *row.Reason
			}
			WriteError(w, r, http.StatusNotImplemented, "capability_unavailable", message, false, map[string]any{"capability": capability})
			return false
		}
	}
	WriteError(w, r, http.StatusNotImplemented, "capability_unavailable", capability+" capability has not been detected", false, map[string]any{"capability": capability})
	return false
}

func (h *Handler) MutateResource(kind, action, risk string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = risk
		var body map[string]any
		if !DecodeStrict(w, r, &body) {
			return
		}
		id, ok := requiredUintBody(w, r, body, "cluster_id")
		if !ok {
			return
		}
		if strings.HasPrefix(action, "rgw_bucket.") {
			if _, err := h.Endpoints.Endpoint(r.Context(), id, "s3"); err != nil {
				WriteError(w, r, http.StatusNotImplemented, "capability_unavailable", "s3 endpoint is not configured", false, map[string]any{"capability": "s3"})
				return
			}
		} else if !h.ensureResourceCapability(w, r, id, kind) {
			return
		}
		if err := ValidateMutationRequest(action, body); err != nil {
			WriteError(w, r, http.StatusBadRequest, "invalid_request", err.Error(), false, nil)
			return
		}
		var generation *uint64
		if value := r.Header.Get("If-Match"); value != "" {
			parsed, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				WriteError(w, r, 400, "invalid_request", "If-Match must be a resource version", false, nil)
				return
			}
			generation = &parsed
		}
		resourceKey := resourceKey(kind, action, r, body)
		annotateAudit(r, action, kind, resourceKey, risk, &id)
		if generation != nil {
			if err := h.checkResourceGeneration(r.Context(), id, kind, resourceKey, *generation); err != nil {
				WriteError(w, r, http.StatusConflict, "resource_conflict", err.Error(), false, nil)
				return
			}
		}
		result, err := h.executeMutation(r, id, kind, action, resourceKey, body)
		if err != nil {
			writeActionError(w, r, err)
			return
		}
		if err := h.persistResourceMutation(r.Context(), id, kind, action, resourceKey, body); err != nil {
			WriteError(w, r, http.StatusInternalServerError, "store_error", err.Error(), false, nil)
			return
		}
		WriteSuccess(w, http.StatusOK, "success", result)
	}
}

func (h *Handler) persistResourceMutation(ctx context.Context, clusterID uint64, kind, action, auditKey string, body map[string]any) error {
	key := resourceLookupKey(kind, auditKey)
	if strings.HasSuffix(action, ".delete") || strings.HasSuffix(action, ".purge") {
		return h.Database().DeleteResourceState(ctx, clusterID, kind, key)
	}
	if !strings.HasSuffix(action, ".create") && !strings.HasSuffix(action, ".update") &&
		!strings.HasSuffix(action, ".set") && !strings.HasSuffix(action, ".quota") &&
		!strings.HasSuffix(action, ".clone") && !strings.HasSuffix(action, ".restore") {
		return nil
	}
	configBody := resourceConfigurationBody(kind, action, body)
	redacted, err := security.RedactJSON(configBody)
	if err != nil {
		return err
	}
	configured, err := json.Marshal(redacted)
	if err != nil {
		return err
	}
	return h.Database().SaveResourceConfiguration(ctx, clusterID, kind, key, string(configured))
}

func resourceConfigurationBody(kind, action string, body map[string]any) map[string]any {
	if kind != "pool" {
		return body
	}
	clean := func(names ...string) map[string]any {
		result := map[string]any{}
		for _, name := range names {
			if value, exists := body[name]; exists {
				result[name] = value
			}
		}
		return result
	}
	if strings.HasSuffix(action, ".create") {
		return clean("name", "pool_type", "pg_num", "pg_autoscale_mode", "size", "applications", "erasure_code_profile", "crush_rule", "compression_mode", "quota_max_bytes", "quota_max_objects", "quota_unit")
	}
	if !strings.HasSuffix(action, ".update") {
		return body
	}
	result := clean("applications")
	if operation, _ := body["operation"].(string); operation == "quota" {
		if field, _ := body["field"].(string); field == "max_bytes" || field == "max_objects" {
			if value, exists := body["value"]; exists {
				result["quota_"+field] = value
			}
		}
		if value, exists := body["quota_unit"]; exists {
			result["quota_unit"] = value
		}
		return result
	}
	if operation, _ := body["operation"].(string); operation == "application" {
		return result
	}
	if operation, _ := body["operation"].(string); operation == "rename" {
		return clean("name")
	}
	field, _ := body["field"].(string)
	if field == "" {
		return clean("name", "pool_type", "pg_num", "pg_autoscale_mode", "size", "applications", "erasure_code_profile", "crush_rule", "compression_mode", "quota_max_bytes", "quota_max_objects", "quota_unit")
	}
	if value, exists := body["value"]; exists {
		result[field] = value
	}
	return result
}

func (h *Handler) checkResourceGeneration(ctx context.Context, clusterID uint64, kind, resourceKey string, generation uint64) error {
	row, err := h.Database().FindResource(ctx, clusterID, kind, resourceLookupKey(kind, resourceKey))
	if errors.Is(err, store.ErrRecordNotFound) && generation == 0 {
		return nil
	}
	if err != nil {
		return err
	}
	if row.ResourceVersion != generation {
		return errors.New("resource generation changed")
	}
	return nil
}

func (h *Handler) executeMutation(r *http.Request, clusterID uint64, kind, action, resourceKey string, body map[string]any) (cephdomain.ActionResult, error) {
	if action == "cluster.refresh" && h.Reconciler != nil {
		modules := stringSliceBody(body, "modules")
		return h.Reconciler.Refresh(r.Context(), clusterID, modules)
	}
	if externalservice.Supports(action) {
		if h.External == nil {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "capability_unavailable", Message: "external action is unavailable"}
		}
		return h.External.Execute(r.Context(), externalservice.Request{ClusterID: clusterID, Action: action, ResourceKey: resourceKey, Parameters: body})
	}
	if h.Mutations == nil {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "capability_unavailable", Message: "native action is unavailable"}
	}
	result, err := h.Mutations.Execute(r.Context(), mutationservice.Request{ClusterID: clusterID, Action: action, ResourceKey: resourceKey, Parameters: body})
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	return result, nil
}

func writeActionError(w http.ResponseWriter, r *http.Request, err error) {
	var actionError *cephdomain.ActionError
	if errors.As(err, &actionError) {
		status := http.StatusBadGateway
		switch actionError.Code {
		case "invalid_request", "invalid_credential", "invalid_endpoint":
			status = http.StatusBadRequest
		case "endpoint_unavailable", "capability_unavailable":
			status = http.StatusNotImplemented
		case "resource_conflict":
			status = http.StatusConflict
		}
		WriteError(w, r, status, actionError.Code, actionError.Message, actionError.Retryable, actionError.Details)
		return
	}
	WriteError(w, r, http.StatusBadGateway, "action_failed", err.Error(), true, nil)
}

func stringSliceBody(body map[string]any, key string) []string {
	raw, ok := body[key].([]any)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		if value, ok := item.(string); ok {
			values = append(values, value)
		}
	}
	return values
}

func resourceLookupKey(kind, resourceKey string) string {
	path := strings.Split(strings.Trim(resourceKey, "/"), "/")
	after := func(segment string) string {
		for index := 0; index+1 < len(path); index++ {
			if path[index] == segment {
				return path[index+1]
			}
		}
		return ""
	}
	lastResource := func() string {
		for index := len(path) - 1; index >= 0; index-- {
			switch path[index] {
			case "action", "delete", "purge", "zap", "commit", "clone":
				continue
			default:
				return path[index]
			}
		}
		return ""
	}
	switch kind {
	case "device":
		encoded := after("device")
		decoded, err := base64.RawURLEncoding.Strict().DecodeString(encoded)
		if err == nil {
			parts := bytes.SplitN(decoded, []byte{0}, 2)
			if len(parts) == 2 {
				return string(parts[0]) + ":" + string(parts[1])
			}
		}
		return encoded
	case "rbd_image":
		return after("image")
	case "rbd_snapshot":
		return after("image") + "@" + after("snapshot")
	case "rbd_namespace":
		return after("namespace") + "/" + lastResource()
	case "subvolume_group":
		return after("filesystem") + "/" + after("subvolume-group")
	case "subvolume":
		return after("filesystem") + "/" + after("subvolume")
	case "cephfs_snapshot":
		return after("filesystem") + "/" + after("subvolume") + "/" + after("snapshot")
	case "filesystem":
		return after("filesystem")
	case "osd":
		return after("osd")
	case "pool":
		return after("pool")
	case "host":
		return after("host")
	case "service":
		return after("service")
	case "crush_rule":
		return after("crush-rule")
	case "erasure_code_profile":
		return after("erasure-code-profile")
	case "nfs_cluster", "smb_cluster":
		return after("cluster")
	case "nfs_export":
		return after("export")
	case "smb_share":
		return after("share")
	case "rgw_bucket":
		return after("bucket")
	case "nvmeof_subsystem":
		return after("subsystem")
	case "nvmeof_namespace":
		return after("namespace")
	case "nvmeof_listener":
		return after("listener")
	case "nvmeof_host":
		return after("host")
	case "iscsi_target":
		return after("target")
	case "upgrade":
		return "upgrade"
	default:
		return lastResource()
	}
}

func toResourceDTO(row store.CephEntityRecord) resourceDTO {
	data := map[string]any{}
	_ = json.Unmarshal([]byte(row.DiscoveredData), &data)
	if row.ConfiguredData != nil {
		var configured map[string]any
		if err := json.Unmarshal([]byte(*row.ConfiguredData), &configured); err == nil {
			for field, value := range configured {
				data[field] = value
			}
		}
	}
	return resourceDTO{Kind: row.Kind, NaturalKey: row.NaturalKey, Name: row.Name, Status: row.Status, ResourceVersion: row.ResourceVersion, Source: row.Source, ObservedAt: row.ObservedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, Stale: row.StaleAt != nil, Data: data}
}

func (h *Handler) decorateResourceDTOs(ctx context.Context, clusterID uint64, kind string, items []resourceDTO) error {
	if kind != "host" || h.HostProfiles == nil || len(items) == 0 {
		return nil
	}
	hostRows, err := h.HostProfiles.List(ctx, clusterID)
	if err != nil {
		return err
	}
	sshByHost := make(map[string]any, len(hostRows))
	for _, row := range hostRows {
		sshByHost[row.Hostname] = row
	}
	for index := range items {
		data, ok := items[index].Data.(map[string]any)
		if !ok {
			continue
		}
		hostname := ""
		if items[index].Name != nil {
			hostname = *items[index].Name
		}
		if hostname == "" {
			if value, ok := data["hostname"].(string); ok {
				hostname = value
			}
		}
		if ssh, ok := sshByHost[hostname]; ok {
			data["host_ssh"] = ssh
			data["host_ssh_configured"] = true
		} else {
			data["host_ssh_configured"] = false
		}
	}
	return nil
}

func decodeCursor(value string) uint64 {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return 0
	}
	id, _ := strconv.ParseUint(string(decoded), 10, 64)
	return id
}

func storeResourceFilter(kind string, limit int, after uint64, r *http.Request, body map[string]any) store.ResourceFilter {
	filter := store.ResourceFilter{Kind: kind, Name: r.URL.Query().Get("name"), Status: r.URL.Query().Get("status"), Limit: limit, AfterID: after}
	switch kind {
	case "device":
		filter.ParentKind, filter.ParentKey = "host", optionalStringBody(body, "host", "hostname", "name")
	case "rbd_snapshot":
		filter.ParentKind, filter.ParentKey = "rbd_image", optionalStringBody(body, "image_spec")
	case "subvolume_group", "subvolume", "snapshot_schedule", "cephfs_authorization", "cephfs_client", "cephfs_entry":
		filter.ParentKind, filter.ParentKey = "filesystem", optionalStringBody(body, "fs", "name")
	case "cephfs_snapshot":
		filter.ParentKind, filter.ParentKey = "subvolume", optionalStringBody(body, "fs")+"/"+optionalStringBody(body, "subvolume")
	case "nfs_export":
		filter.ParentKind, filter.ParentKey = "nfs_cluster", optionalStringBody(body, "cluster")
	case "smb_share":
		filter.ParentKind, filter.ParentKey = "smb_cluster", optionalStringBody(body, "cluster")
	case "nvmeof_namespace", "nvmeof_listener", "nvmeof_host", "nvmeof_connection":
		filter.ParentKind, filter.ParentKey = "nvmeof_subsystem", optionalStringBody(body, "nqn", "subsystem_nqn")
	}
	if filter.ParentKey == "" {
		filter.ParentKind = ""
	}
	return filter
}

func requestedFilterOptionFields(r *http.Request) []string {
	raw := r.URL.Query().Get("fields")
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	fields := make([]string, 0, len(parts))
	for _, part := range parts {
		field := strings.TrimSpace(part)
		if validResourceField(field) {
			fields = append(fields, field)
		}
	}
	return fields
}

func resourceFieldFilters(r *http.Request) map[string][]string {
	values := r.URL.Query()
	filters := map[string][]string{}
	for key, selected := range values {
		field, ok := strings.CutPrefix(key, "filter.")
		if !ok || !validResourceField(field) {
			continue
		}
		for _, value := range selected {
			value = strings.TrimSpace(value)
			if value != "" {
				filters[field] = append(filters[field], value)
			}
		}
	}
	if len(filters) == 0 {
		return nil
	}
	return filters
}

func validResourceField(field string) bool {
	if field == "" || len(field) > 96 {
		return false
	}
	for _, char := range field {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' {
			continue
		}
		return false
	}
	return true
}

func readResourceKey(kind string, body map[string]any) string {
	switch kind {
	case "overview":
		return "overview"
	case "health_check":
		return optionalStringBody(body, "code")
	case "rgw_status":
		return "status"
	case "osd_flag":
		return "flags"
	case "rbd_mirroring":
		return "mirroring"
	case "upgrade":
		return "upgrade"
	case "host":
		return optionalStringBody(body, "host", "hostname", "name")
	case "service", "daemon", "mgr", "mon", "mds", "crush_rule", "erasure_code_profile", "nfs_cluster", "smb_cluster", "rgw_realm", "rgw_zonegroup", "rgw_zone":
		return optionalStringBody(body, "name")
	case "mgr_module":
		return optionalStringBody(body, "name")
	case "osd":
		return optionalStringBody(body, "osd_id")
	case "device":
		return optionalStringBody(body, "device_id", "device")
	case "pool":
		return optionalStringBody(body, "pool", "name")
	case "rbd_image":
		return optionalStringBody(body, "image_spec")
	case "rbd_snapshot":
		return optionalStringBody(body, "image_spec") + "@" + optionalStringBody(body, "snap", "name")
	case "rbd_namespace":
		return optionalStringBody(body, "pool") + "/" + optionalStringBody(body, "namespace", "name")
	case "filesystem":
		return optionalStringBody(body, "fs", "name")
	case "subvolume_group":
		return optionalStringBody(body, "fs") + "/" + optionalStringBody(body, "group", "name")
	case "subvolume":
		return optionalStringBody(body, "fs") + "/" + optionalStringBody(body, "subvolume", "name")
	case "cephfs_snapshot":
		return optionalStringBody(body, "fs") + "/" + optionalStringBody(body, "subvolume") + "/" + optionalStringBody(body, "snap", "name")
	case "nfs_export":
		return optionalStringBody(body, "export_id")
	case "smb_share":
		return optionalStringBody(body, "share_id")
	case "rgw_user", "rgw_key":
		return optionalStringBody(body, "uid")
	case "rgw_account":
		return optionalStringBody(body, "account_id", "id")
	case "rgw_role":
		return optionalStringBody(body, "name")
	case "rgw_bucket", "rgw_bucket_policy":
		return optionalStringBody(body, "bucket_id", "name")
	case "nvmeof_subsystem":
		return optionalStringBody(body, "nqn", "subsystem_nqn")
	case "nvmeof_namespace":
		return optionalStringBody(body, "nsid")
	case "nvmeof_listener":
		return optionalStringBody(body, "listener_id")
	case "nvmeof_host":
		return optionalStringBody(body, "host_nqn")
	case "iscsi_target":
		return optionalStringBody(body, "iqn")
	case "config_value":
		return optionalStringBody(body, "who") + ":" + optionalStringBody(body, "name")
	default:
		return optionalStringBody(body, "name", "id")
	}
}

func resourceKey(kind, action string, r *http.Request, body map[string]any) string {
	pathValue := func(names ...string) string {
		for _, name := range names {
			if body != nil {
				if value := optionalStringBody(body, name); value != "" {
					return value
				}
				if raw, ok := body[name]; ok {
					if id, err := uintFromJSON(raw); err == nil {
						return strconv.FormatUint(id, 10)
					}
				}
			}
		}
		return ""
	}
	segments := func(values ...string) string {
		parts := make([]string, 0, len(values))
		for _, value := range values {
			if value != "" {
				parts = append(parts, value)
			}
		}
		return strings.Join(parts, "/")
	}
	switch kind {
	case "overview":
		return "overview"
	case "health_check":
		return segments("health", strings.TrimPrefix(action, "health."), pathValue("code"))
	case "rgw_status":
		return "status"
	case "osd_flag":
		return "osd-flag"
	case "osd_deployment":
		if strings.HasSuffix(action, ".preview") {
			return "osd-deployment/preview"
		}
		return "osd-deployment"
	case "rbd_mirroring":
		return "rbd/mirroring"
	case "mon":
		return "monitor/action"
	case "host":
		name := pathValue("host", "hostname", "name")
		if action == "host.action" {
			return segments("host", name, "action")
		}
		return segments("host", name)
	case "service":
		return segments("service", pathValue("name", "service_id"))
	case "daemon":
		return segments("daemon", pathValue("name"), "action")
	case "mgr":
		return segments("manager", pathValue("name"), "fail")
	case "mgr_module":
		return segments("manager-module", pathValue("name"))
	case "upgrade":
		return strings.TrimPrefix(action, "upgrade.")
	case "crush_rule":
		return segments("crush-rule", pathValue("name"))
	case "erasure_code_profile":
		return segments("erasure-code-profile", pathValue("name"))
	case "nfs_cluster":
		return segments("nfs", "cluster", pathValue("name"))
	case "smb_cluster":
		return segments("smb", "cluster", pathValue("name"))
	case "osd":
		if action == "osd.action" {
			return segments("osd", pathValue("osd_id"), "action")
		}
		return segments("osd", pathValue("osd_id"))
	case "device":
		if action == "device.zap" {
			return segments("device", pathValue("host"), deviceResourceID(pathValue("host"), pathValue("device"), pathValue("device_id")), "zap")
		}
		return segments("device", pathValue("host"), pathValue("device_id", "device"), "identify")
	case "pool":
		return segments("pool", pathValue("pool", "name"))
	case "rbd_image":
		key := segments("rbd", "image", pathValue("image_spec"))
		if action == "rbd_image.action" {
			key = segments(key, "action")
		}
		return key
	case "rbd_snapshot":
		key := segments("rbd", "image", pathValue("image_spec"), "snapshot", pathValue("snap", "name"))
		if action == "rbd_snapshot.action" {
			key = segments(key, "action")
		}
		return key
	case "rbd_namespace":
		return segments("rbd", "namespace", pathValue("pool"), pathValue("namespace", "name"))
	case "rbd_trash":
		if action == "rbd_trash.restore" {
			return segments("rbd", "trash", pathValue("image_id"), "restore")
		}
		if action == "rbd_trash.purge" {
			return "rbd/trash/purge"
		}
		return segments("rbd", "trash", pathValue("image_id"))
	case "rbd_group":
		return segments("rbd", "group", pathValue("pool"), pathValue("name"))
	case "filesystem":
		return segments("filesystem", pathValue("fs", "name"))
	case "subvolume_group":
		return segments("filesystem", pathValue("fs"), "subvolume-group", pathValue("group", "name"))
	case "subvolume":
		return segments("filesystem", pathValue("fs"), "subvolume", pathValue("subvolume", "name"))
	case "cephfs_client":
		return segments("filesystem", pathValue("fs"), "client", pathValue("client_id"))
	case "cephfs_snapshot":
		key := segments("filesystem", pathValue("fs"), "subvolume", pathValue("subvolume"), "snapshot", pathValue("snap", "name"))
		if action == "cephfs_snapshot.clone" {
			key = segments(key, "clone")
		}
		return key
	case "snapshot_schedule":
		return segments("filesystem", pathValue("fs"), "snapshot-schedule")
	case "cephfs_authorization":
		return segments("filesystem", pathValue("fs"), "authorization")
	case "cephfs_entry":
		return segments("filesystem", pathValue("fs"), "entry")
	case "nfs_export":
		return segments("nfs", "export", pathValue("export_id"))
	case "smb_share":
		return segments("smb", "share", pathValue("share_id"))
	case "rgw_user":
		return segments("rgw", "user", pathValue("uid"))
	case "rgw_key":
		return segments("rgw", "user", pathValue("uid"), "key")
	case "rgw_account":
		return segments("rgw", "account", pathValue("account_id", "id"))
	case "rgw_role":
		return segments("rgw", "role", pathValue("name"))
	case "rgw_bucket":
		return segments("rgw", "bucket", pathValue("bucket_id", "name"))
	case "rgw_bucket_policy":
		return segments("rgw", "bucket", pathValue("bucket_id"), "policy")
	case "rgw_realm":
		return segments("rgw", "realm", pathValue("name"))
	case "rgw_zonegroup":
		return segments("rgw", "zonegroup", pathValue("name"))
	case "rgw_zone":
		return segments("rgw", "zone", pathValue("name"))
	case "rgw_period":
		return "rgw/period/commit"
	case "nvmeof_subsystem":
		return segments("nvmeof", "subsystem", pathValue("nqn", "subsystem_nqn"))
	case "nvmeof_namespace":
		return segments("nvmeof", "subsystem", pathValue("nqn", "subsystem_nqn"), "namespace", pathValue("nsid"))
	case "nvmeof_listener":
		return segments("nvmeof", "subsystem", pathValue("nqn", "subsystem_nqn"), "listener", pathValue("listener_id"))
	case "nvmeof_host":
		return segments("nvmeof", "subsystem", pathValue("nqn", "subsystem_nqn"), "host", pathValue("host_nqn"))
	case "iscsi_target":
		return segments("iscsi", "target", pathValue("iqn"))
	case "config_value":
		return segments("configuration", "value", pathValue("who"), pathValue("name"))
	case "silence":
		return segments("silence", pathValue("silence_id"))
	default:
		if value := pathValue("name", "id"); value != "" {
			return value
		}
		return strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/"), "/")
	}
}

func deviceResourceID(host, device, fallback string) string {
	if host == "" || device == "" {
		return fallback
	}
	return base64.RawURLEncoding.EncodeToString([]byte(host + "\x00" + device))
}
