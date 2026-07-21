package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
	var hosts []ceph.Host
	if ok, err := c.loadSnapshot(ctx, snapshotHosts, "all", &hosts); err != nil {
		return nil, err
	} else if ok {
		return filterHosts(hosts, options), nil
	}

	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.ListHosts(ctx, options)
}

func (c *databaseCephClient) HostDetails(ctx context.Context, hostname string) (map[string]any, error) {
	hosts, err := c.ListHosts(ctx, ceph.ListHostsOptions{})
	if err != nil {
		return nil, err
	}
	for _, host := range hosts {
		if host.Hostname == hostname {
			data, err := json.Marshal(host)
			if err != nil {
				return nil, err
			}
			var payload map[string]any
			if err := json.Unmarshal(data, &payload); err != nil {
				return nil, err
			}
			return payload, nil
		}
	}

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
	if err := client.CreateHost(ctx, request); err != nil {
		return err
	}
	c.syncSnapshots(ctx)
	return nil
}

func (c *databaseCephClient) UpdateHost(ctx context.Context, hostname string, request ceph.UpdateHostRequest) error {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return err
	}
	if err := client.UpdateHost(ctx, hostname, request); err != nil {
		return err
	}
	c.syncSnapshots(ctx)
	return nil
}

func (c *databaseCephClient) DeleteHost(ctx context.Context, hostname string) error {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return err
	}
	if err := client.DeleteHost(ctx, hostname); err != nil {
		return err
	}
	c.syncSnapshots(ctx)
	return nil
}

func (c *databaseCephClient) HostDaemons(ctx context.Context, hostname string) ([]map[string]any, error) {
	daemons, err := c.ListDaemons(ctx, "")
	if err != nil {
		return nil, err
	}
	var filtered []map[string]any
	for _, daemon := range daemons {
		if stringField(daemon, "hostname") == hostname {
			filtered = append(filtered, daemon)
		}
	}
	return filtered, nil
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
	var osds []map[string]any
	if ok, err := c.loadSnapshot(ctx, snapshotOSDs, "all", &osds); err != nil {
		return nil, err
	} else if ok {
		return filterRecords(osds, options.Search), nil
	}

	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.ListOSDs(ctx, options)
}

func (c *databaseCephClient) GetOSD(ctx context.Context, serviceID string) (map[string]any, error) {
	osds, err := c.ListOSDs(ctx, ceph.ListOSDsOptions{})
	if err != nil {
		return nil, err
	}
	for _, osd := range osds {
		if stringField(osd, "id") == serviceID || stringField(osd, "osd") == serviceID || stringField(osd, "service_id") == serviceID {
			return osd, nil
		}
	}

	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.GetOSD(ctx, serviceID)
}

func (c *databaseCephClient) OSDFlags(ctx context.Context) ([]string, error) {
	var flags []string
	if ok, err := c.loadSnapshot(ctx, snapshotOSDFlags, "all", &flags); err != nil {
		return nil, err
	} else if ok {
		return flags, nil
	}

	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.OSDFlags(ctx)
}

func (c *databaseCephClient) ListDaemons(ctx context.Context, daemonTypes string) ([]map[string]any, error) {
	var daemons []map[string]any
	if ok, err := c.loadSnapshot(ctx, snapshotDaemons, "all", &daemons); err != nil {
		return nil, err
	} else if ok {
		return filterDaemons(daemons, daemonTypes), nil
	}

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
	if err := client.ApplyDaemonAction(ctx, daemonName, request); err != nil {
		return err
	}
	c.syncSnapshots(ctx)
	return nil
}

