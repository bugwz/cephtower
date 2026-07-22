package store

import "time"

const (
	UserRoleAdmin = "admin"
	UserRoleUser  = "user"
)

type Setting struct {
	Key       string `gorm:"primaryKey;size:128"`
	Value     string `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Setting) TableName() string {
	return "setting"
}

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

type CephResourceSnapshot struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_resource_snapshot"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Category     string      `gorm:"size:64;not null;uniqueIndex:idx_ceph_resource_snapshot"`
	ResourceKey  string      `gorm:"size:256;not null;uniqueIndex:idx_ceph_resource_snapshot"`
	Payload      string      `gorm:"type:longtext;not null"`
	LastSyncedAt time.Time   `gorm:"not null;index"`
	LastError    string      `gorm:"type:text;not null;default:''"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephResourceSnapshot) TableName() string {
	return "ceph_resource_snapshot"
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

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex;size:64;not null"`
	DisplayName  string `gorm:"size:96;not null"`
	Email        string `gorm:"uniqueIndex;size:128"`
	Role         string `gorm:"size:24;not null;index"`
	Permissions  string `gorm:"type:text;not null"`
	PasswordHash string `gorm:"type:text;not null"`
	Enabled      bool   `gorm:"not null;default:true"`
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (User) TableName() string {
	return "user"
}

type PasswordResetCode struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	User      User      `gorm:"constraint:OnDelete:CASCADE"`
	CodeHash  string    `gorm:"type:text;not null"`
	Used      bool      `gorm:"not null;default:false"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (PasswordResetCode) TableName() string {
	return "password_reset_code"
}

type UserSession struct {
	ID        uint      `gorm:"primaryKey"`
	Token     string    `gorm:"uniqueIndex;size:96;not null"`
	UserID    uint      `gorm:"not null;index"`
	User      User      `gorm:"constraint:OnDelete:CASCADE"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (UserSession) TableName() string {
	return "user_session"
}
