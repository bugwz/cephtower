package handler

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
)

func requiredUintBody(w http.ResponseWriter, r *http.Request, body map[string]any, name string) (uint64, bool) {
	value, exists := body[name]
	if !exists {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", name+" is required", false, nil)
		return 0, false
	}
	id, err := uintFromJSON(value)
	if err != nil || id == 0 {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", name+" must be a positive integer", false, nil)
		return 0, false
	}
	return id, true
}

func requiredStringBody(w http.ResponseWriter, r *http.Request, body map[string]any, name string) (string, bool) {
	value, exists := body[name]
	if !exists {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", name+" is required", false, nil)
		return "", false
	}
	text, ok := value.(string)
	text = strings.TrimSpace(text)
	if !ok || text == "" {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", name+" must be a non-empty string", false, nil)
		return "", false
	}
	return text, true
}

func optionalStringBody(body map[string]any, names ...string) string {
	for _, name := range names {
		value, ok := body[name].(string)
		if ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (h *Handler) scopedBody(w http.ResponseWriter, r *http.Request) (map[string]any, uint64, bool) {
	var body map[string]any
	if !DecodeStrict(w, r, &body) {
		return nil, 0, false
	}
	id, ok := requiredUintBody(w, r, body, "cluster_id")
	if !ok {
		return nil, 0, false
	}
	if _, err := h.Clusters.Get(r.Context(), id); err != nil {
		clusterError(w, r, err)
		return nil, 0, false
	}
	return body, id, true
}

func uintFromJSON(value any) (uint64, error) {
	switch typed := value.(type) {
	case float64:
		if math.Trunc(typed) != typed || typed <= 0 {
			return 0, fmt.Errorf("not a positive integer")
		}
		return uint64(typed), nil
	case int:
		if typed <= 0 {
			return 0, fmt.Errorf("not a positive integer")
		}
		return uint64(typed), nil
	case uint64:
		if typed == 0 {
			return 0, fmt.Errorf("not a positive integer")
		}
		return typed, nil
	case string:
		return strconv.ParseUint(strings.TrimSpace(typed), 10, 64)
	default:
		return 0, fmt.Errorf("not a positive integer")
	}
}
