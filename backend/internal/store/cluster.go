package store

import (
	"context"
	"time"
)

type CephCluster struct {
	ID                uint   `gorm:"primaryKey"`
	Name              string `gorm:"uniqueIndex;size:128;not null"`
	MonitorHost       string `gorm:"type:text;not null"`
	Keyring           string `gorm:"type:text;not null"`
	DashboardUsername string `gorm:"size:128;not null"`
	DashboardPassword string `gorm:"type:text;not null"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (CephCluster) TableName() string {
	return "ceph_cluster"
}

func (d *Database) ListClusters(ctx context.Context) ([]CephCluster, error) {
	var clusters []CephCluster
	err := d.db.WithContext(ctx).Order("id asc").Find(&clusters).Error
	return clusters, err
}

func (d *Database) FindCluster(ctx context.Context, id uint) (CephCluster, error) {
	var cluster CephCluster
	err := d.db.WithContext(ctx).First(&cluster, id).Error
	return cluster, err
}

func (d *Database) FirstCluster(ctx context.Context) (CephCluster, error) {
	var cluster CephCluster
	err := d.db.WithContext(ctx).Order("id asc").First(&cluster).Error
	return cluster, err
}

func (d *Database) CreateCluster(ctx context.Context, cluster *CephCluster) error {
	return d.db.WithContext(ctx).Create(cluster).Error
}
func (d *Database) SaveCluster(ctx context.Context, cluster *CephCluster) error {
	return d.db.WithContext(ctx).Save(cluster).Error
}
func (d *Database) DeleteCluster(ctx context.Context, cluster *CephCluster) error {
	return d.db.WithContext(ctx).Delete(cluster).Error
}

func (d *Database) DeleteClusterResources(ctx context.Context, clusterID uint) error {
	models := []any{&CephClusterHost{}, &CephClusterOSD{}, &CephClusterOSDFlag{}, &CephClusterDaemon{}, &CephClusterService{}, &CephClusterMon{}, &CephClusterMgr{}, &CephClusterMDS{}, &CephClusterMgrModule{}, &CephClusterConfiguration{}, &CephDataFetchRun{}, &CephClusterSummary{}, &CephClusterHealthCheck{}, &CephPool{}, &CephRBDImage{}, &CephFilesystem{}, &CephRGWDaemon{}, &CephRGWUser{}, &CephRGWBucket{}, &CephNVMeoFGateway{}, &CephNVMeoFSubsystem{}, &CephISCSITarget{}, &CephNFSExport{}, &CephClusterSettingSnapshot{}, &CephClusterFeatureToggle{}, &CephClusterIntegrationStatus{}, &CephClusterSettingChange{}}
	for _, model := range models {
		if err := d.db.WithContext(ctx).Where("cluster_id = ?", clusterID).Delete(model).Error; err != nil {
			return err
		}
	}
	return nil
}

type ClusterDiscovery struct {
	Hosts         []CephClusterHost
	OSDs          []CephClusterOSD
	OSDFlags      []CephClusterOSDFlag
	Daemons       []CephClusterDaemon
	Services      []CephClusterService
	Mons          []CephClusterMon
	Mgrs          []CephClusterMgr
	MDSs          []CephClusterMDS
	MgrModules    []CephClusterMgrModule
	Configuration []CephClusterConfiguration
}

func (d *Database) LoadClusterDiscovery(ctx context.Context, clusterID uint) (ClusterDiscovery, error) {
	result := ClusterDiscovery{}
	queries := []struct {
		order string
		out   any
	}{{"hostname asc", &result.Hosts}, {"osd_id asc", &result.OSDs}, {"name asc", &result.OSDFlags}, {"name asc", &result.Daemons}, {"service_name asc", &result.Services}, {"name asc", &result.Mons}, {"name asc", &result.Mgrs}, {"filesystem asc, name asc", &result.MDSs}, {"name asc", &result.MgrModules}, {"who asc, name asc", &result.Configuration}}
	for _, query := range queries {
		if err := d.db.WithContext(ctx).Where("cluster_id = ?", clusterID).Order(query.order).Find(query.out).Error; err != nil {
			return ClusterDiscovery{}, err
		}
	}
	return result, nil
}

type CephClusterHost struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_cluster_host"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Hostname     string      `gorm:"size:256;not null;uniqueIndex:idx_ceph_cluster_host"`
	Addr         string      `gorm:"size:128"`
	CephVersion  string      `gorm:"size:128"`
	Status       string      `gorm:"size:64"`
	Labels       string      `gorm:"type:longtext;not null"`
	Sources      string      `gorm:"type:longtext;not null"`
	Payload      string      `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time   `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephClusterHost) TableName() string {
	return "ceph_cluster_host"
}

type CephClusterOSD struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_cluster_osd"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	OSDID        string      `gorm:"size:64;not null;uniqueIndex:idx_ceph_cluster_osd"`
	Hostname     string      `gorm:"size:256"`
	Status       string      `gorm:"size:64"`
	Payload      string      `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time   `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephClusterOSD) TableName() string {
	return "ceph_cluster_osd"
}

type CephClusterOSDFlag struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_cluster_osd_flag"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Name         string      `gorm:"size:128;not null;uniqueIndex:idx_ceph_cluster_osd_flag"`
	DiscoveredAt time.Time   `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephClusterOSDFlag) TableName() string {
	return "ceph_cluster_osd_flag"
}

