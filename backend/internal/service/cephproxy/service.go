package cephproxy

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"

	ceph "cephtower/backend/internal/integration/ceph"
	"cephtower/backend/internal/integration/ceph/dashboard"
	"cephtower/backend/internal/service/collector"
	"cephtower/backend/internal/store"
)

type ListHostsOptions = ceph.ListHostsOptions
type Host = ceph.Host
type HostRequest = ceph.HostRequest
type UpdateHostRequest = ceph.UpdateHostRequest
type ListOSDsOptions = ceph.ListOSDsOptions
type DaemonActionRequest = ceph.DaemonActionRequest
type ClusterSummary = ceph.ClusterSummary

type Client interface {
	ClusterSummary(context.Context) (ceph.ClusterSummary, error)
	Version(context.Context) (string, error)
	HealthFull(context.Context) (map[string]any, error)
	HealthMinimal(context.Context) (map[string]any, error)
	ListHosts(context.Context, ceph.ListHostsOptions) ([]ceph.Host, error)
	HostDetails(context.Context, string) (map[string]any, error)
	CreateHost(context.Context, ceph.HostRequest) error
	UpdateHost(context.Context, string, ceph.UpdateHostRequest) error
	DeleteHost(context.Context, string) error
	HostDaemons(context.Context, string) ([]map[string]any, error)
	HostDevices(context.Context, string) ([]map[string]any, error)
	HostInventory(context.Context, string) (map[string]any, error)
	ListOSDs(context.Context, ceph.ListOSDsOptions) ([]map[string]any, error)
	GetOSD(context.Context, string) (map[string]any, error)
	OSDFlags(context.Context) ([]string, error)
	ListDaemons(context.Context, string) ([]map[string]any, error)
	ApplyDaemonAction(context.Context, string, ceph.DaemonActionRequest) error
	Raw(context.Context, string, string, url.Values, any) (json.RawMessage, error)
}

type Service struct{ client Client }

func ErrorStatus(err error) (int, bool) {
	var apiErr *dashboard.APIError
	if !errors.As(err, &apiErr) {
		return 0, false
	}
	return apiErr.StatusCode, true
}

func New(database func() *store.Database, submit func(string, func(context.Context) error) error, worker *collector.Service, workDir string) *Service {
	return &Service{client: collector.NewDatabaseCephClientWithSubmitter(database, submit, workDir).UseCollector(worker)}
}

func (s *Service) ClusterSummary(ctx context.Context) (ceph.ClusterSummary, error) {
	return s.client.ClusterSummary(ctx)
}
func (s *Service) Version(ctx context.Context) (string, error) { return s.client.Version(ctx) }
func (s *Service) HealthFull(ctx context.Context) (map[string]any, error) {
	return s.client.HealthFull(ctx)
}
func (s *Service) HealthMinimal(ctx context.Context) (map[string]any, error) {
	return s.client.HealthMinimal(ctx)
}
func (s *Service) ListHosts(ctx context.Context, options ceph.ListHostsOptions) ([]ceph.Host, error) {
	return s.client.ListHosts(ctx, options)
}
func (s *Service) HostDetails(ctx context.Context, hostname string) (map[string]any, error) {
	return s.client.HostDetails(ctx, hostname)
}
func (s *Service) CreateHost(ctx context.Context, request ceph.HostRequest) error {
	return s.client.CreateHost(ctx, request)
}
func (s *Service) UpdateHost(ctx context.Context, hostname string, request ceph.UpdateHostRequest) error {
	return s.client.UpdateHost(ctx, hostname, request)
}
func (s *Service) DeleteHost(ctx context.Context, hostname string) error {
	return s.client.DeleteHost(ctx, hostname)
}
func (s *Service) HostDaemons(ctx context.Context, hostname string) ([]map[string]any, error) {
	return s.client.HostDaemons(ctx, hostname)
}
func (s *Service) HostDevices(ctx context.Context, hostname string) ([]map[string]any, error) {
	return s.client.HostDevices(ctx, hostname)
}
func (s *Service) HostInventory(ctx context.Context, hostname string) (map[string]any, error) {
	return s.client.HostInventory(ctx, hostname)
}
func (s *Service) ListOSDs(ctx context.Context, options ceph.ListOSDsOptions) ([]map[string]any, error) {
	return s.client.ListOSDs(ctx, options)
}
func (s *Service) GetOSD(ctx context.Context, id string) (map[string]any, error) {
	return s.client.GetOSD(ctx, id)
}
func (s *Service) OSDFlags(ctx context.Context) ([]string, error) { return s.client.OSDFlags(ctx) }
func (s *Service) ListDaemons(ctx context.Context, types string) ([]map[string]any, error) {
	return s.client.ListDaemons(ctx, types)
}
func (s *Service) ApplyDaemonAction(ctx context.Context, name string, request ceph.DaemonActionRequest) error {
	return s.client.ApplyDaemonAction(ctx, name, request)
}
func (s *Service) Raw(ctx context.Context, method, path string, query url.Values, body any) (json.RawMessage, error) {
	return s.client.Raw(ctx, method, path, query, body)
}
