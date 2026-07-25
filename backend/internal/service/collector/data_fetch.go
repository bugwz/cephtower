package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"cephtower/backend/internal/integration/ceph"
	"cephtower/backend/internal/integration/ceph/command"
	"cephtower/backend/internal/integration/ceph/dashboard"
	"cephtower/backend/internal/store"
)

// RunDueDataFetchSettings executes one due-check pass. Scheduling and process
// lifecycle are owned by the task manager rather than this service.
func (service Service) RunDueDataFetchSettings(ctx context.Context) error {
	service.runDueDataFetchSettings(ctx)
	return nil
}

func (service Service) RunDue(ctx context.Context) error {
	return service.RunDueDataFetchSettings(ctx)
}

func (service Service) runDueDataFetchSettings(ctx context.Context) {
	db := service.database()
	if db == nil {
		return
	}
	if err := EnsureDefaultSystemSettings(ctx, db); err != nil {
		slog.Warn("ensure ceph data fetch settings", "error", err)
		return
	}
	configs, err := dataFetchConfigs(ctx, db)
	if err != nil {
		slog.Warn("list ceph data fetch settings", "error", err)
		return
	}
	clusters, err := db.ListClusters(ctx)
	if err != nil {
		slog.Warn("list ceph clusters for data fetch", "error", err)
		return
	}
	for _, cluster := range clusters {
		for _, config := range configs {
			if !config.Enabled {
				continue
			}
			due, err := dataFetchDue(ctx, db, cluster.ID, config)
			if err != nil {
				slog.Warn("check ceph data fetch due", "cluster_id", cluster.ID, "module", config.Module, "error", err)
				continue
			}
			if !due {
				continue
			}
			if err := service.RunDataFetchConfig(ctx, cluster.ID, config); err != nil {
				slog.Warn("run ceph data fetch", "cluster_id", cluster.ID, "module", config.Module, "error", err)
			}
		}
	}
}

func EnsureDefaultSystemSettings(ctx context.Context, db *store.Database) error {
	settings := make([]store.Setting, 0, len(defaultDataFetchModules))
	for _, item := range defaultDataFetchModules {
		config := defaultDataFetchConfig(item)
		setting := store.Setting{
			Key:   DataFetchSettingKey(item.module),
			Value: mustJSON(config),
		}
		settings = append(settings, setting)
	}
	return db.EnsureSettings(ctx, settings)
}

func defaultDataFetchConfig(item dataFetchModuleDefault) DataFetchConfig {
	return DataFetchConfig{
		Module:              item.module,
		Enabled:             true,
		IntervalSeconds:     item.intervalSeconds,
		TimeoutSeconds:      30,
		JitterSeconds:       30,
		FetchSource:         item.source,
		Priority:            item.priority,
		MaxRetries:          3,
		RetryBackoffSeconds: 30,
	}
}

func DataFetchSettingKey(module string) string {
	return DataFetchSettingPrefix + module
}

func dataFetchConfigs(ctx context.Context, db *store.Database) ([]DataFetchConfig, error) {
	settings, err := db.ListSettings(ctx, DataFetchSettingPrefix)
	if err != nil {
		return nil, err
	}
	configs := make([]DataFetchConfig, 0, len(settings))
	for _, setting := range settings {
		var config DataFetchConfig
		if err := json.Unmarshal([]byte(setting.Value), &config); err != nil {
			continue
		}
		if config.Module == "" {
			config.Module = setting.Key[len(DataFetchSettingPrefix):]
		}
		NormalizeDataFetchConfig(&config)
		configs = append(configs, config)
	}
	return configs, nil
}

func NormalizeDataFetchConfig(config *DataFetchConfig) {
	if config.IntervalSeconds < 10 {
		config.IntervalSeconds = 300
	}
	if config.TimeoutSeconds <= 0 {
		config.TimeoutSeconds = 30
	}
	if config.FetchSource == "" {
		config.FetchSource = fetchSourceCommand
	}
	if config.MaxRetries < 0 {
		config.MaxRetries = 0
	}
	if config.RetryBackoffSeconds < 0 {
		config.RetryBackoffSeconds = 0
	}
}