func (c *databaseCephClient) Raw(ctx context.Context, method string, path string, query url.Values, body any) (json.RawMessage, error) {
	if method == http.MethodGet {
		if payload, ok, err := c.rawSnapshot(ctx, rawSnapshotCategory(path), "all"); err != nil {
			return nil, err
		} else if ok {
			return payload, nil
		}
	}

	client, err := c.dashboardClient(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := client.Raw(ctx, method, path, query, body)
	if err != nil {
		return nil, err
	}
	if method != http.MethodGet {
		c.syncSnapshots(ctx)
	}
	return payload, nil
}

func (c *databaseCephClient) loadSnapshot(ctx context.Context, category string, key string, out any) (bool, error) {
	payload, ok, err := c.rawSnapshot(ctx, category, key)
	if err != nil || !ok {
		return ok, err
	}
	if len(payload) == 0 || string(payload) == "null" {
		return true, nil
	}
	return true, json.Unmarshal(payload, out)
}

func (c *databaseCephClient) rawSnapshot(ctx context.Context, category string, key string) (json.RawMessage, bool, error) {
	if category == "" {
		return nil, false, nil
	}

	var cluster store.CephCluster
	if err := c.database().WithContext(ctx).
		Where("enabled = ? AND dashboard_enabled = ?", true, true).
		Order("id asc").
		First(&cluster).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return json.RawMessage(`null`), true, nil
		}
		return nil, false, err
	}

	var snapshot store.CephResourceSnapshot
	if err := c.database().WithContext(ctx).
		Where("cluster_id = ? AND category = ? AND resource_key = ?", cluster.ID, category, key).
		First(&snapshot).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return defaultSnapshotPayload(category), true, nil
		}
		return nil, false, err
	}
	if snapshot.LastError != "" && strings.TrimSpace(snapshot.Payload) == "null" {
		return nil, false, errors.New(snapshot.LastError)
	}
	return json.RawMessage(snapshot.Payload), true, nil
}

func defaultSnapshotPayload(category string) json.RawMessage {
	switch category {
	case snapshotHosts, snapshotOSDs, snapshotOSDFlags, snapshotDaemons, snapshotServices, snapshotMgrModules, snapshotConfiguration:
		return json.RawMessage(`[]`)
	case snapshotMonitor:
		return json.RawMessage(`{"in_quorum":[],"out_quorum":[]}`)
	default:
		return json.RawMessage(`null`)
	}
}

func (c *databaseCephClient) syncSnapshots(ctx context.Context) {
	go newCephResourceSyncer(c.database).Sync(context.WithoutCancel(ctx))
}

func rawSnapshotCategory(path string) string {
	switch path {
	case "/api/service":
		return snapshotServices
	case "/api/monitor":
		return snapshotMonitor
	case "/api/mgr/module":
		return snapshotMgrModules
	case "/api/cluster_conf":
		return snapshotConfiguration
	default:
		return ""
	}
}

func filterHosts(hosts []ceph.Host, options ceph.ListHostsOptions) []ceph.Host {
	if options.Search == "" {
		return hosts
	}
	normalized := strings.ToLower(options.Search)
	var filtered []ceph.Host
	for _, host := range hosts {
		if strings.Contains(strings.ToLower(host.Hostname+" "+host.Addr), normalized) {
			filtered = append(filtered, host)
		}
	}
	return filtered
}

func filterDaemons(daemons []map[string]any, daemonTypes string) []map[string]any {
	if daemonTypes == "" {
		return daemons
	}
	allowed := map[string]bool{}
	for _, value := range strings.Split(daemonTypes, ",") {
		allowed[strings.TrimSpace(value)] = true
	}
	var filtered []map[string]any
	for _, daemon := range daemons {
		if allowed[stringField(daemon, "daemon_type")] || allowed[typeFromDaemonName(stringField(daemon, "daemon_name"))] {
			filtered = append(filtered, daemon)
		}
	}
	return filtered
}

func filterRecords(records []map[string]any, search string) []map[string]any {
	if strings.TrimSpace(search) == "" {
		return records
	}
	normalized := strings.ToLower(search)
	var filtered []map[string]any
	for _, record := range records {
		data, _ := json.Marshal(record)
		if strings.Contains(strings.ToLower(string(data)), normalized) {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

func stringField(record map[string]any, key string) string {
	value, ok := record[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		return strings.TrimSuffix(strings.TrimSuffix(strconv.FormatFloat(typed, 'f', -1, 64), ".0"), ".")
	default:
		return strings.TrimSpace(strings.Trim(recordString(value), "\""))
	}
}

func recordString(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

func typeFromDaemonName(name string) string {
	if index := strings.Index(name, "."); index > 0 {
		return name[:index]
	}
	return ""
}
