package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// Dashboard request proxy.
// serveDashboardPath applies  Dashboard access control and proxies a request.
func (h *Handler) serveDashboardPath(w http.ResponseWriter, r *http.Request, cephPath string) {
	if !requireDashboardAccess(w, r) {
		return
	}
	h.proxyDashboardPath(w, r, cephPath, redactDashboardPayload(cephPath, r.Method))
}

func (h *Handler) proxyDashboardPath(w http.ResponseWriter, r *http.Request, cephPath string, redact bool) {
	request, ok := dashboardRequestFromHTTP(w, r, cephPath)
	if !ok {
		return
	}

	path := cephPath
	for name, value := range request.PathParameters {
		path = strings.ReplaceAll(path, "{"+name+"}", url.PathEscape(value))
	}

	payload, err := h.ceph.Raw(r.Context(), r.Method, path, request.Query, request.Body)
	if err != nil {
		writeCephError(w, err)
		return
	}
	if redact {
		payload, err = json.Marshal(redactAnyJSON(payload))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	writeDashboardJSON(w, http.StatusOK, payload)
}

func dashboardRequestFromHTTP(w http.ResponseWriter, r *http.Request, pathTemplate string) (dashboardRequest, bool) {
	request := dashboardRequest{
		PathParameters: make(dashboardPathParameters),
		Query:          r.URL.Query(),
	}
	for _, name := range pathParameterNames(pathTemplate) {
		value := strings.TrimSpace(r.PathValue(name))
		if value == "" {
			writeError(w, http.StatusBadRequest, errors.New(name+" path parameter is required"))
			return dashboardRequest{}, false
		}
		request.PathParameters[name] = value
	}

	body, ok := rawRequestBody(w, r)
	if !ok {
		return dashboardRequest{}, false
	}
	if body == nil {
		return request, true
	}
	request.Body = body.(json.RawMessage)
	return request, true
}
