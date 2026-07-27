package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type APIErrorData struct {
	ErrorCode string         `json:"error_code"`
	Retryable bool           `json:"retryable"`
	Details   map[string]any `json:"details,omitempty"`
	RequestID string         `json:"request_id"`
}

func WriteSuccess(w http.ResponseWriter, status int, message string, data any) {
	if message == "" {
		message = "success"
	}
	write(w, status, APIResponse{Code: 0, Message: message, Data: data})
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, errorCode, message string, retryable bool, details map[string]any) {
	if message == "" {
		message = http.StatusText(status)
	}
	write(w, status, APIResponse{Code: status, Message: message, Data: APIErrorData{ErrorCode: errorCode, Retryable: retryable, Details: details, RequestID: RequestID(r)}})
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	if status >= 400 {
		WriteError(w, &http.Request{}, status, "http_error", fmt.Sprint(payload), false, nil)
		return
	}
	WriteSuccess(w, status, "success", payload)
}

func write(w http.ResponseWriter, status int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

func DecodeStrict(w http.ResponseWriter, r *http.Request, out any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", err.Error(), false, nil)
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", "request body must contain one JSON value", false, nil)
		return false
	}
	return true
}
