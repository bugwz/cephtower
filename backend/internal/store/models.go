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

type CephCluster struct {
	ID                    uint   `gorm:"primaryKey"`
	Name                  string `gorm:"uniqueIndex;size:128;not null"`
	Description           string `gorm:"type:text;not null;default:''"`
	FSID                  string `gorm:"size:64;index"`
	Enabled               bool   `gorm:"not null;default:true;index"`
	DashboardEnabled      bool   `gorm:"not null;default:false"`
	DashboardBaseURL      string `gorm:"size:512"`
	DashboardUsername     string `gorm:"size:128"`
	DashboardPassword     string `gorm:"type:text"`
	DashboardInsecureTLS  bool   `gorm:"not null;default:false"`
	CommandEnabled        bool   `gorm:"not null;default:false"`
	CommandBin            string `gorm:"size:256;not null;default:'ceph'"`
	CommandCluster        string `gorm:"size:128"`
	CommandConf           string `gorm:"size:512"`
	CommandName           string `gorm:"size:128"`
	CommandKeyring        string `gorm:"size:512"`
	CommandKeyringContent string `gorm:"type:text"`
	CommandTimeoutSeconds int    `gorm:"not null;default:15"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
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

type UserSession struct {
	ID        uint      `gorm:"primaryKey"`
	Token     string    `gorm:"uniqueIndex;size:96;not null"`
	UserID    uint      `gorm:"not null;index"`
	User      User      `gorm:"constraint:OnDelete:CASCADE"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
