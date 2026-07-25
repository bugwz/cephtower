package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cephtower/backend/internal/api/v1/handler"
	authservice "cephtower/backend/internal/service/auth"
	"cephtower/backend/internal/store"
)

func TestServerRoutesVersionedAPI(t *testing.T) {
	apiHandler := handler.New(nil, handler.Dependencies{AuthService: serverAuthStub{}})
	routes := NewServer(apiHandler).Routes()

	t.Run("public health route", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
		recorder := httptest.NewRecorder()

		routes.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
		}
		if origin := recorder.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Fatalf("CORS origin = %q, want *", origin)
		}
	})

	t.Run("unknown API route uses versioned envelope", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/missing", nil)
		request.Header.Set("Authorization", "Bearer test-token")
		recorder := httptest.NewRecorder()

		routes.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
		}
		var response struct {
			Code int `json:"code"`
		}
		if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response.Code != http.StatusNotFound {
			t.Fatalf("response code = %d, want %d", response.Code, http.StatusNotFound)
		}
	})
}

type serverAuthStub struct{}

func (serverAuthStub) Login(context.Context, string, string) (authservice.LoginResult, error) {
	return authservice.LoginResult{}, nil
}

func (serverAuthStub) UserForToken(context.Context, string) (store.User, error) {
	return store.User{Role: store.UserRoleAdmin, Enabled: true}, nil
}

func (serverAuthStub) ListUsers(context.Context) ([]store.User, error) {
	return nil, nil
}

func (serverAuthStub) CreateUser(context.Context, authservice.CreateUserInput) (store.User, error) {
	return store.User{}, nil
}

func (serverAuthStub) UpdateUser(context.Context, uint, authservice.UpdateUserInput) (store.User, error) {
	return store.User{}, nil
}

func (serverAuthStub) RequestPasswordReset(context.Context, string) error {
	return nil
}

func (serverAuthStub) ConfirmPasswordReset(context.Context, string, string, string) error {
	return nil
}
