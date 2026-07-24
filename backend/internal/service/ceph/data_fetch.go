package ceph

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"cephtower/backend/internal/integrations/ceph"
	"cephtower/backend/internal/integrations/ceph/command"
	"cephtower/backend/internal/integrations/ceph/dashboard"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

const (
	DataFetchSettingPrefix          = "ceph.data_fetch."
	fetchModuleSummary              = "summary"
	fetchModuleHealth               = "health"
	fetchModuleHosts                = "hosts"
	fetchModuleOSDs                 = "osds"
	fetchModuleOSDFlags             = "osd_flags"
	fetchModuleDaemons              = "daemons"
	fetchModuleServices             = "services"
	fetchModuleMonitors             = "monitors"
	fetchModuleManagers             = "managers"
	fetchModuleMDS                  = "mds"
	fetchModuleMgrModules           = "mgr_modules"
	fetchModuleClusterConfiguration = "cluster_configuration"
	fetchModulePools                = "pools"
	fetchModuleRBDImages            = "rbd_images"
	fetchModuleCephFS               = "cephfs"
	fetchModuleRGWDaemons           = "rgw_daemons"
	fetchModuleRGWUsers             = "rgw_users"
	fetchModuleRGWBuckets           = "rgw_buckets"
	fetchSourceCommand              = "command"
	fetchSourceDashboard            = "dashboard"
)

type Service struct {
	database func() *gorm.DB
	workDir  string
}

func NewService(database func() *gorm.DB, workDirs ...string) Service {
	workDir := ""
	if len(workDirs) > 0 {
		workDir = workDirs[0]
	}
	return Service{database: database, workDir: workDir}
}

type dataFetchModuleDefault struct {
	module          string
	source          string
	intervalSeconds int
	priority        int
}

type DataFetchConfig struct {
	Module              string `json:"module"`
	Enabled             bool   `json:"enabled"`
	IntervalSeconds     int    `json:"interval_seconds"`
	TimeoutSeconds      int    `json:"timeout_seconds"`
	JitterSeconds       int    `json:"jitter_seconds"`
	FetchSource         string `json:"fetch_source"`
	Priority            int    `json:"priority"`
	MaxRetries          int    `json:"max_retries"`
	RetryBackoffSeconds int    `json:"retry_backoff_seconds"`
}

type dataFetchResult struct {
	source          string
	recordsUpserted int
	recordsDeleted  int
}

var defaultDataFetchModules = []dataFetchModuleDefault{
	{fetchModuleSummary, fetchSourceDashboard, 60, 10},
	{fetchModuleHealth, fetchSourceDashboard, 60, 20},
	{fetchModuleHosts, fetchSourceCommand, 300, 30},
	{fetchModuleDaemons, fetchSourceCommand, 300, 40},
	{fetchModuleServices, fetchSourceCommand, 300, 50},
	{fetchModuleMonitors, fetchSourceCommand, 300, 60},
	{fetchModuleManagers, fetchSourceCommand, 300, 70},
	{fetchModuleMDS, fetchSourceCommand, 300, 80},
	{fetchModuleOSDs, fetchSourceCommand, 300, 90},
	{fetchModuleOSDFlags, fetchSourceCommand, 300, 100},
	{fetchModuleMgrModules, fetchSourceCommand, 600, 110},
	{fetchModuleClusterConfiguration, fetchSourceCommand, 900, 120},
	{fetchModulePools, fetchSourceDashboard, 300, 130},
	{fetchModuleRBDImages, fetchSourceDashboard, 300, 140},
	{fetchModuleCephFS, fetchSourceDashboard, 300, 150},
	{fetchModuleRGWDaemons, fetchSourceDashboard, 300, 160},
	{fetchModuleRGWUsers, fetchSourceDashboard, 600, 170},
	{fetchModuleRGWBuckets, fetchSourceDashboard, 600, 180},
}

func StartDataFetchScheduler(database func() *gorm.DB, workDirs ...string) context.CancelFunc {
	workDir := ""
	if len(workDirs) > 0 {
		workDir = workDirs[0]
	}
	service := Service{database: database, workDir: workDir}
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		service.runDueDataFetchSettings(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				service.runDueDataFetchSettings(ctx)
			}
		}
	}()

	return cancel
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
	var clusters []store.CephCluster
	if err := db.WithContext(ctx).Order("id asc").Find(&clusters).Error; err != nil {
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

func EnsureDefaultSystemSettings(ctx context.Context, db *gorm.DB) error {
	for _, item := range defaultDataFetchModules {
		config := defaultDataFetchConfig(item)
		setting := store.Setting{
			Key:   DataFetchSettingKey(item.module),
			Value: mustJSON(config),
		}
		if err := db.WithContext(ctx).Where("`key` = ?", setting.Key).FirstOrCreate(&setting).Error; err != nil {
			return err
		}
	}
	return nil
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

func dataFetchConfigs(ctx context.Context, db *gorm.DB) ([]DataFetchConfig, error) {
	var settings []store.Setting
	if err := db.WithContext(ctx).
		Where("`key` LIKE ?", DataFetchSettingPrefix+"%").
		Order("`key` asc").
		Find(&settings).Error; err != nil {
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

func dataFetchDue(ctx context.Context, db *gorm.DB, clusterID uint, config DataFetchConfig) (bool, error) {
	var latest store.CephDataFetchRun
	err := db.WithContext(ctx).
		Where("cluster_id = ? AND module = ?", clusterID, config.Module).
		Order("started_at desc").
		First(&latest).Error
	if err == gorm.ErrRecordNotFound {
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
	return time.Since(latest.StartedAt) >= interval, nil
}

func (service Service) RunDataFetchConfig(ctx context.Context, clusterID uint, config DataFetchConfig) error {
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
	_ = db.WithContext(ctx).Create(&run).Error

	timeout := time.Duration(config.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := service.fetchClusterModule(runCtx, clusterID, config.Module)
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
	_ = db.WithContext(ctx).Model(&store.CephDataFetchRun{}).Where("id = ?", run.ID).Updates(runUpdates).Error
	return err
}

func RunDataFetchModule(ctx context.Context, database func() *gorm.DB, clusterID uint, module string) (dataFetchResult, error) {
	return Service{database: database}.fetchClusterModule(ctx, clusterID, module)
}

func (service Service) fetchClusterModule(ctx context.Context, clusterID uint, module string) (dataFetchResult, error) {
	db := service.database()
	var cluster store.CephCluster
	if err := db.WithContext(ctx).First(&cluster, clusterID).Error; err != nil {
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
	default:
		return dataFetchResult{}, fmt.Errorf("unsupported ceph data fetch module %q", module)
	}
}

func commandFetch(ctx context.Context, workDir string, db *gorm.DB, cluster *store.CephCluster, module string, run func(*command.CommandClient) error) (dataFetchResult, error) {
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

func dashboardFetchArray(ctx context.Context, workDir string, db *gorm.DB, cluster *store.CephCluster, module string, path string, query url.Values, save func(context.Context, *gorm.DB, uint, []map[string]any) int) (dataFetchResult, error) {
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

func countModuleRecords(ctx context.Context, db *gorm.DB, clusterID uint, module string) (int, error) {
	var count int64
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
	default:
		return 0, nil
	}
	err := db.WithContext(ctx).Model(model).Where("cluster_id = ?", clusterID).Count(&count).Error
	return int(count), err
}