type CephClusterDaemon struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_cluster_daemon"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Name         string      `gorm:"size:256;not null;uniqueIndex:idx_ceph_cluster_daemon"`
	DaemonType   string      `gorm:"size:64;index"`
	Hostname     string      `gorm:"size:256;index"`
	Status       string      `gorm:"size:64"`
	Payload      string      `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time   `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephClusterDaemon) TableName() string {
	return "ceph_cluster_daemon"
}

type CephClusterService struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_cluster_service"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	ServiceName  string      `gorm:"size:256;not null;uniqueIndex:idx_ceph_cluster_service"`
	ServiceType  string      `gorm:"size:64;index"`
	Payload      string      `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time   `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephClusterService) TableName() string {
	return "ceph_cluster_service"
}

type CephClusterMon struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_cluster_mon"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Name         string      `gorm:"size:128;not null;uniqueIndex:idx_ceph_cluster_mon"`
	Rank         string      `gorm:"size:64"`
	Addr         string      `gorm:"size:256"`
	PublicAddr   string      `gorm:"size:256"`
	Status       string      `gorm:"size:64;index"`
	Payload      string      `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time   `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephClusterMon) TableName() string {
	return "ceph_cluster_mon"
}

type CephClusterMgr struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_cluster_mgr"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Name         string      `gorm:"size:128;not null;uniqueIndex:idx_ceph_cluster_mgr"`
	Addr         string      `gorm:"size:256"`
	Hostname     string      `gorm:"size:256;index"`
	Status       string      `gorm:"size:64;index"`
	Payload      string      `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time   `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephClusterMgr) TableName() string {
	return "ceph_cluster_mgr"
}

type CephClusterMDS struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_cluster_mds"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Name         string      `gorm:"size:128;not null;uniqueIndex:idx_ceph_cluster_mds"`
	Filesystem   string      `gorm:"size:128;index"`
	Rank         string      `gorm:"size:64"`
	GID          string      `gorm:"size:64"`
	Addr         string      `gorm:"size:256"`
	Hostname     string      `gorm:"size:256;index"`
	State        string      `gorm:"size:64;index"`
	Payload      string      `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time   `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephClusterMDS) TableName() string {
	return "ceph_cluster_mds"
}

type CephClusterMgrModule struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_cluster_mgr_module"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Name         string      `gorm:"size:128;not null;uniqueIndex:idx_ceph_cluster_mgr_module"`
	Enabled      bool        `gorm:"not null;default:false;index"`
	Payload      string      `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time   `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephClusterMgrModule) TableName() string {
	return "ceph_cluster_mgr_module"
}

type CephClusterConfiguration struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_cluster_configuration"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Who          string      `gorm:"size:128;uniqueIndex:idx_ceph_cluster_configuration"`
	Name         string      `gorm:"size:256;not null;uniqueIndex:idx_ceph_cluster_configuration"`
	Level        string      `gorm:"size:64"`
	Value        string      `gorm:"type:text"`
	Payload      string      `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time   `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephClusterConfiguration) TableName() string {
	return "ceph_cluster_configuration"
}

