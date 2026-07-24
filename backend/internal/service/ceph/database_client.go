package ceph

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
	workDir  string
}

func NewDatabaseCephClient(database func() *gorm.DB, workDirs ...string) *databaseCephClient {
	workDir := ""
	if len(workDirs) > 0 {
		workDir = workDirs[0]
	}
	return &databaseCephClient{database: database, workDir: workDir}
}

func (c *databaseCephClient) dashboardClient(ctx context.Context) (*dashboard.DashboardClient, error) {
	var cluster store.CephCluster
	err := c.database().
		WithContext(ctx).
		Order("id asc").
		First(&cluster).
		Error
	if err != nil {
		return nil, err
	}
	baseURL, err := dashboardBaseURLForCluster(ctx, c.workDir, &cluster)
	if err != nil {
		return nil, err
	}

	return dashboard.NewDashboardClient(dashboard.Config{
		BaseURL:     baseURL,
		Username:    cluster.DashboardUsername,
		Password:    cluster.DashboardPassword,
		InsecureTLS: false,
	}), nil
}

func (c *databaseCephClient) ClusterSummary(ctx context.Context) (ceph.ClusterSummary, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil {
		return ceph.ClusterSummary{}, err
	}
	if !ok {
		return ceph.ClusterSummary{HealthStatus: "unknown"}, nil
	}

	var record store.CephClusterSummary
	err = c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ceph.ClusterSummary{HealthStatus: "unknown"}, nil
	}
	if err != nil {
		return ceph.ClusterSummary{}, err
	}

	return ceph.ClusterSummary{
		HealthStatus:      firstNonEmptyString(record.HealthStatus, "unknown"),
		Version:           record.Version,
		MgrID:             record.MgrID,
		MgrHost:           record.MgrHost,
		HaveMonConnection: strconv.FormatBool(record.HaveMonConnection),
		ExecutingTasks:    stringListFromJSON(record.ExecutingTasks),
		FinishedTasks:     taskSummariesFromJSON(record.FinishedTasks),
		Raw:               mapPayload(record.Payload),
	}, nil
}

func (c *databaseCephClient) Version(ctx context.Context) (string, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil || !ok {
		return "", err
	}
	var record store.CephClusterSummary
	err = c.database().WithContext(ctx).Select("version").Where("cluster_id = ?", clusterID).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return record.Version, nil
}

func (c *databaseCephClient) HealthFull(ctx context.Context) (map[string]any, error) {
	return c.localHealth(ctx)
}

func (c *databaseCephClient) HealthMinimal(ctx context.Context) (map[string]any, error) {
	return c.localHealth(ctx)
}

func (c *databaseCephClient) ListHosts(ctx context.Context, options ceph.ListHostsOptions) ([]ceph.Host, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return []ceph.Host{}, nil
	}
	var records []store.CephClusterHost
	if err := c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("hostname asc").Find(&records).Error; err != nil {
		return nil, err
	}
	hosts := make([]ceph.Host, 0, len(records))
	for _, record := range records {
		hosts = append(hosts, ceph.Host{
			Hostname:    record.Hostname,
			Addr:        record.Addr,
			CephVersion: record.CephVersion,
			Status:      record.Status,
			Labels:      stringListFromJSON(record.Labels),
			Sources:     hostSourcesFromJSON(record.Sources),
		})
	}
	return filterHosts(hosts, options), nil
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
	c.fetchModule(ctx, fetchModuleHosts)
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
	c.fetchModule(ctx, fetchModuleHosts)
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
	c.fetchModule(ctx, fetchModuleHosts)
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
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return []map[string]any{}, nil
	}
	var records []store.CephClusterOSD
	if err := c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("osd_id asc").Find(&records).Error; err != nil {
		return nil, err
	}
	osds := make([]map[string]any, 0, len(records))
	for _, record := range records {
		payload := mapPayload(record.Payload)
		if len(payload) == 0 {
			payload = map[string]any{"id": record.OSDID, "hostname": record.Hostname, "status": record.Status}
		}
		osds = append(osds, payload)
	}
	return filterRecords(osds, options.Search), nil
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
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return []string{}, nil
	}
	var records []store.CephClusterOSDFlag
	if err := c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("name asc").Find(&records).Error; err != nil {
		return nil, err
	}
	flags := make([]string, 0, len(records))
	for _, record := range records {
		flags = append(flags, record.Name)
	}
	return flags, nil
}

