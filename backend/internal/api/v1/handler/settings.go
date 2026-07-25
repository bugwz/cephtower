package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	settingsservice "cephtower/backend/internal/service/settings"
)

type SettingUpdateRequest map[string]any
type SettingGroupUpdateRequest map[string]any

func (h *Handler) ListSettings(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	result, err := h.settings.List(r.Context(), r.URL.Query())
	if err != nil {
		writeCephError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetSetting(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	result, err := h.settings.Get(r.Context(), r.PathValue("name"))
	if err != nil {
		writeCephError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	body, ok := rawRequestBody(w, r)
	if !ok {
		return
	}
	user, _ := currentUser(r)
	result, err := h.settings.UpdateAll(r.Context(), r.URL.Query(), body, user.ID)
	if err != nil {
		writeCephError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	body, ok := settingUpdateBody(w, r)
	if !ok {
		return
	}
	user, _ := currentUser(r)
	result, err := h.settings.Update(r.Context(), r.PathValue("name"), body["value"], user.ID)
	if err != nil {
		writeCephError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) ResetSetting(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	result, err := h.settings.Reset(r.Context(), r.PathValue("name"))
	if err != nil {
		writeCephError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) ListSettingGroups(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	result, err := h.settings.ListGroups(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetSettingGroup(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	result, err := h.settings.GetGroup(r.Context(), cephSettingGroupForRequest(r))
	if errors.Is(err, settingsservice.ErrGroupNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeCephError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateSettingGroup(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	updates, ok := settingGroupUpdateBody(w, r)
	if !ok {
		return
	}
	user, _ := currentUser(r)
	result, err := h.settings.UpdateGroup(r.Context(), cephSettingGroupForRequest(r), updates, user.ID)
	switch {
	case errors.Is(err, settingsservice.ErrGroupNotFound):
		writeError(w, http.StatusNotFound, err)
		return
	case errors.Is(err, settingsservice.ErrInvalidGroup):
		writeError(w, http.StatusBadRequest, err)
		return
	case err != nil:
		writeCephError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func settingUpdateBody(w http.ResponseWriter, r *http.Request) (map[string]any, bool) {
	var body map[string]any
	if !decodeJSON(w, r, &body) {
		return nil, false
	}
	if _, ok := body["value"]; ok {
		return body, true
	}
	if len(body) == 1 {
		for _, value := range body {
			return map[string]any{"value": value}, true
		}
	}
	writeError(w, http.StatusBadRequest, errors.New("setting update requires a value"))
	return nil, false
}

func settingGroupUpdateBody(w http.ResponseWriter, r *http.Request) (map[string]any, bool) {
	var body map[string]any
	if !decodeJSON(w, r, &body) {
		return nil, false
	}
	if nested, ok := body["settings"].(map[string]any); ok {
		return nested, true
	}
	if nested, ok := body["values"].(map[string]any); ok {
		return nested, true
	}
	return body, true
}

func cephSettingGroupForRequest(r *http.Request) string {
	if group := settingsservice.NormalizeGroup(r.PathValue("group")); group != "" {
		return group
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1"), "/"), "/")
	if len(parts) >= 2 {
		return settingsservice.NormalizeGroup(parts[0])
	}
	return ""
}

// Generic Dashboard proxy support remains in the transport layer because it
// only maps HTTP paths and methods. Sensitive-value policy is owned by settings.
func (h *Handler) proxyTemplatePath(w http.ResponseWriter, r *http.Request, cephPath string, redact bool) {
	path := cephPath
	for _, name := range pathParameterNames(path) {
		path = strings.ReplaceAll(path, "{"+name+"}", url.PathEscape(r.PathValue(name)))
	}
	h.proxyRawPath(w, r, path, redact)
}

func (h *Handler) proxyRawPath(w http.ResponseWriter, r *http.Request, path string, redact bool) {
	body, ok := rawRequestBody(w, r)
	if !ok {
		return
	}
	payload, err := h.ceph.Raw(r.Context(), r.Method, path, r.URL.Query(), body)
	if err != nil {
		writeCephError(w, err)
		return
	}
	if redact {
		writeJSON(w, http.StatusOK, settingsservice.RedactAny(payload))
		return
	}
	writeRawJSON(w, http.StatusOK, payload)
}

func requireDashboardAccess(w http.ResponseWriter, r *http.Request) bool {
	if isDashboardAdminPath(r.URL.Path) {
		return requireAdmin(w, r)
	}
	if r.Method == http.MethodGet {
		if _, ok := currentUser(r); !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
			return false
		}
		return true
	}
	return requireAdmin(w, r)
}

func isDashboardAdminPath(path string) bool {
	normalized := strings.ToLower(path)
	for _, marker := range []string{"/settings", "/grafana/settings", "/prometheus/settings", "/rgw/settings", "/iscsi/settings", "/nfs/settings", "/dashboard/settings", "/dashboard/user", "/dashboard/role", "/dashboard/auth", "/dashboard/feedback/api_key", "/auth", "/dashboard-user", "/dashboard-role", "/feedback/api-key"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func redactAnyJSON(payload json.RawMessage) any { return settingsservice.RedactAny(payload) }
func redactDashboardPayload(path, method string) bool {
	if method != http.MethodGet {
		return true
	}
	path = strings.ToLower(path)
	return strings.Contains(path, "settings") || strings.Contains(path, "user") || strings.Contains(path, "role") || strings.Contains(path, "feedback") || strings.Contains(path, "auth") || strings.Contains(path, "key")
}
