package ceph

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cephtower/backend/internal/integrations/ceph"
	"cephtower/backend/internal/integrations/ceph/command"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

type ClusterDiscoverer func(context.Context, *gorm.DB, *store.CephCluster) error

const (
	DefaultCephCommandBin            = "ceph"
	DefaultCephCommandName           = "client.admin"
	DefaultCephCommandTimeoutSeconds = 15
)

func DiscoverAndSyncCephCluster(ctx context.Context, db *gorm.DB, cluster *store.CephCluster) error {
	return DiscoverAndSyncCephClusterWithWorkDir(ctx, db, cluster, "")
}

func DiscoverAndSyncCephClusterWithWorkDir(ctx context.Context, db *gorm.DB, cluster *store.CephCluster, workDir string) error {
	if cluster == nil {
		return nil
	}

	client, cleanup, err := commandClientForCluster(workDir, cluster)
	if err != nil {
		return err
	}
	defer cleanup()

	fsid, err := client.FSID(ctx)
	if err != nil {
		return fmt.Errorf("discover ceph fsid: %w", err)
	}
	_ = strings.TrimSpace(fsid)

	services, err := client.MgrServices(ctx)
	if err != nil {
		return fmt.Errorf("discover mgr services: %w", err)
	}
	dashboardURL := dashboardURLFromMgrServices(services)
	if dashboardURL == "" {
		return fmt.Errorf("dashboard service address was not found from ceph mgr services")
	}

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
	saveDiscoveredOSDFlags(ctx, db, cluster.ID, func() ([]string, error) {
		dump, err := client.OSDDump(ctx)
		if err != nil {
			return nil, err
		}
		return osdFlagsFromDump(dump), nil
	})
	saveDiscoveredDaemons(ctx, db, cluster.ID, func() ([]map[string]any, error) {
		return client.OrchPS(ctx, command.OrchPSOptions{})
	})
	saveDiscoveredServices(ctx, db, cluster.ID, func() ([]map[string]any, error) {
		return client.OrchList(ctx, command.OrchListOptions{})
	})
	saveDiscoveredMons(ctx, db, cluster.ID, func() (map[string]any, error) {
		return client.MonDump(ctx)
	})
	saveDiscoveredMgrs(ctx, db, cluster.ID, func() (map[string]any, error) {
		return client.MgrDump(ctx)
	})
	saveDiscoveredMDSs(ctx, db, cluster.ID, func() (map[string]any, error) {
		return client.FSDump(ctx)
	})
	saveDiscoveredMgrModules(ctx, db, cluster.ID, func() (map[string]any, error) {
		return client.MgrModuleList(ctx)
	})
	saveDiscoveredConfiguration(ctx, db, cluster.ID, func() ([]map[string]any, error) {
		return client.ConfigDump(ctx)
	})

	if err := SyncCephClusterRuntimeFiles(workDir, cluster); err != nil {
		return fmt.Errorf("persist ceph runtime files: %w", err)
	}
	return nil
}

func commandClientForCluster(workDir string, cluster *store.CephCluster) (*command.CommandClient, func(), error) {
	paths := cephClusterRuntimePaths(workDir, cluster.ID)
	cleanup := func() {}
	confPath := paths.confPath
	keyringPath := paths.keyringPath
	confMatches := runtimeFileMatches(confPath, cephConfigContent(cluster.MonitorHost))
	keyringMatches := runtimeFileMatches(keyringPath, keyringFileContent(DefaultCephCommandName, cluster.Keyring))
	if !confMatches || !keyringMatches {
		fallbackDir := filepath.Join(os.TempDir(), "cephtower")
		if err := os.MkdirAll(fallbackDir, 0o700); err != nil {
			return nil, cleanup, fmt.Errorf("create ceph fallback directory: %w", err)
		}
		cleanupFiles := make([]string, 0, 2)
		if !confMatches {
			fallbackPath, err := writeTempRuntimeFile(fallbackDir, fmt.Sprintf("cluster-%d-conf-*", cluster.ID), cephConfigContent(cluster.MonitorHost))
			if err != nil {
				return nil, cleanup, err
			}
			confPath = fallbackPath
			cleanupFiles = append(cleanupFiles, fallbackPath)
		}
		if !keyringMatches {
			fallbackPath, err := writeTempRuntimeFile(fallbackDir, fmt.Sprintf("cluster-%d-keyring-*", cluster.ID), keyringFileContent(DefaultCephCommandName, cluster.Keyring))
			if err != nil {
				for _, path := range cleanupFiles {
					_ = os.Remove(path)
				}
				return nil, cleanup, err
			}
			keyringPath = fallbackPath
			cleanupFiles = append(cleanupFiles, fallbackPath)
		}
		cleanup = func() {
			for _, path := range cleanupFiles {
				_ = os.Remove(path)
			}
		}
	}

	timeout := time.Duration(DefaultCephCommandTimeoutSeconds) * time.Second
	client := command.NewCommandClient(command.Config{
		Bin:     DefaultCephCommandBin,
		Cluster: "",
		Conf:    confPath,
		MonHost: "",
		Name:    DefaultCephCommandName,
		Keyring: keyringPath,
		Timeout: timeout,
	})
	return client, cleanup, nil
}

