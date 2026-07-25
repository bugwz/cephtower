package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	cephproxyservice "cephtower/backend/internal/service/cephproxy"
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

// WriteJSON writes the shared API response envelope for server middleware.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	writeJSON(w, status, payload)
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
	return status
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, errorResponse{Error: err.Error()})
}

func writeCephError(w http.ResponseWriter, err error) {
	if status, ok := cephproxyservice.ErrorStatus(err); ok {
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

type messageResponse struct {
	Message string `json:"message"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// dashboardRequest is the normalized input passed to Dashboard-backed
// handlers. Dashboard APIs intentionally retain their JSON body because the
// payload schema is owned by the connected Ceph version.
type dashboardRequest struct {
	PathParameters dashboardPathParameters
	Query          url.Values
	Body           json.RawMessage
}

type dashboardPathParameters map[string]string

// dashboardResponse preserves Dashboard's version-dependent JSON payload
// within CephTower's common API response envelope.
type dashboardResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func writeDashboardJSON(w http.ResponseWriter, status int, payload json.RawMessage) {
	if len(payload) == 0 {
		payload = json.RawMessage("null")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusForAPIResponse(status))
	_ = json.NewEncoder(w).Encode(dashboardResponse{
		Code:    0,
		Message: "success",
		Data:    payload,
	})
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
	if response, ok := payload.(messageResponse); ok && response.Message != "" {
		return response.Message
	}
	if response, ok := payload.(errorResponse); ok && response.Error != "" {
		return response.Error
	}
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
	if action, ok := payload.(MessageResponse); ok && action.Message != "" {
		return action.Message
	}
	return fallback
}
