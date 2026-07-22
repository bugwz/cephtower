package v1

import (
	"net/http"

	"cephtower/backend/internal/integrations/ceph"
)

func (api *API) ListOSDs(w http.ResponseWriter, r *http.Request) {
	osds, err := api.ceph.ListOSDs(r.Context(), ceph.ListOSDsOptions{
		Offset: intQuery(r.URL.Query(), "offset"),
		Limit:  intQuery(r.URL.Query(), "limit"),
		Search: r.URL.Query().Get("search"),
		Sort:   r.URL.Query().Get("sort"),
	})
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, osds)
}

func (api *API) OSDDetails(w http.ResponseWriter, r *http.Request) {
	payload, err := api.ceph.GetOSD(r.Context(), r.PathValue("id"))
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (api *API) OSDFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := api.ceph.OSDFlags(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string][]string{"flags": flags})
}