func dashboardBaseURLForCluster(ctx context.Context, workDir string, cluster *store.CephCluster) (string, error) {
	client, cleanup, err := commandClientForCluster(workDir, cluster)
	if err != nil {
		return "", err
	}
	defer cleanup()

	services, err := client.MgrServices(ctx)
	if err != nil {
		return "", fmt.Errorf("discover mgr services: %w", err)
	}
	baseURL := dashboardURLFromMgrServices(services)
	if baseURL == "" {
		return "", fmt.Errorf("dashboard service address was not found from ceph mgr services")
	}
	return baseURL, nil
}

type cephClusterRuntimeFilePaths struct {
	dir         string
	confPath    string
	keyringPath string
}

func cephClusterRuntimePaths(workDir string, clusterID uint) cephClusterRuntimeFilePaths {
	dir := filepath.Join(workDir, "ceph", strconv.FormatUint(uint64(clusterID), 10))
	return cephClusterRuntimeFilePaths{
		dir:         dir,
		confPath:    filepath.Join(dir, "ceph.conf"),
		keyringPath: filepath.Join(dir, "ceph.client.admin.keyring"),
	}
}

func SyncCephClusterRuntimeFiles(workDir string, cluster *store.CephCluster) error {
	if cluster == nil {
		return nil
	}
	paths := cephClusterRuntimePaths(workDir, cluster.ID)
	if err := os.MkdirAll(paths.dir, 0o700); err != nil {
		return fmt.Errorf("create ceph runtime directory: %w", err)
	}
	if err := syncRuntimeFile(paths.confPath, cephConfigContent(cluster.MonitorHost)); err != nil {
		return fmt.Errorf("sync ceph config: %w", err)
	}
	if err := syncRuntimeFile(paths.keyringPath, keyringFileContent(DefaultCephCommandName, cluster.Keyring)); err != nil {
		return fmt.Errorf("sync ceph keyring: %w", err)
	}
	return nil
}

func DeleteCephClusterRuntimeFiles(workDir string, clusterID uint) error {
	paths := cephClusterRuntimePaths(workDir, clusterID)
	if err := os.RemoveAll(paths.dir); err != nil {
		return fmt.Errorf("remove ceph runtime directory: %w", err)
	}
	return nil
}

func SyncCephRuntimeFiles(ctx context.Context, db *gorm.DB, workDir string) error {
	var clusters []store.CephCluster
	if err := db.WithContext(ctx).Order("id asc").Find(&clusters).Error; err != nil {
		return err
	}
	for index := range clusters {
		if err := SyncCephClusterRuntimeFiles(workDir, &clusters[index]); err != nil {
			return fmt.Errorf("sync ceph cluster %d runtime files: %w", clusters[index].ID, err)
		}
	}
	return nil
}

func cephConfigContent(monitorHost string) string {
	return fmt.Sprintf("[global]\nmon host = %s\n", strings.TrimSpace(monitorHost))
}

func syncRuntimeFile(path, content string) error {
	current, err := os.ReadFile(path)
	if err == nil && string(current) == content {
		return os.Chmod(path, 0o600)
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o600)
}

func writeTempRuntimeFile(dir, pattern, content string) (string, error) {
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", fmt.Errorf("create ceph fallback file: %w", err)
	}
	path := file.Name()
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", err
	}
	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return path, nil
}

func regularFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func runtimeFileMatches(path, expected string) bool {
	if !regularFileExists(path) {
		return false
	}
	content, err := os.ReadFile(path)
	return err == nil && string(content) == expected
}

func keyringFileContent(entity string, key string) string {
	key = strings.TrimSpace(key)
	if strings.Contains(key, "[") || strings.Contains(key, "key =") {
		return key
	}
	entity = strings.TrimSpace(entity)
	if entity == "" {
		entity = "client.admin"
	}
	return fmt.Sprintf("[%s]\nkey = %s\n", entity, key)
}