type CephClusterSummary struct {
	ID                uint        `gorm:"primaryKey"`
	ClusterID         uint        `gorm:"not null;uniqueIndex"`
	Cluster           CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	HealthStatus      string      `gorm:"size:64;index"`
	Version           string      `gorm:"size:128"`
	MgrID             string      `gorm:"size:128"`
	MgrHost           string      `gorm:"size:256"`
	HaveMonConnection bool        `gorm:"not null;default:false"`
	ExecutingTasks    string      `gorm:"type:longtext;not null"`
	FinishedTasks     string      `gorm:"type:longtext;not null"`
	Payload           string      `gorm:"type:longtext;not null"`
	DiscoveredAt      time.Time   `gorm:"not null;index"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (CephClusterSummary) TableName() string {
	return "ceph_cluster_summary"
}

type CephClusterHealthCheck struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_cluster_health_check"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Name         string      `gorm:"size:256;not null;uniqueIndex:idx_ceph_cluster_health_check"`
	Severity     string      `gorm:"size:64;index"`
	Summary      string      `gorm:"type:text"`
	Detail       string      `gorm:"type:longtext;not null"`
	Muted        bool        `gorm:"not null;default:false"`
	Count        int
	Payload      string    `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephClusterHealthCheck) TableName() string {
	return "ceph_cluster_health_check"
}

type CephPool struct {
	ID                  uint        `gorm:"primaryKey"`
	ClusterID           uint        `gorm:"not null;uniqueIndex:idx_ceph_pool"`
	Cluster             CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	PoolID              string      `gorm:"size:64;index"`
	PoolName            string      `gorm:"size:256;not null;uniqueIndex:idx_ceph_pool"`
	Type                string      `gorm:"size:64;index"`
	Size                int
	MinSize             int
	PGNum               int
	PGPlacementNum      int
	PGAutoscaleMode     string `gorm:"size:64"`
	CrushRule           string `gorm:"size:128"`
	ErasureCodeProfile  string `gorm:"size:128"`
	ApplicationMetadata string `gorm:"type:longtext;not null"`
	QuotaMaxBytes       int64
	QuotaMaxObjects     int64
	UsedBytes           int64
	MaxAvailBytes       int64
	Objects             int64
	Payload             string    `gorm:"type:longtext;not null"`
	DiscoveredAt        time.Time `gorm:"not null;index"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (CephPool) TableName() string {
	return "ceph_pool"
}

type CephRBDImage struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_rbd_image"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	PoolName     string      `gorm:"size:256;not null;index"`
	Namespace    string      `gorm:"size:256"`
	ImageName    string      `gorm:"size:256;not null"`
	ImageSpec    string      `gorm:"size:768;not null;uniqueIndex:idx_ceph_rbd_image"`
	ImageID      string      `gorm:"size:128;index"`
	SizeBytes    int64
	ObjectSize   int
	Features     string `gorm:"type:longtext;not null"`
	StripeCount  int
	StripeUnit   int64
	Parent       string    `gorm:"type:longtext;not null"`
	Snapshots    string    `gorm:"type:longtext;not null"`
	MirrorMode   string    `gorm:"size:64"`
	Trash        bool      `gorm:"not null;default:false;index"`
	Payload      string    `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephRBDImage) TableName() string {
	return "ceph_rbd_image"
}

type CephFilesystem struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_filesystem"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	FSID         string      `gorm:"size:64;not null;uniqueIndex:idx_ceph_filesystem"`
	Name         string      `gorm:"size:256;not null;index"`
	MetadataPool string      `gorm:"size:256"`
	DataPools    string      `gorm:"type:longtext;not null"`
	MDSMap       string      `gorm:"type:longtext;not null"`
	StandbyCount int
	ClientCount  int
	UsedBytes    int64
	AvailBytes   int64
	Payload      string    `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephFilesystem) TableName() string {
	return "ceph_filesystem"
}

type CephRGWDaemon struct {
	ID             uint        `gorm:"primaryKey"`
	ClusterID      uint        `gorm:"not null;uniqueIndex:idx_ceph_rgw_daemon"`
	Cluster        CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	ServiceID      string      `gorm:"size:256;not null;uniqueIndex:idx_ceph_rgw_daemon"`
	Hostname       string      `gorm:"size:256;index"`
	ZoneName       string      `gorm:"size:256;index"`
	FrontendConfig string      `gorm:"type:text"`
	Version        string      `gorm:"size:128"`
	Payload        string      `gorm:"type:longtext;not null"`
	DiscoveredAt   time.Time   `gorm:"not null;index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (CephRGWDaemon) TableName() string {
	return "ceph_rgw_daemon"
}

