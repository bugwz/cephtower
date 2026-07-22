package api

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"cephtower/backend/internal/integrations/ceph"
	"cephtower/backend/internal/integrations/ceph/command"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

func discoverAndSyncCephCluster(ctx context.Context, db *gorm.DB, cluster *store.CephCluster) error {
	if cluster == nil || !cluster.CommandEnabled {
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
	if strings.TrimSpace(fsid) != "" {
		cluster.FSID = strings.TrimSpace(fsid)
	}

	services, err := client.MgrServices(ctx)
	if err != nil {
		return fmt.Errorf("discover mgr services: %w", err)
	}
	dashboardURL := dashboardURLFromMgrServices(services)
	if cluster.DashboardEnabled {
		if dashboardURL == "" && strings.TrimSpace(cluster.DashboardBaseURL) == "" {
			return fmt.Errorf("dashboard service address was not found from ceph mgr services")
		}
		if dashboardURL != "" {
			cluster.DashboardBaseURL = dashboardURL
		}
	}

	if err := db.WithContext(ctx).Save(cluster).Error; err != nil {
		return err
	}

	saveDiscoveredSnapshot(ctx, db, cluster.ID, snapshotHosts, func() (any, error) {
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
	saveDiscoveredSnapshot(ctx, db, cluster.ID, snapshotOSDs, func() (any, error) {
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
	saveDiscoveredSnapshot(ctx, db, cluster.ID, snapshotOSDFlags, func() (any, error) {
		dump, err := client.OSDDump(ctx)
		if err != nil {
			return nil, err
		}
		return osdFlagsFromDump(dump), nil
	})
	saveDiscoveredSnapshot(ctx, db, cluster.ID, snapshotDaemons, func() (any, error) {
		return client.OrchPS(ctx, command.OrchPSOptions{})
	})
	saveDiscoveredSnapshot(ctx, db, cluster.ID, snapshotServices, func() (any, error) {
		return client.OrchList(ctx, command.OrchListOptions{})
	})
	saveDiscoveredSnapshot(ctx, db, cluster.ID, snapshotMonitor, func() (any, error) {
		return client.MonDump(ctx)
	})
	saveDiscoveredSnapshot(ctx, db, cluster.ID, snapshotMgrModules, func() (any, error) {
		return client.MgrModuleList(ctx)
	})
	saveDiscoveredSnapshot(ctx, db, cluster.ID, snapshotConfiguration, func() (any, error) {
		return client.ConfigDump(ctx)
	})

	return nil
}

func commandClientForCluster(cluster *store.CephCluster) (*command.CommandClient, func(), error) {
	keyring := strings.TrimSpace(cluster.CommandKeyring)
	cleanup := func() {}
	if strings.TrimSpace(cluster.CommandKeyringContent) != "" {
		file, err := os.CreateTemp("", "cephtower-ceph-keyring-*")
		if err != nil {
			return nil, cleanup, err
		}
		keyring = file.Name()
		cleanup = func() {
			_ = os.Remove(file.Name())
		}
		if err := file.Chmod(0o600); err != nil {
			_ = file.Close()
			cleanup()
			return nil, func() {}, err
		}
		if _, err := file.WriteString(keyringFileContent(cluster.CommandName, cluster.CommandKeyringContent)); err != nil {
			_ = file.Close()
			cleanup()
			return nil, func() {}, err
		}
		if err := file.Close(); err != nil {
			cleanup()
			return nil, func() {}, err
		}
	}

	timeout := time.Duration(cluster.CommandTimeoutSeconds) * time.Second
	client := command.NewCommandClient(command.Config{
		Bin:     cluster.CommandBin,
		Cluster: cluster.CommandCluster,
		Conf:    cluster.CommandConf,
		Name:    cluster.CommandName,
		Keyring: keyring,
		Timeout: timeout,
	})
	return client, cleanup, nil
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

func saveDiscoveredSnapshot(ctx context.Context, db *gorm.DB, clusterID uint, category string, load func() (any, error)) {
	payload, err := load()
	if err != nil {
		newCephResourceSyncer(nil).recordError(ctx, db, clusterID, category, err)
		return
	}
	_ = saveSnapshot(ctx, db, clusterID, category, "all", payload)
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
