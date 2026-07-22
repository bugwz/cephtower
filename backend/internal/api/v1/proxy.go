package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

func (api *API) ProxyCephPath(w http.ResponseWriter, r *http.Request) {
	routePath := strings.TrimPrefix(r.Pattern, r.Method+" "+PathPrefix)
	cephPath, ok := dashboardProxyPaths[routePath]
	if !ok {
		writeError(w, http.StatusNotFound, errors.New("not found"))
		return
	}

	api.ProxyCeph(r.Method, cephPath)(w, r)
}

func (api *API) ProxyCeph(method string, cephPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := cephPath
		for _, name := range pathParameterNames(path) {
			path = strings.ReplaceAll(path, "{"+name+"}", url.PathEscape(r.PathValue(name)))
		}

		body, ok := rawRequestBody(w, r)
		if !ok {
			return
		}

		payload, err := api.ceph.Raw(r.Context(), method, path, r.URL.Query(), body)
		if err != nil {
			writeCephError(w, err)
			return
		}

		writeRawJSON(w, http.StatusOK, payload)
	}
}

func pathParameterNames(path string) []string {
	var names []string
	for {
		start := strings.Index(path, "{")
		if start < 0 {
			return names
		}
		rest := path[start+1:]
		end := strings.Index(rest, "}")
		if end < 0 {
			return names
		}
		names = append(names, rest[:end])
		path = rest[end+1:]
	}
}

var dashboardProxyPaths = map[string]string{
	"/osd/{id}/device":    "/api/osd/{id}/devices",
	"/osd/{id}/histogram": "/api/osd/{id}/histogram",
	"/osd/{id}/mark":      "/api/osd/{id}/mark",
	"/osd/{id}/reweight":  "/api/osd/{id}/reweight",
	"/osd/{id}/scrub":     "/api/osd/{id}/scrub",
	"/osd/{id}/smart":     "/api/osd/{id}/smart",

	"/monitor":                                          "/api/monitor",
	"/mgr/module":                                       "/api/mgr/module",
	"/mgr/module/{name}/enable":                         "/api/mgr/module/{name}/enable",
	"/mgr/module/{name}/disable":                        "/api/mgr/module/{name}/disable",
	"/service":                                          "/api/service",
	"/service/known-type":                               "/api/service/known_types",
	"/service/{name}":                                   "/api/service/{name}",
	"/service/{name}/daemon":                            "/api/service/{name}/daemons",
	"/pool":                                             "/api/pool",
	"/pool/{name}":                                      "/api/pool/{name}",
	"/pool/{name}/configuration":                        "/api/pool/{name}/configuration",
	"/block/image":                                      "/api/block/image",
	"/block/image/default-feature":                      "/api/block/image/default_features",
	"/block/image/clone-format-version":                 "/api/block/image/clone_format_version",
	"/block/image/trash":                                "/api/block/image/trash",
	"/block/image/trash/purge":                          "/api/block/image/trash/purge",
	"/block/image/trash/{image}":                        "/api/block/image/trash/{image}",
	"/block/image/trash/{image}/restore":                "/api/block/image/trash/{image}/restore",
	"/block/image/{image}":                              "/api/block/image/{image}",
	"/block/image/{image}/copy":                         "/api/block/image/{image}/copy",
	"/block/image/{image}/flatten":                      "/api/block/image/{image}/flatten",
	"/block/image/{image}/snapshot":                     "/api/block/image/{image}/snap",
	"/block/image/{image}/snapshot/{snapshot}":          "/api/block/image/{image}/snap/{snapshot}",
	"/block/image/{image}/snapshot/{snapshot}/clone":    "/api/block/image/{image}/snap/{snapshot}/clone",
	"/block/image/{image}/snapshot/{snapshot}/rollback": "/api/block/image/{image}/snap/{snapshot}/rollback",
	"/block/mirroring/summary":                          "/api/block/mirroring/summary",

	"/filesystem":                "/api/cephfs",
	"/filesystem/{id}":           "/api/cephfs/{id}",
	"/filesystem/{id}/client":    "/api/cephfs/{id}/clients",
	"/filesystem/{id}/root":      "/api/cephfs/{id}/get_root_directory",
	"/filesystem/{id}/directory": "/api/cephfs/{id}/ls_dir",
	"/filesystem/{id}/quota":     "/api/cephfs/{id}/quota",
	"/filesystem/{id}/statfs":    "/api/cephfs/{id}/statfs",
	"/object/gateway":            "/api/rgw/daemon",
	"/object/gateway/{id}":       "/api/rgw/daemon/{id}",
	"/object/user":               "/api/rgw/user",
	"/object/user/{uid}":         "/api/rgw/user/{uid}",
	"/object/bucket":             "/api/rgw/bucket",
	"/object/bucket/{bucket}":    "/api/rgw/bucket/{bucket}",
	"/object/account":            "/api/rgw/accounts",
	"/object/account/{id}":       "/api/rgw/accounts/{id}",
	"/configuration":             "/api/cluster_conf",
	"/configuration/filter":      "/api/cluster_conf/filter",
	"/configuration/{name}":      "/api/cluster_conf/{name}",
	"/log":                       "/api/logs/all",
}

func rawRequestBody(w http.ResponseWriter, r *http.Request) (any, bool) {
	if r.Body == nil || r.ContentLength == 0 {
		return nil, true
	}

	var body json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return nil, false
	}
	if len(body) == 0 {
		return nil, true
	}

	return body, true
}
