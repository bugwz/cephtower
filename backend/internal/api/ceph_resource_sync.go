package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"cephtower/backend/internal/integrations/ceph"
	"cephtower/backend/internal/integrations/ceph/dashboard"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	snapshotHosts         = "hosts"
	snapshotOSDs          = "osds"
	snapshotOSDFlags      = "osd_flags"
	snapshotDaemons       = "daemons"
	snapshotServices      = "services"
	snapshotMonitor       = "monitor"
	snapshotMgrModules    = "mgr_modules"
	snapshotConfiguration = "configuration"
)

type cephResourceSyncer struct {
	database func() *gorm.DB
}

func newCephResourceSyncer(database func() *gorm.DB) *cephResourceSyncer {
	return &cephResourceSyncer{database: database}
}

func (s *cephResourceSyncer) Sync(ctx context.Context) {
	db := s.database()
	if db == nil {
		return
	}

	var clusters []store.CephCluster
	if err := db.WithContext(ctx).
		Order("id asc").
		Find(&clusters).Error; err != nil {
		slog.Warn("list ceph clusters for resource sync", "error", err)
		return
	}

	for _, cluster := range clusters {
		if err := s.syncCluster(ctx, db, cluster); err != nil {
			slog.Warn("sync ceph resource snapshots", "cluster", cluster.Name, "error", err)
		}
	}
}

func (s *cephResourceSyncer) syncCluster(ctx context.Context, db *gorm.DB, cluster store.CephCluster) error {
	baseURL, err := dashboardBaseURLForCluster(ctx, &cluster)
	if err != nil {
		return err
	}
	client := dashboard.NewDashboardClient(dashboard.Config{
		BaseURL:     baseURL,
		Username:    cluster.DashboardUsername,
		Password:    cluster.DashboardPassword,
		InsecureTLS: false,
	})

	hostOptions := ceph.ListHostsOptions{IncludeServiceInstances: boolPtr(true)}
	if hosts, err := client.ListHosts(ctx, hostOptions); err != nil {
		s.recordError(ctx, db, cluster.ID, snapshotHosts, err)
	} else if err := saveSnapshot(ctx, db, cluster.ID, snapshotHosts, "all", hosts); err != nil {
		return err
	}

	if osds, err := client.ListOSDs(ctx, ceph.ListOSDsOptions{}); err != nil {
		s.recordError(ctx, db, cluster.ID, snapshotOSDs, err)
	} else if err := saveSnapshot(ctx, db, cluster.ID, snapshotOSDs, "all", osds); err != nil {
		return err
	}

	if flags, err := client.OSDFlags(ctx); err != nil {
		s.recordError(ctx, db, cluster.ID, snapshotOSDFlags, err)
	} else if err := saveSnapshot(ctx, db, cluster.ID, snapshotOSDFlags, "all", flags); err != nil {
		return err
	}

	if daemons, err := client.ListDaemons(ctx, ""); err != nil {
		s.recordError(ctx, db, cluster.ID, snapshotDaemons, err)
	} else if err := saveSnapshot(ctx, db, cluster.ID, snapshotDaemons, "all", daemons); err != nil {
		return err
	}

	if services, err := client.Raw(ctx, http.MethodGet, "/api/service", nil, nil); err != nil {
		s.recordError(ctx, db, cluster.ID, snapshotServices, err)
	} else if err := saveRawSnapshot(ctx, db, cluster.ID, snapshotServices, "all", services); err != nil {
		return err
	}

	if monitor, err := client.Raw(ctx, http.MethodGet, "/api/monitor", nil, nil); err != nil {
		s.recordError(ctx, db, cluster.ID, snapshotMonitor, err)
	} else if err := saveRawSnapshot(ctx, db, cluster.ID, snapshotMonitor, "all", monitor); err != nil {
		return err
	}

	if modules, err := client.Raw(ctx, http.MethodGet, "/api/mgr/module", nil, nil); err != nil {
		s.recordError(ctx, db, cluster.ID, snapshotMgrModules, err)
	} else if err := saveRawSnapshot(ctx, db, cluster.ID, snapshotMgrModules, "all", modules); err != nil {
		return err
	}

	if configuration, err := client.Raw(ctx, http.MethodGet, "/api/cluster_conf", nil, nil); err != nil {
		s.recordError(ctx, db, cluster.ID, snapshotConfiguration, err)
	} else if err := saveRawSnapshot(ctx, db, cluster.ID, snapshotConfiguration, "all", configuration); err != nil {
		return err
	}

	return nil
}

func (s *cephResourceSyncer) recordError(ctx context.Context, db *gorm.DB, clusterID uint, category string, err error) {
	snapshot := store.CephResourceSnapshot{
		ClusterID:    clusterID,
		Category:     category,
		ResourceKey:  "all",
		Payload:      "null",
		LastSyncedAt: time.Now(),
		LastError:    err.Error(),
	}
	_ = db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "cluster_id"}, {Name: "category"}, {Name: "resource_key"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"last_synced_at",
			"last_error",
			"updated_at",
		}),
	}).Create(&snapshot).Error
}

func saveSnapshot(ctx context.Context, db *gorm.DB, clusterID uint, category string, key string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return saveRawSnapshot(ctx, db, clusterID, category, key, data)
}

func saveRawSnapshot(ctx context.Context, db *gorm.DB, clusterID uint, category string, key string, payload []byte) error {
	snapshot := store.CephResourceSnapshot{
		ClusterID:    clusterID,
		Category:     category,
		ResourceKey:  key,
		Payload:      string(payload),
		LastSyncedAt: time.Now(),
		LastError:    "",
	}
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "cluster_id"}, {Name: "category"}, {Name: "resource_key"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"payload",
			"last_synced_at",
			"last_error",
			"updated_at",
		}),
	}).Create(&snapshot).Error
}

func boolPtr(value bool) *bool {
	return &value
}
