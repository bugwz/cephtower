package store

import (
	"context"
	"time"
)

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

type CephClusterSettingSnapshot struct {
	ID            uint        `gorm:"primaryKey"`
	ClusterID     uint        `gorm:"not null;uniqueIndex:idx_ceph_cluster_setting_snapshot"`
	Cluster       CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Name          string      `gorm:"size:256;not null;uniqueIndex:idx_ceph_cluster_setting_snapshot"`
	Group         string      `gorm:"size:128;index"`
	Type          string      `gorm:"size:128"`
	Default       bool        `gorm:"not null;default:false"`
	Sensitive     bool        `gorm:"not null;default:false"`
	ValueSet      bool        `gorm:"not null;default:false"`
	ValueRedacted string      `gorm:"type:longtext;not null"`
	Payload       string      `gorm:"type:longtext;not null"`
	DiscoveredAt  time.Time   `gorm:"not null;index"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (CephClusterSettingSnapshot) TableName() string {
	return "ceph_cluster_setting_snapshot"
}

type CephClusterFeatureToggle struct {
	ID           uint        `gorm:"primaryKey"`
	ClusterID    uint        `gorm:"not null;uniqueIndex:idx_ceph_cluster_feature_toggle"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Name         string      `gorm:"size:128;not null;uniqueIndex:idx_ceph_cluster_feature_toggle"`
	Enabled      bool        `gorm:"not null;default:false;index"`
	Payload      string      `gorm:"type:longtext;not null"`
	DiscoveredAt time.Time   `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (CephClusterFeatureToggle) TableName() string {
	return "ceph_cluster_feature_toggle"
}

type CephClusterIntegrationStatus struct {
	ID          uint        `gorm:"primaryKey"`
	ClusterID   uint        `gorm:"not null;uniqueIndex:idx_ceph_cluster_integration_status"`
	Cluster     CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Integration string      `gorm:"size:128;not null;uniqueIndex:idx_ceph_cluster_integration_status"`
	Configured  bool        `gorm:"not null;default:false;index"`
	Healthy     bool        `gorm:"not null;default:false;index"`
	Message     string      `gorm:"type:text;not null"`
	Payload     string      `gorm:"type:longtext;not null"`
	CheckedAt   time.Time   `gorm:"not null;index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (CephClusterIntegrationStatus) TableName() string {
	return "ceph_cluster_integration_status"
}

type CephClusterSettingChange struct {
	ID               uint        `gorm:"primaryKey"`
	ClusterID        uint        `gorm:"not null;index"`
	Cluster          CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	SettingName      string      `gorm:"size:256;not null;index"`
	OldValueRedacted string      `gorm:"type:longtext;not null"`
	NewValueRedacted string      `gorm:"type:longtext;not null"`
	OperatorUserID   uint        `gorm:"index"`
	Source           string      `gorm:"size:128;not null"`
	Status           string      `gorm:"size:32;not null;index"`
	Error            string      `gorm:"type:text;not null"`
	CreatedAt        time.Time
}

func (CephClusterSettingChange) TableName() string {
	return "ceph_cluster_setting_change"
}

func (d *Database) EnsureSettings(ctx context.Context, settings []Setting) error {
	for i := range settings {
		if err := d.db.WithContext(ctx).Where("`key` = ?", settings[i].Key).FirstOrCreate(&settings[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func (d *Database) ListSettings(ctx context.Context, prefix string) ([]Setting, error) {
	query := d.db.WithContext(ctx).Order("`key` asc")
	if prefix != "" {
		query = query.Where("`key` LIKE ?", prefix+"%")
	}
	var settings []Setting
	return settings, query.Find(&settings).Error
}

func (d *Database) UpsertSetting(ctx context.Context, key, value string) (Setting, error) {
	setting := Setting{Key: key, Value: value}
	err := d.db.WithContext(ctx).Where("`key` = ?", key).Assign(Setting{Value: value}).FirstOrCreate(&setting).Error
	return setting, err
}

func (d *Database) DeleteSettingsByPrefix(ctx context.Context, prefix string) error {
	return d.db.WithContext(ctx).Where("`key` LIKE ?", prefix+"%").Delete(&Setting{}).Error
}

func (d *Database) FindSetting(ctx context.Context, key string) (Setting, error) {
	var setting Setting
	err := d.db.WithContext(ctx).Where("`key` = ?", key).First(&setting).Error
	return setting, err
}

func (d *Database) RecordSettingChange(ctx context.Context, change *CephClusterSettingChange) error {
	return d.db.WithContext(ctx).Create(change).Error
}