func (c *databaseCephClient) ListDaemons(ctx context.Context, daemonTypes string) ([]map[string]any, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return []map[string]any{}, nil
	}
	var records []store.CephClusterDaemon
	if err := c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("name asc").Find(&records).Error; err != nil {
		return nil, err
	}
	daemons := make([]map[string]any, 0, len(records))
	for _, record := range records {
		payload := mapPayload(record.Payload)
		if len(payload) == 0 {
			payload = map[string]any{"daemon_name": record.Name, "daemon_type": record.DaemonType, "hostname": record.Hostname, "status": record.Status}
		}
		daemons = append(daemons, payload)
	}
	return filterDaemons(daemons, daemonTypes), nil
}

func (c *databaseCephClient) ApplyDaemonAction(ctx context.Context, daemonName string, request ceph.DaemonActionRequest) error {
	client, err := c.dashboardClient(ctx)
	if err != nil {
		return err
	}
	if err := client.ApplyDaemonAction(ctx, daemonName, request); err != nil {
		return err
	}
	c.fetchModule(ctx, fetchModuleDaemons)
	return nil
}

func (c *databaseCephClient) Raw(ctx context.Context, method string, path string, query url.Values, body any) (json.RawMessage, error) {
	if method == http.MethodGet {
		if payload, ok, err := c.localRaw(ctx, path); err != nil {
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
		c.fetchModulesForPath(ctx, path)
	}
	return payload, nil
}

func (c *databaseCephClient) localRaw(ctx context.Context, path string) (json.RawMessage, bool, error) {
	switch path {
	case "/api/monitor":
		payload, err := c.localMonitorRaw(ctx)
		return payload, true, err
	case "/api/service":
		payload, err := c.localServicesRaw(ctx)
		return payload, true, err
	case "/api/service/known_types":
		return marshalLocalRaw([]string{"mon", "mgr", "osd", "mds", "rgw", "rbd-mirror", "nfs", "iscsi", "nvmeof"})
	case "/api/mgr/module":
		payload, err := c.localMgrModulesRaw(ctx)
		return payload, true, err
	case "/api/pool":
		payload, err := c.localPoolsRaw(ctx)
		return payload, true, err
	case "/api/block/image":
		payload, err := c.localRBDImagesRaw(ctx, false)
		return payload, true, err
	case "/api/block/image/trash":
		payload, err := c.localRBDImagesRaw(ctx, true)
		return payload, true, err
	case "/api/block/image/default_features":
		return marshalLocalRaw(map[string]any{"features": []string{}})
	case "/api/block/image/clone_format_version":
		return marshalLocalRaw(map[string]any{"clone_format_version": 2})
	case "/api/block/mirroring/summary":
		return marshalLocalRaw(map[string]any{})
	case "/api/cephfs":
		payload, err := c.localFilesystemsRaw(ctx)
		return payload, true, err
	case "/api/rgw/daemon":
		payload, err := c.localRGWDaemonsRaw(ctx)
		return payload, true, err
	case "/api/rgw/user":
		payload, err := c.localRGWUsersRaw(ctx)
		return payload, true, err
	case "/api/rgw/bucket":
		payload, err := c.localRGWBucketsRaw(ctx)
		return payload, true, err
	case "/api/rgw/accounts":
		return marshalLocalRaw([]map[string]any{})
	case "/api/cluster_conf", "/api/cluster_conf/filter":
		payload, err := c.localConfigurationRaw(ctx)
		return payload, true, err
	case "/api/logs/all":
		return marshalLocalRaw(map[string]any{"audit_log": []map[string]any{}, "clog": []map[string]any{}})
	default:
		switch {
		case strings.HasPrefix(path, "/api/service/") && strings.HasSuffix(path, "/daemons"):
			name := strings.TrimSuffix(strings.TrimPrefix(path, "/api/service/"), "/daemons")
			payload, err := c.localServiceDaemonsRaw(ctx, name)
			return payload, true, err
		case strings.HasPrefix(path, "/api/pool/") && strings.HasSuffix(path, "/configuration"):
			return marshalLocalRaw([]map[string]any{})
		case strings.HasPrefix(path, "/api/cephfs/"):
			return marshalLocalRaw(map[string]any{})
		case strings.HasPrefix(path, "/api/rgw/daemon/"):
			payload, err := c.localRGWDaemonRaw(ctx, strings.TrimPrefix(path, "/api/rgw/daemon/"))
			return payload, true, err
		case strings.HasPrefix(path, "/api/rgw/user/"):
			payload, err := c.localRGWUserRaw(ctx, strings.TrimPrefix(path, "/api/rgw/user/"))
			return payload, true, err
		case strings.HasPrefix(path, "/api/rgw/bucket/"):
			payload, err := c.localRGWBucketRaw(ctx, strings.TrimPrefix(path, "/api/rgw/bucket/"))
			return payload, true, err
		case strings.HasPrefix(path, "/api/rgw/accounts/"):
			return marshalLocalRaw(map[string]any{})
		case strings.HasPrefix(path, "/api/cluster_conf/"):
			payload, err := c.localConfigurationItemRaw(ctx, strings.TrimPrefix(path, "/api/cluster_conf/"))
			return payload, true, err
		case strings.HasPrefix(path, "/api/osd/"):
			return marshalLocalRaw(map[string]any{})
		default:
			return nil, false, nil
		}
	}
}

func marshalLocalRaw(payload any) (json.RawMessage, bool, error) {
	data, err := json.Marshal(payload)
	return data, true, err
}

func (c *databaseCephClient) localMonitorRaw(ctx context.Context) (json.RawMessage, error) {
	response := map[string]any{
		"in_quorum":  []map[string]any{},
		"out_quorum": []map[string]any{},
		"mons":       []map[string]any{},
	}

	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return json.Marshal(response)
	}

	var records []store.CephClusterMon
	if err := c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("name asc").Find(&records).Error; err != nil {
		return nil, err
	}

	inQuorum := make([]map[string]any, 0, len(records))
	outQuorum := make([]map[string]any, 0, len(records))
	mons := make([]map[string]any, 0, len(records))
	for _, record := range records {
		mon := mapPayload(record.Payload)
		if len(mon) == 0 {
			mon = map[string]any{}
		}
		mon["name"] = firstNonEmptyString(stringField(mon, "name"), record.Name)
		mon["rank"] = firstNonEmptyString(stringField(mon, "rank"), record.Rank)
		mon["addr"] = firstNonEmptyString(stringField(mon, "addr"), record.Addr)
		mon["public_addr"] = firstNonEmptyString(stringField(mon, "public_addr"), record.PublicAddr)
		mon["status"] = firstNonEmptyString(stringField(mon, "status"), record.Status)
		mons = append(mons, mon)
		if record.Status == "in_quorum" {
			inQuorum = append(inQuorum, mon)
		} else {
			outQuorum = append(outQuorum, mon)
		}
	}

	response["in_quorum"] = inQuorum
	response["out_quorum"] = outQuorum
	response["mons"] = mons
	return json.Marshal(response)
}