func saveDiscoveredHosts(ctx context.Context, db *gorm.DB, clusterID uint, load func() ([]ceph.Host, error)) {
	hosts, err := load()
	if err != nil {
		return
	}
	now := time.Now()
	records := make([]store.CephClusterHost, 0, len(hosts))
	seen := map[string]bool{}
	for _, host := range hosts {
		hostname := strings.TrimSpace(host.Hostname)
		if hostname == "" || seen[hostname] {
			continue
		}
		seen[hostname] = true
		records = append(records, store.CephClusterHost{
			ClusterID:    clusterID,
			Hostname:     hostname,
			Addr:         host.Addr,
			CephVersion:  host.CephVersion,
			Status:       host.Status,
			Labels:       mustJSON(host.Labels),
			Sources:      mustJSON(host.Sources),
			Payload:      mustJSON(host),
			DiscoveredAt: now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephClusterHost{}, records)
}

func saveDiscoveredOSDs(ctx context.Context, db *gorm.DB, clusterID uint, load func() ([]map[string]any, error)) {
	osds, err := load()
	if err != nil {
		return
	}
	now := time.Now()
	records := make([]store.CephClusterOSD, 0, len(osds))
	seen := map[string]bool{}
	for index, osd := range osds {
		osdID := firstStringField(osd, "id", "osd", "service_id", "daemon_id", "name")
		if osdID == "" {
			osdID = fmt.Sprintf("index-%d", index)
		}
		if seen[osdID] {
			continue
		}
		seen[osdID] = true
		records = append(records, store.CephClusterOSD{
			ClusterID:    clusterID,
			OSDID:        osdID,
			Hostname:     firstStringField(osd, "hostname", "host"),
			Status:       firstStringField(osd, "status", "state"),
			Payload:      mustJSON(osd),
			DiscoveredAt: now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephClusterOSD{}, records)
}

func saveDiscoveredOSDFlags(ctx context.Context, db *gorm.DB, clusterID uint, load func() ([]string, error)) {
	flags, err := load()
	if err != nil {
		return
	}
	now := time.Now()
	seen := map[string]bool{}
	var records []store.CephClusterOSDFlag
	for _, flag := range flags {
		flag = strings.TrimSpace(flag)
		if flag == "" || seen[flag] {
			continue
		}
		seen[flag] = true
		records = append(records, store.CephClusterOSDFlag{
			ClusterID:    clusterID,
			Name:         flag,
			DiscoveredAt: now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephClusterOSDFlag{}, records)
}

func saveDiscoveredDaemons(ctx context.Context, db *gorm.DB, clusterID uint, load func() ([]map[string]any, error)) {
	daemons, err := load()
	if err != nil {
		return
	}
	now := time.Now()
	records := make([]store.CephClusterDaemon, 0, len(daemons))
	seen := map[string]bool{}
	for index, daemon := range daemons {
		name := firstStringField(daemon, "daemon_name", "name", "daemon_id", "service_name")
		if name == "" {
			name = fmt.Sprintf("index-%d", index)
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		daemonType := firstStringField(daemon, "daemon_type", "type", "service_type")
		if daemonType == "" {
			daemonType = typeFromDaemonName(name)
		}
		records = append(records, store.CephClusterDaemon{
			ClusterID:    clusterID,
			Name:         name,
			DaemonType:   daemonType,
			Hostname:     firstStringField(daemon, "hostname", "host"),
			Status:       firstStringField(daemon, "status", "state"),
			Payload:      mustJSON(daemon),
			DiscoveredAt: now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephClusterDaemon{}, records)
}

func saveDiscoveredServices(ctx context.Context, db *gorm.DB, clusterID uint, load func() ([]map[string]any, error)) {
	services, err := load()
	if err != nil {
		return
	}
	now := time.Now()
	records := make([]store.CephClusterService, 0, len(services))
	seen := map[string]bool{}
	for index, service := range services {
		name := firstStringField(service, "service_name", "service_id", "name")
		serviceType := firstStringField(service, "service_type", "type")
		if name == "" {
			name = serviceType
		}
		if name == "" {
			name = fmt.Sprintf("index-%d", index)
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		records = append(records, store.CephClusterService{
			ClusterID:    clusterID,
			ServiceName:  name,
			ServiceType:  serviceType,
			Payload:      mustJSON(service),
			DiscoveredAt: now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephClusterService{}, records)
}

func saveDiscoveredMons(ctx context.Context, db *gorm.DB, clusterID uint, load func() (map[string]any, error)) {
	monitor, err := load()
	if err != nil {
		return
	}
	now := time.Now()
	quorum := stringSet(stringList(monitor["quorum_names"]))
	monmap := mapField(monitor, "monmap")
	monRecords := mapSliceField(monmap, "mons")
	if len(monRecords) == 0 {
		monRecords = mapSliceField(monitor, "mons")
	}
	records := make([]store.CephClusterMon, 0, len(monRecords))
	seen := map[string]bool{}
	for index, mon := range monRecords {
		name := firstStringField(mon, "name", "mon")
		if name == "" {
			name = fmt.Sprintf("index-%d", index)
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		status := "out_quorum"
		if quorum[name] {
			status = "in_quorum"
		}
		records = append(records, store.CephClusterMon{
			ClusterID:    clusterID,
			Name:         name,
			Rank:         firstStringField(mon, "rank"),
			Addr:         firstStringField(mon, "addr"),
			PublicAddr:   firstStringField(mon, "public_addr"),
			Status:       status,
			Payload:      mustJSON(mon),
			DiscoveredAt: now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephClusterMon{}, records)
}

func saveDiscoveredMgrs(ctx context.Context, db *gorm.DB, clusterID uint, load func() (map[string]any, error)) {
	mgrDump, err := load()
	if err != nil {
		return
	}
	now := time.Now()
	var records []store.CephClusterMgr
	activeName := strings.TrimSpace(stringField(mgrDump, "active_name"))
	if activeName != "" {
		records = append(records, store.CephClusterMgr{
			ClusterID:    clusterID,
			Name:         activeName,
			Addr:         stringField(mgrDump, "active_addr"),
			Status:       "active",
			Payload:      mustJSON(mgrDump),
			DiscoveredAt: now,
		})
	}
	for index, standby := range mapSliceField(mgrDump, "standbys") {
		name := firstStringField(standby, "name", "mgr_name", "id")
		if name == "" {
			name = fmt.Sprintf("standby-%d", index)
		}
		records = append(records, store.CephClusterMgr{
			ClusterID:    clusterID,
			Name:         name,
			Addr:         firstStringField(standby, "addr", "mgr_addr"),
			Hostname:     firstStringField(standby, "hostname", "host"),
			Status:       "standby",
			Payload:      mustJSON(standby),
			DiscoveredAt: now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephClusterMgr{}, dedupeMgrs(records))
}

func saveDiscoveredMDSs(ctx context.Context, db *gorm.DB, clusterID uint, load func() (map[string]any, error)) {
	fsDump, err := load()
	if err != nil {
		return
	}
	now := time.Now()
	var records []store.CephClusterMDS
	for fsIndex, fs := range mapSliceField(fsDump, "filesystems") {
		mdsmap := mapField(fs, "mdsmap")
		fsName := firstStringField(mdsmap, "fs_name", "name")
		if fsName == "" {
			fsName = firstStringField(fs, "name")
		}
		if fsName == "" {
			fsName = fmt.Sprintf("fs-%d", fsIndex)
		}
		for _, mds := range mapValues(mapField(mdsmap, "info")) {
			name := firstStringField(mds, "name")
			gid := firstStringField(mds, "gid")
			if name == "" {
				name = gid
			}
			if name == "" {
				continue
			}
			records = append(records, store.CephClusterMDS{
				ClusterID:    clusterID,
				Name:         name,
				Filesystem:   fsName,
				Rank:         firstStringField(mds, "rank"),
				GID:          gid,
				Addr:         firstStringField(mds, "addr"),
				Hostname:     firstStringField(mds, "hostname", "host"),
				State:        firstStringField(mds, "state"),
				Payload:      mustJSON(mds),
				DiscoveredAt: now,
			})
		}
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephClusterMDS{}, dedupeMDSs(records))
}

func saveDiscoveredMgrModules(ctx context.Context, db *gorm.DB, clusterID uint, load func() (map[string]any, error)) {
	modules, err := load()
	if err != nil {
		return
	}
	now := time.Now()
	records := mgrModuleRecords(clusterID, now, modules, true, "enabled_modules")
	records = append(records, mgrModuleRecords(clusterID, now, modules, false, "disabled_modules")...)
	if len(records) == 0 {
		records = append(records, store.CephClusterMgrModule{
			ClusterID:    clusterID,
			Name:         "all",
			Payload:      mustJSON(modules),
			DiscoveredAt: now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephClusterMgrModule{}, records)
}

func saveDiscoveredConfiguration(ctx context.Context, db *gorm.DB, clusterID uint, load func() ([]map[string]any, error)) {
	configuration, err := load()
	if err != nil {
		return
	}
	now := time.Now()
	records := make([]store.CephClusterConfiguration, 0, len(configuration))
	seen := map[string]bool{}
	for index, config := range configuration {
		name := firstStringField(config, "name", "option")
		if name == "" {
			name = fmt.Sprintf("index-%d", index)
		}
		who := firstStringField(config, "who", "section", "daemon")
		key := who + "\x00" + name
		if seen[key] {
			continue
		}
		seen[key] = true
		records = append(records, store.CephClusterConfiguration{
			ClusterID:    clusterID,
			Who:          who,
			Name:         name,
			Level:        firstStringField(config, "level", "source"),
			Value:        firstStringField(config, "value"),
			Payload:      mustJSON(config),
			DiscoveredAt: now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephClusterConfiguration{}, records)
}

func saveDiscoveredSummary(ctx context.Context, db *gorm.DB, clusterID uint, payload json.RawMessage) error {
	var summary map[string]any
	if err := json.Unmarshal(payload, &summary); err != nil {
		return err
	}
	now := time.Now()
	record := store.CephClusterSummary{
		ClusterID:         clusterID,
		HealthStatus:      firstStringField(summary, "health_status"),
		Version:           firstStringField(summary, "version"),
		MgrID:             firstStringField(summary, "mgr_id"),
		MgrHost:           firstStringField(summary, "mgr_host"),
		HaveMonConnection: boolField(summary, "have_mon_connection"),
		ExecutingTasks:    mustJSON(summary["executing_tasks"]),
		FinishedTasks:     mustJSON(summary["finished_tasks"]),
		Payload:           string(payload),
		DiscoveredAt:      now,
	}
	return db.WithContext(ctx).
		Where("cluster_id = ?", clusterID).
		Assign(record).
		FirstOrCreate(&record).Error
}

func saveDiscoveredHealthChecks(ctx context.Context, db *gorm.DB, clusterID uint, payload json.RawMessage) int {
	var health map[string]any
	if err := json.Unmarshal(payload, &health); err != nil {
		return 0
	}
	checks := mapField(health, "checks")
	now := time.Now()
	records := make([]store.CephClusterHealthCheck, 0, len(checks))
	for name, value := range checks {
		check, ok := value.(map[string]any)
		if !ok {
			continue
		}
		records = append(records, store.CephClusterHealthCheck{
			ClusterID:    clusterID,
			Name:         name,
			Severity:     firstStringField(check, "severity"),
			Summary:      firstStringField(check, "summary"),
			Detail:       mustJSON(check["detail"]),
			Muted:        boolField(check, "muted"),
			Count:        intField(check, "count"),
			Payload:      mustJSON(check),
			DiscoveredAt: now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephClusterHealthCheck{}, records)
	return len(records)
}

func saveDiscoveredPools(ctx context.Context, db *gorm.DB, clusterID uint, pools []map[string]any) int {
	now := time.Now()
	records := make([]store.CephPool, 0, len(pools))
	seen := map[string]bool{}
	for index, pool := range pools {
		name := firstStringField(pool, "pool_name", "name")
		if name == "" {
			name = fmt.Sprintf("pool-%d", index)
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		stats := mapField(pool, "stats")
		records = append(records, store.CephPool{
			ClusterID:           clusterID,
			PoolID:              firstStringField(pool, "pool", "pool_id"),
			PoolName:            name,
			Type:                firstStringField(pool, "type"),
			Size:                intField(pool, "size"),
			MinSize:             intField(pool, "min_size"),
			PGNum:               intField(pool, "pg_num"),
			PGPlacementNum:      intField(pool, "pg_placement_num"),
			PGAutoscaleMode:     firstStringField(pool, "pg_autoscale_mode"),
			CrushRule:           firstStringField(pool, "crush_rule"),
			ErasureCodeProfile:  firstStringField(pool, "erasure_code_profile"),
			ApplicationMetadata: mustJSON(pool["application_metadata"]),
			QuotaMaxBytes:       int64Field(pool, "quota_max_bytes"),
			QuotaMaxObjects:     int64Field(pool, "quota_max_objects"),
			UsedBytes:           firstInt64Field(stats, "bytes_used", "stored", "used_bytes"),
			MaxAvailBytes:       firstInt64Field(stats, "max_avail", "max_avail_bytes"),
			Objects:             firstInt64Field(stats, "objects", "num_objects"),
			Payload:             mustJSON(pool),
			DiscoveredAt:        now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephPool{}, records)
	return len(records)
}

func saveDiscoveredRBDImages(ctx context.Context, db *gorm.DB, clusterID uint, images []map[string]any) int {
	now := time.Now()
	var records []store.CephRBDImage
	seen := map[string]bool{}
	for index, item := range images {
		poolName := firstStringField(item, "pool_name", "pool")
		values := stringList(item["value"])
		if len(values) == 0 {
			values = []string{firstStringField(item, "name", "image", "image_name")}
		}
		for _, imageName := range values {
			imageName = strings.TrimSpace(imageName)
			if imageName == "" {
				imageName = fmt.Sprintf("image-%d", index)
			}
			imageSpec := imageName
			if poolName != "" {
				imageSpec = poolName + "/" + imageName
			}
			if seen[imageSpec] {
				continue
			}
			seen[imageSpec] = true
			records = append(records, store.CephRBDImage{
				ClusterID:    clusterID,
				PoolName:     poolName,
				Namespace:    firstStringField(item, "namespace"),
				ImageName:    imageName,
				ImageSpec:    imageSpec,
				ImageID:      firstStringField(item, "id", "image_id"),
				SizeBytes:    int64Field(item, "size"),
				ObjectSize:   intField(item, "obj_size"),
				Features:     mustJSON(item["features"]),
				Parent:       mustJSON(item["parent"]),
				Snapshots:    mustJSON(item["snapshots"]),
				MirrorMode:   firstStringField(item, "mirror_mode"),
				Trash:        boolField(item, "trash"),
				Payload:      mustJSON(item),
				DiscoveredAt: now,
			})
		}
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephRBDImage{}, records)
	return len(records)
}

func saveDiscoveredFilesystems(ctx context.Context, db *gorm.DB, clusterID uint, filesystems []map[string]any) int {
	now := time.Now()
	records := make([]store.CephFilesystem, 0, len(filesystems))
	seen := map[string]bool{}
	for index, fs := range filesystems {
		mdsmap := mapField(fs, "mdsmap")
		name := firstStringField(fs, "name", "fs_name")
		if name == "" {
			name = firstStringField(mdsmap, "fs_name", "name")
		}
		fsID := firstStringField(fs, "id", "fs_id")
		if fsID == "" {
			fsID = firstStringField(mdsmap, "fs_id")
		}
		if fsID == "" {
			fsID = fmt.Sprintf("fs-%d", index)
		}
		if seen[fsID] {
			continue
		}
		seen[fsID] = true
		records = append(records, store.CephFilesystem{
			ClusterID:    clusterID,
			FSID:         fsID,
			Name:         name,
			MetadataPool: firstStringField(fs, "metadata_pool"),
			DataPools:    mustJSON(fs["data_pools"]),
			MDSMap:       mustJSON(mdsmap),
			StandbyCount: intField(fs, "standby_count"),
			ClientCount:  intField(fs, "client_count"),
			UsedBytes:    int64Field(fs, "used_bytes"),
			AvailBytes:   int64Field(fs, "avail_bytes"),
			Payload:      mustJSON(fs),
			DiscoveredAt: now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephFilesystem{}, records)
	return len(records)
}

func saveDiscoveredRGWDaemons(ctx context.Context, db *gorm.DB, clusterID uint, daemons []map[string]any) int {
	now := time.Now()
	records := make([]store.CephRGWDaemon, 0, len(daemons))
	seen := map[string]bool{}
	for index, daemon := range daemons {
		serviceID := firstStringField(daemon, "id", "service_id", "daemon_id", "name")
		if serviceID == "" {
			serviceID = fmt.Sprintf("rgw-%d", index)
		}
		if seen[serviceID] {
			continue
		}
		seen[serviceID] = true
		records = append(records, store.CephRGWDaemon{
			ClusterID:      clusterID,
			ServiceID:      serviceID,
			Hostname:       firstStringField(daemon, "hostname", "host"),
			ZoneName:       firstStringField(daemon, "zone_name", "zone"),
			FrontendConfig: firstStringField(daemon, "frontend_config", "frontend"),
			Version:        firstStringField(daemon, "version", "ceph_version"),
			Payload:        mustJSON(daemon),
			DiscoveredAt:   now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephRGWDaemon{}, records)
	return len(records)
}

func saveDiscoveredRGWUsers(ctx context.Context, db *gorm.DB, clusterID uint, users []map[string]any) int {
	now := time.Now()
	records := make([]store.CephRGWUser, 0, len(users))
	seen := map[string]bool{}
	for index, user := range users {
		uid := firstStringField(user, "uid", "user_id", "id")
		if uid == "" {
			uid = fmt.Sprintf("user-%d", index)
		}
		if seen[uid] {
			continue
		}
		seen[uid] = true
		records = append(records, store.CephRGWUser{
			ClusterID:    clusterID,
			UID:          uid,
			DisplayName:  firstStringField(user, "display_name", "name"),
			Email:        firstStringField(user, "email"),
			Suspended:    boolField(user, "suspended"),
			MaxBuckets:   intField(user, "max_buckets"),
			Subusers:     mustJSON(user["subusers"]),
			KeysRedacted: mustJSON(redactedKeys(user["keys"])),
			Caps:         mustJSON(user["caps"]),
			Quota:        mustJSON(user["quota"]),
			Stats:        mustJSON(user["stats"]),
			Payload:      mustJSON(redactRGWSecrets(user)),
			DiscoveredAt: now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephRGWUser{}, records)
	return len(records)
}

func saveDiscoveredRGWBuckets(ctx context.Context, db *gorm.DB, clusterID uint, buckets []map[string]any) int {
	now := time.Now()
	records := make([]store.CephRGWBucket, 0, len(buckets))
	seen := map[string]bool{}
	for index, bucket := range buckets {
		name := firstStringField(bucket, "bucket", "name")
		if name == "" {
			name = fmt.Sprintf("bucket-%d", index)
		}
		tenant := firstStringField(bucket, "tenant")
		key := tenant + "\x00" + name
		if seen[key] {
			continue
		}
		seen[key] = true
		stats := mapField(bucket, "stats")
		records = append(records, store.CephRGWBucket{
			ClusterID:     clusterID,
			Tenant:        tenant,
			Bucket:        name,
			Owner:         firstStringField(bucket, "owner", "uid"),
			Zonegroup:     firstStringField(bucket, "zonegroup"),
			PlacementRule: firstStringField(bucket, "placement_rule"),
			Versioning:    firstStringField(bucket, "versioning"),
			ObjectCount:   firstInt64Field(bucket, "num_objects", "object_count"),
			UsedBytes:     firstInt64Field(bucket, "size", "used_bytes"),
			Quota:         mustJSON(bucket["quota"]),
			Lifecycle:     mustJSON(bucket["lifecycle"]),
			Encryption:    mustJSON(bucket["encryption"]),
			Payload:       mustJSON(bucket),
			DiscoveredAt:  now,
		})
		if records[len(records)-1].ObjectCount == 0 {
			records[len(records)-1].ObjectCount = firstInt64Field(stats, "num_objects", "objects")
		}
		if records[len(records)-1].UsedBytes == 0 {
			records[len(records)-1].UsedBytes = firstInt64Field(stats, "size", "size_actual", "bytes_used")
		}
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephRGWBucket{}, records)
	return len(records)
}

func replaceDiscoveredRecords[T any](ctx context.Context, db *gorm.DB, clusterID uint, model any, records []T) {
	if err := db.WithContext(ctx).Where("cluster_id = ?", clusterID).Delete(model).Error; err != nil {
		return
	}
	if len(records) > 0 {
		_ = db.WithContext(ctx).Create(&records).Error
	}
}

func mustJSON(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		return "null"
	}
	return string(payload)
}

func firstStringField(record map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(stringField(record, key)); value != "" {
			return value
		}
	}
	return ""
}

func mapField(record map[string]any, key string) map[string]any {
	value, ok := record[key].(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return value
}

func mapSliceField(record map[string]any, key string) []map[string]any {
	values, ok := record[key].([]any)
	if !ok {
		return nil
	}
	records := make([]map[string]any, 0, len(values))
	for _, value := range values {
		if record, ok := value.(map[string]any); ok {
			records = append(records, record)
		}
	}
	return records
}

func mapValues(record map[string]any) []map[string]any {
	values := make([]map[string]any, 0, len(record))
	for _, value := range record {
		if item, ok := value.(map[string]any); ok {
			values = append(values, item)
		}
	}
	return values
}

func stringSet(values []string) map[string]bool {
	set := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			set[value] = true
		}
	}
	return set
}

func dedupeMgrs(records []store.CephClusterMgr) []store.CephClusterMgr {
	seen := map[string]bool{}
	result := make([]store.CephClusterMgr, 0, len(records))
	for _, record := range records {
		if strings.TrimSpace(record.Name) == "" || seen[record.Name] {
			continue
		}
		seen[record.Name] = true
		result = append(result, record)
	}
	return result
}

func dedupeMDSs(records []store.CephClusterMDS) []store.CephClusterMDS {
	seen := map[string]bool{}
	result := make([]store.CephClusterMDS, 0, len(records))
	for _, record := range records {
		key := record.Name
		if strings.TrimSpace(record.Name) == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, record)
	}
	return result
}

func mgrModuleRecords(clusterID uint, discoveredAt time.Time, modules map[string]any, enabled bool, key string) []store.CephClusterMgrModule {
	names := stringList(modules[key])
	records := make([]store.CephClusterMgrModule, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		records = append(records, store.CephClusterMgrModule{
			ClusterID:    clusterID,
			Name:         name,
			Enabled:      enabled,
			Payload:      mustJSON(map[string]any{"name": name, "enabled": enabled}),
			DiscoveredAt: discoveredAt,
		})
	}
	return records
}

func dashboardURLFromMgrServices(services map[string]any) string {
	for _, key := range []string{"dashboard", "dashboard_ssl"} {
		if value, ok := services[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimRight(strings.TrimSpace(value), "/")
		}
	}
	return ""
}

func commandHostsToAPIHosts(records []map[string]any) []ceph.Host {
	hosts := make([]ceph.Host, 0, len(records))
	for _, record := range records {
		hostname := stringField(record, "hostname")
		if hostname == "" {
			hostname = stringField(record, "host")
		}
		if hostname == "" {
			continue
		}
		hosts = append(hosts, ceph.Host{
			Hostname:    hostname,
			Addr:        stringField(record, "addr"),
			CephVersion: stringField(record, "ceph_version"),
			Status:      stringField(record, "status"),
			Labels:      stringSliceField(record, "labels"),
			Sources:     ceph.HostSources{Ceph: true, Orchestrator: true},
		})
	}
	return hosts
}

func nodeListToAPIHosts(nodes map[string]any) []ceph.Host {
	seen := map[string]bool{}
	var hosts []ceph.Host
	for _, value := range nodes {
		for _, hostname := range stringList(value) {
			if hostname == "" || seen[hostname] {
				continue
			}
			seen[hostname] = true
			hosts = append(hosts, ceph.Host{
				Hostname: hostname,
				Sources:  ceph.HostSources{Ceph: true},
			})
		}
	}
	return hosts
}

func osdTreeNodes(tree map[string]any) []map[string]any {
	values, ok := tree["nodes"].([]any)
	if !ok {
		return nil
	}
	osds := make([]map[string]any, 0, len(values))
	for _, value := range values {
		node, ok := value.(map[string]any)
		if !ok || stringField(node, "type") != "osd" {
			continue
		}
		osds = append(osds, node)
	}
	return osds
}

func osdFlagsFromDump(dump map[string]any) []string {
	flags := map[string]bool{}
	for _, key := range []string{"flags", "cluster_flags"} {
		if value, ok := dump[key]; ok {
			for _, flag := range stringList(value) {
				flags[flag] = true
			}
		}
	}
	result := make([]string, 0, len(flags))
	for flag := range flags {
		result = append(result, flag)
	}
	return result
}

func stringSliceField(record map[string]any, key string) []string {
	value, ok := record[key]
	if !ok {
		return nil
	}
	return stringList(value)
}

func boolField(record map[string]any, key string) bool {
	switch value := record[key].(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(value, "true") || value == "1" || strings.EqualFold(value, "yes")
	case float64:
		return value != 0
	case int:
		return value != 0
	default:
		return false
	}
}

func intField(record map[string]any, key string) int {
	return int(int64Field(record, key))
}

func int64Field(record map[string]any, key string) int64 {
	switch value := record[key].(type) {
	case int:
		return int64(value)
	case int64:
		return value
	case float64:
		return int64(value)
	case json.Number:
		number, _ := value.Int64()
		return number
	case string:
		var number int64
		_, _ = fmt.Sscanf(strings.TrimSpace(value), "%d", &number)
		return number
	default:
		return 0
	}
}

func firstInt64Field(record map[string]any, keys ...string) int64 {
	for _, key := range keys {
		if value := int64Field(record, key); value != 0 {
			return value
		}
	}
	return 0
}

func redactRGWSecrets(record map[string]any) map[string]any {
	copied := map[string]any{}
	for key, value := range record {
		if key == "keys" {
			copied[key] = redactedKeys(value)
			continue
		}
		copied[key] = value
	}
	return copied
}

func redactedKeys(value any) any {
	keys, ok := value.([]any)
	if !ok {
		return value
	}
	redacted := make([]any, 0, len(keys))
	for _, item := range keys {
		key, ok := item.(map[string]any)
		if !ok {
			redacted = append(redacted, item)
			continue
		}
		copied := map[string]any{}
		for name, field := range key {
			if strings.Contains(strings.ToLower(name), "secret") {
				copied[name] = ""
				continue
			}
			copied[name] = field
		}
		redacted = append(redacted, copied)
	}
	return redacted
}

func stringList(value any) []string {
	switch typed := value.(type) {
	case string:
		return strings.FieldsFunc(typed, func(r rune) bool {
			return r == ',' || r == ' '
		})
	case []string:
		return typed
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
				values = append(values, strings.TrimSpace(text))
			}
		}
		return values
	default:
		return nil
	}
}