type CephRGWUser struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_rgw_user"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	UID          string      `gorm:"size:256;not null;uniqueIndex:idx_ceph_rgw_user"`
	DisplayName  string      `gorm:"size:256"`
	Email        string      `gorm:"size:256;index"`
	Suspended    bool        `gorm:"not null;default:false;index"`
	MaxBuckets   int
	Subusers     string    `gorm:"type:longtext;not null"`
	KeysRedacted string    `gorm:"type:longtext;not null"`
	Caps         string    `gorm:"type:longtext;not null"`
	Quota        string    `gorm:"type:longtext;not null"`
	Stats        string    `gorm:"type:longtext;not null"`
	Payload      string    `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephRGWUser) TableName() string {
	return "ceph_rgw_user"
}

type CephRGWBucket struct {
	ID            uint        `gorm:"primaryKey"`
	ClusterID     uint        `gorm:"not null;uniqueIndex:idx_ceph_rgw_bucket"`
	Cluster       CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Tenant        string      `gorm:"size:256;uniqueIndex:idx_ceph_rgw_bucket"`
	Bucket        string      `gorm:"size:256;not null;uniqueIndex:idx_ceph_rgw_bucket"`
	Owner         string      `gorm:"size:256;index"`
	Zonegroup     string      `gorm:"size:256"`
	PlacementRule string      `gorm:"size:256"`
	Versioning    string      `gorm:"size:64"`
	ObjectCount   int64
	UsedBytes     int64
	Quota         string    `gorm:"type:longtext;not null"`
	Lifecycle     string    `gorm:"type:longtext;not null"`
	Encryption    string    `gorm:"type:longtext;not null"`
	Payload       string    `gorm:"type:longtext;not null"`
	DiscoveredAt  time.Time `gorm:"not null;index"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (CephRGWBucket) TableName() string {
	return "ceph_rgw_bucket"
}

type CephNVMeoFGateway struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_nvmeof_gateway"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	GroupName    string      `gorm:"size:256;not null;uniqueIndex:idx_ceph_nvmeof_gateway"`
	Hostname     string      `gorm:"size:256;not null;uniqueIndex:idx_ceph_nvmeof_gateway"`
	TRAddr       string      `gorm:"size:256;not null;uniqueIndex:idx_ceph_nvmeof_gateway"`
	Status       string      `gorm:"size:64;index"`
	Version      string      `gorm:"size:128"`
	Listeners    string      `gorm:"type:longtext;not null"`
	Stats        string      `gorm:"type:longtext;not null"`
	Payload      string      `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time   `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephNVMeoFGateway) TableName() string {
	return "ceph_nvmeof_gateway"
}

type CephNVMeoFSubsystem struct {
	ID            uint        `gorm:"primaryKey"`
	ClusterID     uint        `gorm:"not null;uniqueIndex:idx_ceph_nvmeof_subsystem"`
	Cluster       CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	NQN           string      `gorm:"size:512;not null;uniqueIndex:idx_ceph_nvmeof_subsystem"`
	SerialNumber  string      `gorm:"size:128"`
	ModelNumber   string      `gorm:"size:128"`
	MaxNamespaces int
	Namespaces    string    `gorm:"type:longtext;not null"`
	Hosts         string    `gorm:"type:longtext;not null"`
	Listeners     string    `gorm:"type:longtext;not null"`
	Connections   string    `gorm:"type:longtext;not null"`
	Payload       string    `gorm:"type:longtext;not null"`
	DiscoveredAt  time.Time `gorm:"not null;index"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (CephNVMeoFSubsystem) TableName() string {
	return "ceph_nvmeof_subsystem"
}

type CephISCSITarget struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_iscsi_target"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	TargetIQN    string      `gorm:"size:512;not null;uniqueIndex:idx_ceph_iscsi_target"`
	Portals      string      `gorm:"type:longtext;not null"`
	Disks        string      `gorm:"type:longtext;not null"`
	Clients      string      `gorm:"type:longtext;not null"`
	Groups       string      `gorm:"type:longtext;not null"`
	Payload      string      `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time   `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephISCSITarget) TableName() string {
	return "ceph_iscsi_target"
}

type CephNFSExport struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_nfs_export"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	NFSClusterID string      `gorm:"size:256;not null;uniqueIndex:idx_ceph_nfs_export"`
	ExportID     string      `gorm:"size:128;not null;uniqueIndex:idx_ceph_nfs_export"`
	Path         string      `gorm:"size:1024"`
	Pseudo       string      `gorm:"size:1024"`
	AccessType   string      `gorm:"size:64"`
	Squash       string      `gorm:"size:128"`
	Protocols    string      `gorm:"type:longtext;not null"`
	Transports   string      `gorm:"type:longtext;not null"`
	FSAL         string      `gorm:"type:longtext;not null"`
	Payload      string      `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time   `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephNFSExport) TableName() string {
	return "ceph_nfs_export"
}
