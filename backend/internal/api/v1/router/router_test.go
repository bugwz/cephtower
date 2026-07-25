package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"cephtower/backend/internal/api/v1/handler"
)

func TestRegisterV1Routes(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, handler.New(nil, handler.Dependencies{}))

	tests := []struct {
		name        string
		method      string
		path        string
		wantPattern string
	}{
		{name: "health uses version prefix", method: http.MethodGet, path: "/api/v1/healthz", wantPattern: "GET /api/v1/healthz"},
		{name: "setting uses singular path", method: http.MethodGet, path: "/api/v1/setting", wantPattern: "GET /api/v1/setting"},
		{name: "plural setting path is absent", method: http.MethodGet, path: "/api/v1/settings", wantPattern: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			_, pattern := mux.Handler(request)
			if pattern != test.wantPattern {
				t.Fatalf("pattern = %q, want %q", pattern, test.wantPattern)
			}
		})
	}
}

func TestRegisterRejectsUnsupportedMethod(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, handler.New(nil, handler.Dependencies{}))

	request := httptest.NewRequest(http.MethodPost, "/api/v1/healthz", nil)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
	if allow := recorder.Header().Get("Allow"); allow != "GET, HEAD" {
		t.Fatalf("Allow = %q, want %q", allow, "GET, HEAD")
	}
}
