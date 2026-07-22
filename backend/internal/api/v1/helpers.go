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
	w.WriteHeader(httpStatusForAPIResponse(status))
	_ = json.NewEncoder(w).Encode(apiResponseForStatus(status, payload))
}

func writeRawJSON(w http.ResponseWriter, status int, payload json.RawMessage) {
	if len(payload) == 0 {
		writeJSON(w, status, nil)
		return
	}

	var data any
	if err := json.Unmarshal(payload, &data); err != nil {
		data = payload
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusForAPIResponse(status))
	_ = json.NewEncoder(w).Encode(apiResponseForStatus(status, data))
}

func httpStatusForAPIResponse(status int) int {
	if status == http.StatusUnauthorized || status == http.StatusForbidden || status >= http.StatusInternalServerError {
		return status
	}
	return http.StatusOK
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

type apiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func apiResponseForStatus(status int, payload any) apiResponse {
	if status >= http.StatusBadRequest {
		return apiResponse{
			Code:    status,
			Message: responseMessage(payload, http.StatusText(status)),
			Data:    nil,
		}
	}

	return apiResponse{
		Code:    0,
		Message: responseMessage(payload, "success"),
		Data:    payload,
	}
}

func responseMessage(payload any, fallback string) string {
	if values, ok := payload.(map[string]string); ok {
		for _, key := range []string{"message", "error"} {
			if message := values[key]; message != "" {
				return message
			}
		}
	}
	if values, ok := payload.(map[string]any); ok {
		for _, key := range []string{"message", "error"} {
			if message, ok := values[key].(string); ok && message != "" {
				return message
			}
		}
	}
	return fallback
}