func dataFetchDue(ctx context.Context, db *store.Database, clusterID uint, config DataFetchConfig) (bool, error) {
	latest, err := db.LatestDataFetchRun(ctx, clusterID, config.Module)
	if err == store.ErrRecordNotFound {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if latest.Status == "running" {
		timeout := time.Duration(config.TimeoutSeconds) * time.Second
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		return time.Since(latest.StartedAt) > timeout, nil
	}
	interval := time.Duration(config.IntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	interval += deterministicJitter(clusterID, config.Module, config.JitterSeconds)
	return time.Since(latest.StartedAt) >= interval, nil
}

func (service Service) RunDataFetchConfig(ctx context.Context, clusterID uint, config DataFetchConfig) error {
	key := fmt.Sprintf("%d:%s", clusterID, config.Module)
	if !service.startRun(key) {
		return fmt.Errorf("%w: %s", ErrModuleRunning, key)
	}
	defer service.finishRun(key)
	db := service.database()
	if db == nil {
		return nil
	}
	startedAt := time.Now()
	run := store.CephDataFetchRun{
		ClusterID: clusterID,
		Module:    config.Module,
		Status:    "running",
		Source:    config.FetchSource,
		StartedAt: startedAt,
		Error:     "",
	}
	_ = db.CreateDataFetchRun(ctx, &run)

	timeout := time.Duration(config.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := service.fetchWithRetry(runCtx, clusterID, config)
	finishedAt := time.Now()
	status := "success"
	lastError := ""
	if err != nil {
		status = "failed"
		lastError = err.Error()
	}
	runUpdates := map[string]any{
		"status":           status,
		"source":           result.source,
		"finished_at":      finishedAt,
		"duration_ms":      int(finishedAt.Sub(startedAt).Milliseconds()),
		"records_upserted": result.recordsUpserted,
		"records_deleted":  result.recordsDeleted,
		"error":            lastError,
	}
	_ = db.FinishDataFetchRun(ctx, run.ID, runUpdates)
	return err
}

func (service Service) fetchWithRetry(ctx context.Context, clusterID uint, config DataFetchConfig) (dataFetchResult, error) {
	var result dataFetchResult
	var err error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		result, err = service.fetchClusterModule(ctx, clusterID, config.Module)
		if err == nil {
			return result, nil
		}
		if attempt == config.MaxRetries {
			break
		}
		backoff := time.Duration(config.RetryBackoffSeconds) * time.Second * time.Duration(attempt+1)
		if backoff <= 0 {
			continue
		}
		slog.Warn("retry ceph data fetch", "cluster_id", clusterID, "module", config.Module, "attempt", attempt+2, "error", err)
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			stopTimer(timer)
			return result, ctx.Err()
		case <-timer.C:
		}
	}
	return result, err
}

func deterministicJitter(clusterID uint, module string, seconds int) time.Duration {
	if seconds <= 0 {
		return 0
	}
	hash := fnv.New32a()
	_, _ = fmt.Fprintf(hash, "%d:%s", clusterID, module)
	return time.Duration(hash.Sum32()%uint32(seconds+1)) * time.Second
}

func stopTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}

func (service Service) startRun(key string) bool {
	if service.runs == nil {
		return true
	}
	service.runs.mu.Lock()
	defer service.runs.mu.Unlock()
	if _, exists := service.runs.active[key]; exists {
		return false
	}
	service.runs.active[key] = struct{}{}
	return true
}

func (service Service) finishRun(key string) {
	if service.runs == nil {
		return
	}
	service.runs.mu.Lock()
	delete(service.runs.active, key)
	service.runs.mu.Unlock()
}

func (service Service) RunConfig(ctx context.Context, clusterID uint, config DataFetchConfig) error {
	return service.RunDataFetchConfig(ctx, clusterID, config)
}

func RunDataFetchModule(ctx context.Context, database func() *store.Database, clusterID uint, module string) (dataFetchResult, error) {
	return NewService(database).fetchClusterModule(ctx, clusterID, module)
}

