package v1

import (
	"context"
	"encoding/json"
	"net/url"

	"cephtower/backend/internal/config"
	cephapi "cephtower/backend/internal/integrations/ceph"
	cephservice "cephtower/backend/internal/service/ceph"
	"gorm.io/gorm"
)

const PathPrefix = "/api/v1"

type CephClient interface {
	ClusterSummary(ctx context.Context) (cephapi.ClusterSummary, error)
	Version(ctx context.Context) (string, error)
	HealthFull(ctx context.Context) (map[string]any, error)
	HealthMinimal(ctx context.Context) (map[string]any, error)
	ListHosts(ctx context.Context, options cephapi.ListHostsOptions) ([]cephapi.Host, error)
	HostDetails(ctx context.Context, hostname string) (map[string]any, error)
	CreateHost(ctx context.Context, request cephapi.HostRequest) error
	UpdateHost(ctx context.Context, hostname string, request cephapi.UpdateHostRequest) error
	DeleteHost(ctx context.Context, hostname string) error
	HostDaemons(ctx context.Context, hostname string) ([]map[string]any, error)
	HostDevices(ctx context.Context, hostname string) ([]map[string]any, error)
	HostInventory(ctx context.Context, hostname string) (map[string]any, error)
	ListOSDs(ctx context.Context, options cephapi.ListOSDsOptions) ([]map[string]any, error)
	GetOSD(ctx context.Context, serviceID string) (map[string]any, error)
	OSDFlags(ctx context.Context) ([]string, error)
	ListDaemons(ctx context.Context, daemonTypes string) ([]map[string]any, error)
	ApplyDaemonAction(ctx context.Context, daemonName string, request cephapi.DaemonActionRequest) error
	Raw(ctx context.Context, method string, path string, query url.Values, body any) (json.RawMessage, error)
}

type API struct {
	ceph              CephClient
	currentConfig     func() config.Config
	database          func() *gorm.DB
	replaceDatabase   func(config.Config, *gorm.DB) *gorm.DB
	clusterDiscoverer cephservice.ClusterDiscoverer
}

func NewAPI(cephClient CephClient, deps Dependencies) *API {
	return &API{
		ceph:              cephClient,
		currentConfig:     deps.CurrentConfig,
		database:          deps.Database,
		replaceDatabase:   deps.ReplaceDatabase,
		clusterDiscoverer: deps.ClusterDiscoverer,
	}
}

type Dependencies struct {
	CurrentConfig     func() config.Config
	Database          func() *gorm.DB
	ReplaceDatabase   func(config.Config, *gorm.DB) *gorm.DB
	ClusterDiscoverer cephservice.ClusterDiscoverer
}