func (c *databaseCephClient) localHealth(ctx context.Context) (map[string]any, error) {
	response := map[string]any{
		"health": map[string]any{
			"status": "unknown",
			"checks": map[string]any{},
		},
		"checks": map[string]any{},
	}

	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return response, nil
	}

	var summary store.CephClusterSummary
	err = c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).First(&summary).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	status := firstNonEmptyString(summary.HealthStatus, "unknown")

	var records []store.CephClusterHealthCheck
	if err := c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("name asc").Find(&records).Error; err != nil {
		return nil, err
	}
	checks := make(map[string]any, len(records))
	for _, record := range records {
		check := mapPayload(record.Payload)
		check["severity"] = firstNonEmptyString(stringField(check, "severity"), record.Severity)
		check["summary"] = firstNonEmptyString(stringField(check, "summary"), record.Summary)
		if _, ok := check["muted"]; !ok {
			check["muted"] = record.Muted
		}
		if _, ok := check["count"]; !ok {
			check["count"] = record.Count
		}
		if _, ok := check["detail"]; !ok {
			check["detail"] = decodeJSONValue(record.Detail, []any{})
		}
		checks[record.Name] = check
	}

	response["health"] = map[string]any{"status": status, "checks": checks}
	response["checks"] = checks
	return response, nil
}

