package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"cephtower/backend/internal/integrations/ceph/dashboard"
)

func intQuery(query url.Values, name string) *int {
	value := query.Get(name)
	if value == "" {
		return nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil
	}

	return &parsed
}

func boolQuery(query url.Values, name string) *bool {
	value := query.Get(name)
	if value == "" {
		return nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil
	}

	return &parsed
}

func decodeRequestJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return false
	}

	return true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeRawJSON(w http.ResponseWriter, status int, payload json.RawMessage) {
	if len(payload) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}

func writeCephError(w http.ResponseWriter, err error) {
	var apiErr *dashboard.APIError
	if errors.As(err, &apiErr) {
		status := apiErr.StatusCode
		if status == 0 {
			status = http.StatusBadGateway
		}
		writeError(w, status, err)
		return
	}

	writeError(w, http.StatusBadGateway, err)
}