func (service Service) fetchClusterModule(ctx context.Context, clusterID uint, module string) (dataFetchResult, error) {
	db := service.database()
	cluster, err := db.FindCluster(ctx, clusterID)
	if err != nil {
		return dataFetchResult{}, err
	}

	switch module {
	case fetchModuleSummary:
		payload, err := dashboardRaw(ctx, service.workDir, &cluster, http.MethodGet, "/api/summary", nil, nil)
		if err != nil {
			return dataFetchResult{source: fetchSourceDashboard}, err
		}
		return dataFetchResult{source: fetchSourceDashboard, recordsUpserted: 1}, saveDiscoveredSummary(ctx, db, cluster.ID, payload)
	case fetchModuleHealth:
		payload, err := dashboardRaw(ctx, service.workDir, &cluster, http.MethodGet, "/api/health/full", nil, nil)
		if err != nil {
			return dataFetchResult{source: fetchSourceDashboard}, err
		}
		count := saveDiscoveredHealthChecks(ctx, db, cluster.ID, payload)
		return dataFetchResult{source: fetchSourceDashboard, recordsUpserted: count}, nil
	case fetchModuleHosts:
		return commandFetch(ctx, service.workDir, db, &cluster, module, func(client *command.CommandClient) error {
			saveDiscoveredHosts(ctx, db, cluster.ID, func() ([]ceph.Host, error) {
				hosts, err := client.OrchHosts(ctx, command.OrchHostListOptions{Detail: true})
				if err == nil {
					return commandHostsToAPIHosts(hosts), nil
				}
				nodes, nodeErr := client.NodeList(ctx)
				if nodeErr != nil {
					return nil, err
				}
				return nodeListToAPIHosts(nodes), nil
			})
			return nil
		})
	case fetchModuleOSDs:
		return commandFetch(ctx, service.workDir, db, &cluster, module, func(client *command.CommandClient) error {
			saveDiscoveredOSDs(ctx, db, cluster.ID, func() ([]map[string]any, error) {
				daemons, err := client.OrchPS(ctx, command.OrchPSOptions{DaemonType: "osd"})
				if err == nil {
					return daemons, nil
				}
				tree, treeErr := client.OSDTree(ctx)
				if treeErr != nil {
					return nil, err
				}
				return osdTreeNodes(tree), nil
			})
			return nil
		})
	case fetchModuleOSDFlags:
		return commandFetch(ctx, service.workDir, db, &cluster, module, func(client *command.CommandClient) error {
			saveDiscoveredOSDFlags(ctx, db, cluster.ID, func() ([]string, error) {
				dump, err := client.OSDDump(ctx)
				if err != nil {
					return nil, err
				}
				return osdFlagsFromDump(dump), nil
			})
			return nil
		})
	case fetchModuleDaemons:
		return commandFetch(ctx, service.workDir, db, &cluster, module, func(client *command.CommandClient) error {
			saveDiscoveredDaemons(ctx, db, cluster.ID, func() ([]map[string]any, error) {
				return client.OrchPS(ctx, command.OrchPSOptions{})
			})
			return nil
		})
	case fetchModuleServices:
		return commandFetch(ctx, service.workDir, db, &cluster, module, func(client *command.CommandClient) error {
			saveDiscoveredServices(ctx, db, cluster.ID, func() ([]map[string]any, error) {
				return client.OrchList(ctx, command.OrchListOptions{})
			})
			return nil
		})
	case fetchModuleMonitors:
		return commandFetch(ctx, service.workDir, db, &cluster, module, func(client *command.CommandClient) error {
			saveDiscoveredMons(ctx, db, cluster.ID, func() (map[string]any, error) {
				return client.MonDump(ctx)
			})
			return nil
		})
	case fetchModuleManagers:
		return commandFetch(ctx, service.workDir, db, &cluster, module, func(client *command.CommandClient) error {
			saveDiscoveredMgrs(ctx, db, cluster.ID, func() (map[string]any, error) {
				return client.MgrDump(ctx)
			})
			return nil
		})
	case fetchModuleMDS:
		return commandFetch(ctx, service.workDir, db, &cluster, module, func(client *command.CommandClient) error {
			saveDiscoveredMDSs(ctx, db, cluster.ID, func() (map[string]any, error) {
				return client.FSDump(ctx)
			})
			return nil
		})
	case fetchModuleMgrModules:
		return commandFetch(ctx, service.workDir, db, &cluster, module, func(client *command.CommandClient) error {
			saveDiscoveredMgrModules(ctx, db, cluster.ID, func() (map[string]any, error) {
				return client.MgrModuleList(ctx)
			})
			return nil
		})
	case fetchModuleClusterConfiguration:
		return commandFetch(ctx, service.workDir, db, &cluster, module, func(client *command.CommandClient) error {
			saveDiscoveredConfiguration(ctx, db, cluster.ID, func() ([]map[string]any, error) {
				return client.ConfigDump(ctx)
			})
			return nil
		})
	case fetchModulePools:
		values := url.Values{"stats": []string{"true"}}
		return dashboardFetchArray(ctx, service.workDir, db, &cluster, module, "/api/pool", values, saveDiscoveredPools)
	case fetchModuleRBDImages:
		return dashboardFetchArray(ctx, service.workDir, db, &cluster, module, "/api/block/image", nil, saveDiscoveredRBDImages)
	case fetchModuleCephFS:
		return dashboardFetchArray(ctx, service.workDir, db, &cluster, module, "/api/cephfs", nil, saveDiscoveredFilesystems)
	case fetchModuleRGWDaemons:
		return dashboardFetchArray(ctx, service.workDir, db, &cluster, module, "/api/rgw/daemon", nil, saveDiscoveredRGWDaemons)
	case fetchModuleRGWUsers:
		return dashboardFetchArray(ctx, service.workDir, db, &cluster, module, "/api/rgw/user", nil, saveDiscoveredRGWUsers)
	case fetchModuleRGWBuckets:
		return dashboardFetchArray(ctx, service.workDir, db, &cluster, module, "/api/rgw/bucket", nil, saveDiscoveredRGWBuckets)
	case fetchModuleSettings:
		payload, err := dashboardRaw(ctx, service.workDir, &cluster, http.MethodGet, "/api/settings", nil, nil)
		if err != nil {
			return dataFetchResult{source: fetchSourceDashboard}, err
		}
		count := saveDiscoveredSettings(ctx, db, cluster.ID, payload)
		return dataFetchResult{source: fetchSourceDashboard, recordsUpserted: count}, nil
	case fetchModuleFeatureToggles:
		payload, err := dashboardRaw(ctx, service.workDir, &cluster, http.MethodGet, "/api/feature_toggles", nil, nil)
		if err != nil {
			return dataFetchResult{source: fetchSourceDashboard}, err
		}
		count := saveDiscoveredFeatureToggles(ctx, db, cluster.ID, payload)
		return dataFetchResult{source: fetchSourceDashboard, recordsUpserted: count}, nil
	case fetchModuleIntegrationStatus:
		count, err := service.fetchIntegrationStatus(ctx, db, &cluster)
		if err != nil {
			return dataFetchResult{source: fetchSourceDashboard}, err
		}
		return dataFetchResult{source: fetchSourceDashboard, recordsUpserted: count}, nil
	default:
		return dataFetchResult{}, fmt.Errorf("unsupported ceph data fetch module %q", module)
	}
}