func (c *databaseCephClient) localServicesRaw(ctx context.Context) (json.RawMessage, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil || !ok {
		return marshalEmptyList(err)
	}
	var records []store.CephClusterService
	if err := c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("service_name asc").Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		item := mapPayload(record.Payload)
		item["service_name"] = firstNonEmptyString(stringField(item, "service_name"), stringField(item, "name"), record.ServiceName)
		item["service_type"] = firstNonEmptyString(stringField(item, "service_type"), stringField(item, "type"), record.ServiceType)
		items = append(items, item)
	}
	return json.Marshal(items)
}

func (c *databaseCephClient) localServiceDaemonsRaw(ctx context.Context, name string) (json.RawMessage, error) {
	daemons, err := c.ListDaemons(ctx, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(daemons)
}

func (c *databaseCephClient) localMgrModulesRaw(ctx context.Context) (json.RawMessage, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil || !ok {
		return marshalEmptyList(err)
	}
	var records []store.CephClusterMgrModule
	if err := c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("name asc").Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		item := mapPayload(record.Payload)
		item["name"] = firstNonEmptyString(stringField(item, "name"), record.Name)
		if _, ok := item["enabled"]; !ok {
			item["enabled"] = record.Enabled
		}
		items = append(items, item)
	}
	return json.Marshal(items)
}

func (c *databaseCephClient) localPoolsRaw(ctx context.Context) (json.RawMessage, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil || !ok {
		return marshalEmptyList(err)
	}
	var records []store.CephPool
	if err := c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("pool_name asc").Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		item := mapPayload(record.Payload)
		item["pool_name"] = firstNonEmptyString(stringField(item, "pool_name"), stringField(item, "name"), record.PoolName)
		item["type"] = firstNonEmptyString(stringField(item, "type"), record.Type)
		items = append(items, item)
	}
	return json.Marshal(items)
}

func (c *databaseCephClient) localRBDImagesRaw(ctx context.Context, trash bool) (json.RawMessage, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil || !ok {
		return marshalEmptyList(err)
	}
	var records []store.CephRBDImage
	if err := c.database().WithContext(ctx).Where("cluster_id = ? AND trash = ?", clusterID, trash).Order("image_spec asc").Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		item := mapPayload(record.Payload)
		item["pool_name"] = firstNonEmptyString(stringField(item, "pool_name"), record.PoolName)
		item["name"] = firstNonEmptyString(stringField(item, "name"), stringField(item, "image"), record.ImageName)
		item["image_spec"] = firstNonEmptyString(stringField(item, "image_spec"), record.ImageSpec)
		items = append(items, item)
	}
	return json.Marshal(items)
}

func (c *databaseCephClient) localFilesystemsRaw(ctx context.Context) (json.RawMessage, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil || !ok {
		return marshalEmptyList(err)
	}
	var records []store.CephFilesystem
	if err := c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("name asc").Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		item := mapPayload(record.Payload)
		item["id"] = firstNonEmptyString(stringField(item, "id"), record.FSID)
		item["name"] = firstNonEmptyString(stringField(item, "name"), record.Name)
		items = append(items, item)
	}
	return json.Marshal(items)
}

