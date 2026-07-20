package v1

import (
	"context"
	"encoding/json"
	"net/url"

	"cephtower/backend/internal/integrations/ceph"
)

type CephClient interface {
	ClusterSummary(ctx context.Context) (ceph.ClusterSummary, error)
	Version(ctx context.Context) (string, error)
	HealthFull(ctx context.Context) (map[string]any, error)
	HealthMinimal(ctx context.Context) (map[string]any, error)
	ListHosts(ctx context.Context, options ceph.ListHostsOptions) ([]ceph.Host, error)
	HostDetails(ctx context.Context, hostname string) (map[string]any, error)
	CreateHost(ctx context.Context, request ceph.HostRequest) error
	UpdateHost(ctx context.Context, hostname string, request ceph.UpdateHostRequest) error
	DeleteHost(ctx context.Context, hostname string) error
	HostDaemons(ctx context.Context, hostname string) ([]map[string]any, error)
	HostDevices(ctx context.Context, hostname string) ([]map[string]any, error)
	HostInventory(ctx context.Context, hostname string) (map[string]any, error)
	ListOSDs(ctx context.Context, options ceph.ListOSDsOptions) ([]map[string]any, error)
	GetOSD(ctx context.Context, serviceID string) (map[string]any, error)
	OSDFlags(ctx context.Context) ([]string, error)
	ListDaemons(ctx context.Context, daemonTypes string) ([]map[string]any, error)
	ApplyDaemonAction(ctx context.Context, daemonName string, request ceph.DaemonActionRequest) error
	Raw(ctx context.Context, method string, path string, query url.Values, body any) (json.RawMessage, error)
}

type API struct {
	ceph CephClient
}
