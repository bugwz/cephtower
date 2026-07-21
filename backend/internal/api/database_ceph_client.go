package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"

	"cephtower/backend/internal/integrations/ceph"
	"cephtower/backend/internal/integrations/ceph/dashboard"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

type databaseCephClient struct {
	database func() *gorm.DB
}

func newDatabaseCephClient(database func() *gorm.DB) *databaseCephClient {
	return &databaseCephClient{database: database}
}

func (c *databaseCephClient) dashboardClient(ctx context.Context) (*dashboard.DashboardClient, error) {
	var cluster store.CephCluster
	err := c.database().
		WithContext(ctx).
		Where("enabled = ? AND dashboard_enabled = ?", true, true).
		Order("id asc").
		First(&cluster).
		Error
	if err != nil {
		return nil, err
	}

	return dashboard.NewDashboardClient(dashboard.Config{
		BaseURL:     cluster.DashboardBaseURL,
		Username:    cluster.DashboardUsername,
		Password:    cluster.DashboardPassword,
		InsecureTLS: cluster.DashboardInsecureTLS,
	}), nil
}

func (c *databaseCephClient) ClusterSummary(ctx context.Context) (ceph.ClusterSummary, error) {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ceph.ClusterSummary{HealthStatus: "unknown"}, nil
		}
		return ceph.ClusterSummary{}, err
	}
	return client.ClusterSummary(ctx)
}

func (c *databaseCephClient) Version(ctx context.Context) (string, error) {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return "", err
	}
	return client.Version(ctx)
}

func (c *databaseCephClient) HealthFull(ctx context.Context) (map[string]any, error) {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.HealthFull(ctx)
}

func (c *databaseCephClient) HealthMinimal(ctx context.Context) (map[string]any, error) {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.HealthMinimal(ctx)
}

func (c *databaseCephClient) ListHosts(ctx context.Context, options ceph.ListHostsOptions) ([]ceph.Host, error) {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.ListHosts(ctx, options)
}

func (c *databaseCephClient) HostDetails(ctx context.Context, hostname string) (map[string]any, error) {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.HostDetails(ctx, hostname)
}

func (c *databaseCephClient) CreateHost(ctx context.Context, request ceph.HostRequest) error {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return err
	}
	return client.CreateHost(ctx, request)
}

func (c *databaseCephClient) UpdateHost(ctx context.Context, hostname string, request ceph.UpdateHostRequest) error {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return err
	}
	return client.UpdateHost(ctx, hostname, request)
}

func (c *databaseCephClient) DeleteHost(ctx context.Context, hostname string) error {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return err
	}
	return client.DeleteHost(ctx, hostname)
}

func (c *databaseCephClient) HostDaemons(ctx context.Context, hostname string) ([]map[string]any, error) {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.HostDaemons(ctx, hostname)
}

func (c *databaseCephClient) HostDevices(ctx context.Context, hostname string) ([]map[string]any, error) {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.HostDevices(ctx, hostname)
}

func (c *databaseCephClient) HostInventory(ctx context.Context, hostname string) (map[string]any, error) {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.HostInventory(ctx, hostname)
}

func (c *databaseCephClient) ListOSDs(ctx context.Context, options ceph.ListOSDsOptions) ([]map[string]any, error) {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.ListOSDs(ctx, options)
}

func (c *databaseCephClient) GetOSD(ctx context.Context, serviceID string) (map[string]any, error) {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.GetOSD(ctx, serviceID)
}

func (c *databaseCephClient) OSDFlags(ctx context.Context) ([]string, error) {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.OSDFlags(ctx)
}

func (c *databaseCephClient) ListDaemons(ctx context.Context, daemonTypes string) ([]map[string]any, error) {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.ListDaemons(ctx, daemonTypes)
}

func (c *databaseCephClient) ApplyDaemonAction(ctx context.Context, daemonName string, request ceph.DaemonActionRequest) error {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return err
	}
	return client.ApplyDaemonAction(ctx, daemonName, request)
}

func (c *databaseCephClient) Raw(ctx context.Context, method string, path string, query url.Values, body any) (json.RawMessage, error) {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Raw(ctx, method, path, query, body)
}