func (c *databaseCephClient) localRGWDaemonsRaw(ctx context.Context) (json.RawMessage, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil || !ok {
		return marshalEmptyList(err)
	}
	var records []store.CephRGWDaemon
	if err := c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("service_id asc").Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		items = append(items, rgwDaemonPayload(record))
	}
	return json.Marshal(items)
}

func (c *databaseCephClient) localRGWDaemonRaw(ctx context.Context, serviceID string) (json.RawMessage, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil || !ok {
		return marshalEmptyObject(err)
	}
	var record store.CephRGWDaemon
	err = c.database().WithContext(ctx).Where("cluster_id = ? AND service_id = ?", clusterID, serviceID).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return json.Marshal(map[string]any{})
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(rgwDaemonPayload(record))
}

func rgwDaemonPayload(record store.CephRGWDaemon) map[string]any {
	item := mapPayload(record.Payload)
	item["id"] = firstNonEmptyString(stringField(item, "id"), stringField(item, "service_id"), record.ServiceID)
	item["service_id"] = firstNonEmptyString(stringField(item, "service_id"), record.ServiceID)
	item["hostname"] = firstNonEmptyString(stringField(item, "hostname"), record.Hostname)
	return item
}

func (c *databaseCephClient) localRGWUsersRaw(ctx context.Context) (json.RawMessage, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil || !ok {
		return marshalEmptyList(err)
	}
	var records []store.CephRGWUser
	if err := c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("uid asc").Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		items = append(items, rgwUserPayload(record))
	}
	return json.Marshal(items)
}

func (c *databaseCephClient) localRGWUserRaw(ctx context.Context, uid string) (json.RawMessage, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil || !ok {
		return marshalEmptyObject(err)
	}
	var record store.CephRGWUser
	err = c.database().WithContext(ctx).Where("cluster_id = ? AND uid = ?", clusterID, uid).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return json.Marshal(map[string]any{})
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(rgwUserPayload(record))
}

func rgwUserPayload(record store.CephRGWUser) map[string]any {
	item := mapPayload(record.Payload)
	item["uid"] = firstNonEmptyString(stringField(item, "uid"), record.UID)
	item["display_name"] = firstNonEmptyString(stringField(item, "display_name"), record.DisplayName)
	item["email"] = firstNonEmptyString(stringField(item, "email"), record.Email)
	return item
}

func (c *databaseCephClient) localRGWBucketsRaw(ctx context.Context) (json.RawMessage, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil || !ok {
		return marshalEmptyList(err)
	}
	var records []store.CephRGWBucket
	if err := c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("bucket asc").Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		items = append(items, rgwBucketPayload(record))
	}
	return json.Marshal(items)
}

func (c *databaseCephClient) localRGWBucketRaw(ctx context.Context, bucket string) (json.RawMessage, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil || !ok {
		return marshalEmptyObject(err)
	}
	var record store.CephRGWBucket
	err = c.database().WithContext(ctx).Where("cluster_id = ? AND bucket = ?", clusterID, bucket).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return json.Marshal(map[string]any{})
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(rgwBucketPayload(record))
}

func rgwBucketPayload(record store.CephRGWBucket) map[string]any {
	item := mapPayload(record.Payload)
	item["bucket"] = firstNonEmptyString(stringField(item, "bucket"), record.Bucket)
	item["owner"] = firstNonEmptyString(stringField(item, "owner"), record.Owner)
	return item
}

func (c *databaseCephClient) localConfigurationRaw(ctx context.Context) (json.RawMessage, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil || !ok {
		return marshalEmptyList(err)
	}
	var records []store.CephClusterConfiguration
	if err := c.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("name asc").Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		items = append(items, configurationPayload(record))
	}
	return json.Marshal(items)
}

