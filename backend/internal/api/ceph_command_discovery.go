package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"cephtower/backend/internal/integrations/ceph"
	"cephtower/backend/internal/integrations/ceph/command"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

const (
	defaultCephCommandBin            = "ceph"
	defaultCephCommandName           = "client.admin"
	defaultCephCommandTimeoutSeconds = 15
)

func discoverAndSyncCephCluster(ctx context.Context, db *gorm.DB, cluster *store.CephCluster) error {
	if cluster == nil {
		return nil
	}

	client, cleanup, err := commandClientForCluster(cluster)
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

	return nil
}

func commandClientForCluster(cluster *store.CephCluster) (*command.CommandClient, func(), error) {
	cleanup := func() {}
	file, err := os.CreateTemp("", "cephtower-ceph-keyring-*")
	if err != nil {
		return nil, cleanup, err
	}
	keyring := file.Name()
	cleanup = func() {
		_ = os.Remove(file.Name())
	}
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		cleanup()
		return nil, func() {}, err
	}
	if _, err := file.WriteString(keyringFileContent(defaultCephCommandName, cluster.Keyring)); err != nil {
		_ = file.Close()
		cleanup()
		return nil, func() {}, err
	}
	if err := file.Close(); err != nil {
		cleanup()
		return nil, func() {}, err
	}

	timeout := time.Duration(defaultCephCommandTimeoutSeconds) * time.Second
	client := command.NewCommandClient(command.Config{
		Bin:     defaultCephCommandBin,
		Cluster: "",
		Conf:    "",
		MonHost: cluster.MonitorHost,
		Name:    defaultCephCommandName,
		Keyring: keyring,
		Timeout: timeout,
	})
	return client, cleanup, nil
}

func dashboardBaseURLForCluster(ctx context.Context, cluster *store.CephCluster) (string, error) {
	client, cleanup, err := commandClientForCluster(cluster)
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
