package handler

import (
	"context"
	"net/http"

	authservice "cephtower/backend/internal/service/auth"
	clusterservice "cephtower/backend/internal/service/cluster"
	endpointservice "cephtower/backend/internal/service/endpoint"
	externalservice "cephtower/backend/internal/service/external"
	hostprofileservice "cephtower/backend/internal/service/hostprofile"
	mutationservice "cephtower/backend/internal/service/mutation"
	reconcilerservice "cephtower/backend/internal/service/reconciler"
	setupservice "cephtower/backend/internal/service/setup"
	"cephtower/backend/internal/store"
)

type Handler struct {
	Auth         *authservice.Service
	Clusters     *clusterservice.Service
	Endpoints    *endpointservice.Service
	External     *externalservice.Service
	HostProfiles *hostprofileservice.Service
	Mutations    *mutationservice.Service
	Reconciler   *reconcilerservice.Service
	Setup        *setupservice.Service
	Database     func() *store.Database
	AuthEnabled  func() bool
}

type Dependencies struct {
	Auth         *authservice.Service
	Clusters     *clusterservice.Service
	Endpoints    *endpointservice.Service
	External     *externalservice.Service
	HostProfiles *hostprofileservice.Service
	Mutations    *mutationservice.Service
	Reconciler   *reconcilerservice.Service
	Setup        *setupservice.Service
	Database     func() *store.Database
	AuthEnabled  func() bool
}

func New(deps Dependencies) *Handler {
	return &Handler{Auth: deps.Auth, Clusters: deps.Clusters, Endpoints: deps.Endpoints, External: deps.External, HostProfiles: deps.HostProfiles, Mutations: deps.Mutations, Reconciler: deps.Reconciler, Setup: deps.Setup, Database: deps.Database, AuthEnabled: deps.AuthEnabled}
}

func (h *Handler) RequireAuth() bool {
	return h.AuthEnabled == nil || h.AuthEnabled()
}

type userContextKey struct{}

type clusterContextKey struct{}

type requestIDContextKey struct{}

func WithUser(ctx context.Context, user store.User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

func DefaultAdminUser() store.User {
	return store.User{Username: "admin", DisplayName: "管理员", Status: "active"}
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