func (c *databaseCephClient) localConfigurationItemRaw(ctx context.Context, name string) (json.RawMessage, error) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil || !ok {
		return marshalEmptyObject(err)
	}
	var record store.CephClusterConfiguration
	err = c.database().WithContext(ctx).Where("cluster_id = ? AND name = ?", clusterID, name).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return json.Marshal(map[string]any{})
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(configurationPayload(record))
}

func configurationPayload(record store.CephClusterConfiguration) map[string]any {
	item := mapPayload(record.Payload)
	item["name"] = firstNonEmptyString(stringField(item, "name"), record.Name)
	item["who"] = firstNonEmptyString(stringField(item, "who"), record.Who)
	item["value"] = firstNonEmptyString(stringField(item, "value"), record.Value)
	return item
}

func (c *databaseCephClient) firstClusterID(ctx context.Context) (uint, bool, error) {
	var cluster store.CephCluster
	if err := c.database().WithContext(ctx).
		Order("id asc").
		First(&cluster).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return cluster.ID, true, nil
}

func (c *databaseCephClient) fetchModule(ctx context.Context, module string) {
	clusterID, ok, err := c.firstClusterID(ctx)
	if err != nil || !ok {
		return
	}
	go func() {
		_, _ = RunDataFetchModule(context.WithoutCancel(ctx), c.database, clusterID, module)
	}()
}

func (c *databaseCephClient) fetchModulesForPath(ctx context.Context, path string) {
	switch {
	case strings.HasPrefix(path, "/api/host"):
		c.fetchModule(ctx, fetchModuleHosts)
	case strings.HasPrefix(path, "/api/osd"):
		c.fetchModule(ctx, fetchModuleOSDs)
	case strings.HasPrefix(path, "/api/daemon"):
		c.fetchModule(ctx, fetchModuleDaemons)
	case strings.HasPrefix(path, "/api/pool"):
		c.fetchModule(ctx, fetchModulePools)
	case strings.HasPrefix(path, "/api/block/image"):
		c.fetchModule(ctx, fetchModuleRBDImages)
	case strings.HasPrefix(path, "/api/cephfs"):
		c.fetchModule(ctx, fetchModuleCephFS)
	case strings.HasPrefix(path, "/api/rgw/daemon"):
		c.fetchModule(ctx, fetchModuleRGWDaemons)
	case strings.HasPrefix(path, "/api/rgw/user"):
		c.fetchModule(ctx, fetchModuleRGWUsers)
	case strings.HasPrefix(path, "/api/rgw/bucket"):
		c.fetchModule(ctx, fetchModuleRGWBuckets)
	case strings.HasPrefix(path, "/api/cluster_conf"):
		c.fetchModule(ctx, fetchModuleClusterConfiguration)
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

func mapPayload(payload string) map[string]any {
	var decoded map[string]any
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return map[string]any{}
	}
	return decoded
}

func stringListFromJSON(payload string) []string {
	var values []string
	if err := json.Unmarshal([]byte(payload), &values); err == nil {
		return values
	}
	return nil
}

func hostSourcesFromJSON(payload string) ceph.HostSources {
	var sources ceph.HostSources
	_ = json.Unmarshal([]byte(payload), &sources)
	return sources
}

func taskSummariesFromJSON(payload string) []ceph.TaskSummary {
	var values []ceph.TaskSummary
	if err := json.Unmarshal([]byte(payload), &values); err == nil {
		return values
	}
	return nil
}

func decodeJSONValue(payload string, fallback any) any {
	var value any
	if err := json.Unmarshal([]byte(payload), &value); err == nil {
		return value
	}
	return fallback
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

func marshalEmptyList(err error) (json.RawMessage, error) {
	if err != nil {
		return nil, err
	}
	return json.Marshal([]map[string]any{})
}

func marshalEmptyObject(err error) (json.RawMessage, error) {
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{})
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func typeFromDaemonName(name string) string {
	if index := strings.Index(name, "."); index > 0 {
		return name[:index]
	}
	return ""
}