func commandFetch(ctx context.Context, workDir string, db *store.Database, cluster *store.CephCluster, module string, run func(*command.CommandClient) error) (dataFetchResult, error) {
	client, cleanup, err := commandClientForCluster(workDir, cluster)
	if err != nil {
		return dataFetchResult{source: fetchSourceCommand}, err
	}
	defer cleanup()
	if err := run(client); err != nil {
		return dataFetchResult{source: fetchSourceCommand}, err
	}
	count, _ := countModuleRecords(ctx, db, cluster.ID, module)
	return dataFetchResult{source: fetchSourceCommand, recordsUpserted: count}, nil
}

func dashboardFetchArray(ctx context.Context, workDir string, db *store.Database, cluster *store.CephCluster, module string, path string, query url.Values, save func(context.Context, *store.Database, uint, []map[string]any) int) (dataFetchResult, error) {
	payload, err := dashboardRaw(ctx, workDir, cluster, http.MethodGet, path, query, nil)
	if err != nil {
		return dataFetchResult{source: fetchSourceDashboard}, err
	}
	records := decodeDashboardRecords(payload)
	count := save(ctx, db, cluster.ID, records)
	return dataFetchResult{source: fetchSourceDashboard, recordsUpserted: count}, nil
}

func dashboardRaw(ctx context.Context, workDir string, cluster *store.CephCluster, method string, path string, query url.Values, body any) (json.RawMessage, error) {
	baseURL, err := dashboardBaseURLForCluster(ctx, workDir, cluster)
	if err != nil {
		return nil, err
	}
	client := dashboard.NewDashboardClient(dashboard.Config{
		BaseURL:     baseURL,
		Username:    cluster.DashboardUsername,
		Password:    cluster.DashboardPassword,
		InsecureTLS: false,
	})
	return client.Raw(ctx, method, path, query, body)
}

func decodeDashboardRecords(payload json.RawMessage) []map[string]any {
	var records []map[string]any
	if err := json.Unmarshal(payload, &records); err == nil {
		return records
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		return nil
	}
	for _, key := range []string{"items", "data", "value", "records", "daemons", "users", "buckets"} {
		if values := mapSliceField(object, key); len(values) > 0 {
			return values
		}
	}
	return []map[string]any{object}
}

func countModuleRecords(ctx context.Context, db *store.Database, clusterID uint, module string) (int, error) {
	var model any
	switch module {
	case fetchModuleHosts:
		model = &store.CephClusterHost{}
	case fetchModuleOSDs:
		model = &store.CephClusterOSD{}
	case fetchModuleOSDFlags:
		model = &store.CephClusterOSDFlag{}
	case fetchModuleDaemons:
		model = &store.CephClusterDaemon{}
	case fetchModuleServices:
		model = &store.CephClusterService{}
	case fetchModuleMonitors:
		model = &store.CephClusterMon{}
	case fetchModuleManagers:
		model = &store.CephClusterMgr{}
	case fetchModuleMDS:
		model = &store.CephClusterMDS{}
	case fetchModuleMgrModules:
		model = &store.CephClusterMgrModule{}
	case fetchModuleClusterConfiguration:
		model = &store.CephClusterConfiguration{}
	case fetchModuleSettings:
		model = &store.CephClusterSettingSnapshot{}
	case fetchModuleFeatureToggles:
		model = &store.CephClusterFeatureToggle{}
	case fetchModuleIntegrationStatus:
		model = &store.CephClusterIntegrationStatus{}
	default:
		return 0, nil
	}
	return db.CountClusterRecords(ctx, clusterID, model)
}
