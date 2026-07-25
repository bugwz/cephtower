package store

import "gorm.io/gorm"

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Setting{},
		&CephCluster{},
		&CephClusterHost{},
		&CephClusterOSD{},
		&CephClusterOSDFlag{},
		&CephClusterDaemon{},
		&CephClusterService{},
		&CephClusterMon{},
		&CephClusterMgr{},
		&CephClusterMDS{},
		&CephClusterMgrModule{},
		&CephClusterConfiguration{},
		&CephDataFetchRun{},
		&CephClusterSummary{},
		&CephClusterHealthCheck{},
		&CephPool{},
		&CephRBDImage{},
		&CephFilesystem{},
		&CephRGWDaemon{},
		&CephRGWUser{},
		&CephRGWBucket{},
		&CephNVMeoFGateway{},
		&CephNVMeoFSubsystem{},
		&CephISCSITarget{},
		&CephNFSExport{},
		&CephClusterSettingSnapshot{},
		&CephClusterFeatureToggle{},
		&CephClusterIntegrationStatus{},
		&CephClusterSettingChange{},
		&User{},
		&PasswordResetCode{},
		&UserSession{},
	)
}
