package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMethodOverrideAllowsPostBodyForGetRoutes(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /resource", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodPost, "/resource", nil)
	request.Header.Set("X-HTTP-Method-Override", http.MethodGet)
	recorder := httptest.NewRecorder()

	withMethodOverride(mux).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected overridden GET route to match, got status %d", recorder.Code)
	}
}

func TestMethodOverrideDoesNotRewriteWithoutHeader(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /resource", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodPost, "/resource", nil)
	recorder := httptest.NewRecorder()

	withMethodOverride(mux).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected POST without override to be rejected, got status %d", recorder.Code)
	}
}
