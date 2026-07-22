package ceph

import (
	"context"

	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

func DeleteCephClusterResources(ctx context.Context, db *gorm.DB, clusterID uint) error {
	models := []any{
		&store.CephClusterHost{},
		&store.CephClusterOSD{},
		&store.CephClusterOSDFlag{},
		&store.CephClusterDaemon{},
		&store.CephClusterService{},
		&store.CephClusterMon{},
		&store.CephClusterMgr{},
		&store.CephClusterMDS{},
		&store.CephClusterMgrModule{},
		&store.CephClusterConfiguration{},
		&store.CephDataFetchRun{},
		&store.CephClusterSummary{},
		&store.CephClusterHealthCheck{},
		&store.CephPool{},
		&store.CephRBDImage{},
		&store.CephFilesystem{},
		&store.CephRGWDaemon{},
		&store.CephRGWUser{},
		&store.CephRGWBucket{},
		&store.CephNVMeoFGateway{},
		&store.CephNVMeoFSubsystem{},
		&store.CephISCSITarget{},
		&store.CephNFSExport{},
	}
	for _, model := range models {
		if err := db.WithContext(ctx).Where("cluster_id = ?", clusterID).Delete(model).Error; err != nil {
			return err
		}
	}
	return nil
}
