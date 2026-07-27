package handler

import (
	"context"
	"net/http"
	"net/url"

	authservice "cephtower/backend/internal/service/auth"
	clusterservice "cephtower/backend/internal/service/cluster"
	endpointservice "cephtower/backend/internal/service/endpoint"
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
)

type Handler struct {
	Auth       *authservice.Service
	Clusters   *clusterservice.Service
	Operations *operationservice.Service
	Endpoints  *endpointservice.Service
	External   ExternalReader
	Database   func() *store.Database
}

type ExternalReader interface {
	Read(context.Context, uint64, string, string, url.Values) (any, error)
}

type Dependencies struct {
	Auth       *authservice.Service
	Clusters   *clusterservice.Service
	Operations *operationservice.Service
	Endpoints  *endpointservice.Service
	External   ExternalReader
	Database   func() *store.Database
}

func New(deps Dependencies) *Handler {
	return &Handler{Auth: deps.Auth, Clusters: deps.Clusters, Operations: deps.Operations, Endpoints: deps.Endpoints, External: deps.External, Database: deps.Database}
}

type userContextKey struct{}

type clusterContextKey struct{}

type requestIDContextKey struct{}

func WithUser(ctx context.Context, user store.User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

func CurrentUser(r *http.Request) (store.User, bool) {
	user, ok := r.Context().Value(userContextKey{}).(store.User)
	return user, ok
}

func WithClusterID(ctx context.Context, id uint64) context.Context {
	return context.WithValue(ctx, clusterContextKey{}, id)
}

func ClusterID(r *http.Request) (uint64, bool) {
	id, ok := r.Context().Value(clusterContextKey{}).(uint64)
	return id, ok
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDContextKey{}, id)
}

func RequestID(r *http.Request) string {
	id, _ := r.Context().Value(requestIDContextKey{}).(string)
	return id
}
