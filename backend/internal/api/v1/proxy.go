package v1

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

func (api *API) proxyCephGET(cephPath string) http.HandlerFunc {
	return api.proxyCeph(http.MethodGet, cephPath)
}

func (api *API) proxyCephPOST(cephPath string) http.HandlerFunc {
	return api.proxyCeph(http.MethodPost, cephPath)
}

func (api *API) proxyCephPUT(cephPath string) http.HandlerFunc {
	return api.proxyCeph(http.MethodPut, cephPath)
}

func (api *API) proxyCephGETPath(cephPath string, pathName string) http.HandlerFunc {
	return api.proxyCephPath(http.MethodGet, cephPath, pathName)
}

func (api *API) proxyCephPath(method string, cephPath string, pathName string) http.HandlerFunc {
	return api.proxyCeph(method, renderCephPath(cephPath, map[string]string{
		pathName: "{request:" + pathName + "}",
	}))
}

func (api *API) proxyCephPath2(method string, cephPath string, firstName string, secondName string) http.HandlerFunc {
	return api.proxyCeph(method, renderCephPath(cephPath, map[string]string{
		firstName:  "{request:" + firstName + "}",
		secondName: "{request:" + secondName + "}",
	}))
}

func (api *API) proxyCeph(method string, cephPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := cephPath
		for _, name := range requestPathNames(path) {
			path = strings.ReplaceAll(path, "{request:"+name+"}", url.PathEscape(r.PathValue(name)))
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

func renderCephPath(path string, values map[string]string) string {
	for name, value := range values {
		path = strings.ReplaceAll(path, "{"+name+"}", value)
	}
	return path
}

func requestPathNames(path string) []string {
	var names []string
	for {
		start := strings.Index(path, "{request:")
		if start < 0 {
			return names
		}
		rest := path[start+len("{request:"):]
		end := strings.Index(rest, "}")
		if end < 0 {
			return names
		}
		names = append(names, rest[:end])
		path = rest[end+1:]
	}
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
