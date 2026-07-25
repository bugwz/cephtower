package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLivenessDoesNotDependOnInitialization(t *testing.T) {
	handler := New(nil, Dependencies{})
	live := httptest.NewRecorder()
	handler.Healthz(live, httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil))
	if live.Code != http.StatusOK {
		t.Fatalf("Healthz() status = %d", live.Code)
	}
	ready := httptest.NewRecorder()
	handler.Readyz(ready, httptest.NewRequest(http.MethodGet, "/api/v1/readyz", nil))
	if ready.Code != http.StatusServiceUnavailable {
		t.Fatalf("Readyz() status = %d, want 503", ready.Code)
	}
}
