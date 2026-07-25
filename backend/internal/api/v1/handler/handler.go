package handler

import (
	"context"
	"encoding/json"

	authservice "cephtower/backend/internal/service/auth"
	cephproxyservice "cephtower/backend/internal/service/cephproxy"
	clusterservice "cephtower/backend/internal/service/cluster"
	settingsservice "cephtower/backend/internal/service/settings"
	setupservice "cephtower/backend/internal/service/setup"
	"cephtower/backend/internal/store"
)

type CephClient = cephproxyservice.Client

type Handler struct {
	ceph           CephClient
	clusters       *clusterservice.Service
	auth           AuthService
	settings       *settingsservice.Service
	systemSettings *settingsservice.SystemService
	setup          *setupservice.Service
}

func New(cephClient CephClient, deps Dependencies) *Handler {
	settingsService := deps.SettingsService
	if settingsService == nil && cephClient != nil {
		settingsService = settingsservice.New(cephClient, nil)
	}
	return &Handler{
		ceph: cephClient, clusters: deps.ClusterService, auth: deps.AuthService,
		settings: settingsService, systemSettings: deps.SystemSettingsService,
		setup: deps.SetupService,
	}
}

type Dependencies struct {
	ClusterService        *clusterservice.Service
	AuthService           AuthService
	SettingsService       *settingsservice.Service
	SystemSettingsService *settingsservice.SystemService
	SetupService          *setupservice.Service
}

type AuthService interface {
	Login(context.Context, string, string) (authservice.LoginResult, error)
	UserForToken(context.Context, string) (store.User, error)
	ListUsers(context.Context) ([]store.User, error)
	CreateUser(context.Context, authservice.CreateUserInput) (store.User, error)
	UpdateUser(context.Context, uint, authservice.UpdateUserInput) (store.User, error)
	RequestPasswordReset(context.Context, string) error
	ConfirmPasswordReset(context.Context, string, string, string) error
}

type TaskSubmitter func(string, func(context.Context) error) error

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type RawJSONRequest = json.RawMessage
type RawJSONResponse = json.RawMessage
