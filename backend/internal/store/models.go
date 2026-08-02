package store

import "time"

type SchemaMigration struct {
	Version   string    `gorm:"primaryKey;size:64"`
	Checksum  string    `gorm:"size:64;not null"`
	AppliedAt time.Time `gorm:"not null"`
}

func (SchemaMigration) TableName() string { return "schema_migration" }

type Setting struct {
	Key       string    `gorm:"primaryKey;size:191"`
	Value     string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Setting) TableName() string { return "setting" }

type User struct {
	ID          uint64  `gorm:"primaryKey;autoIncrement"`
	Username    string  `gorm:"size:128;not null;uniqueIndex:uq_user_username"`
	DisplayName string  `gorm:"size:255;not null;default:''"`
	Email       *string `gorm:"size:320;index:idx_user_email"`
	Password    string  `gorm:"type:text;not null"`
	Status      string  `gorm:"size:32;not null;default:active;index:idx_user_status"`
	LastLoginAt *time.Time
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (User) TableName() string { return "user" }

type PasswordResetCode struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement"`
	UserID     uint64     `gorm:"not null;index:idx_reset_user_expiry,priority:1"`
	User       User       `gorm:"constraint:OnDelete:CASCADE"`
	CodeHash   string     `gorm:"size:64;not null;uniqueIndex:uq_reset_code_hash"`
	ExpiresAt  time.Time  `gorm:"not null;index:idx_reset_user_expiry,priority:2;index:idx_reset_expiry_consumed,priority:1"`
	ConsumedAt *time.Time `gorm:"index:idx_reset_expiry_consumed,priority:2"`
	CreatedAt  time.Time  `gorm:"not null"`
}

func (PasswordResetCode) TableName() string { return "password_reset_code" }

type UserSession struct {
	ID         string     `gorm:"primaryKey;size:36"`
	UserID     uint64     `gorm:"not null;index:idx_session_user_expiry,priority:1"`
	User       User       `gorm:"constraint:OnDelete:CASCADE"`
	TokenHash  string     `gorm:"size:64;not null;uniqueIndex:uq_session_token_hash"`
	SourceIP   *string    `gorm:"size:64"`
	UserAgent  *string    `gorm:"size:512"`
	ExpiresAt  time.Time  `gorm:"not null;index:idx_session_user_expiry,priority:2;index:idx_session_expiry_revoked,priority:1"`
	LastSeenAt time.Time  `gorm:"not null"`
	RevokedAt  *time.Time `gorm:"index:idx_session_expiry_revoked,priority:2"`
	CreatedAt  time.Time  `gorm:"not null"`
}

func (UserSession) TableName() string { return "user_session" }

type Role struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"`
	Name        string    `gorm:"size:128;not null;uniqueIndex:uq_role_name"`
	Description *string   `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (Role) TableName() string { return "role" }

type CephCluster struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement"`
	Name             string     `gorm:"size:128;not null;uniqueIndex:uq_cluster_name"`
	MonitorAddresses string     `gorm:"size:4096;not null"`
	ClientUsername   string     `gorm:"size:128;not null"`
	ClientKey        string     `gorm:"type:text;not null"`
	DiscoveredData   string     `json:"-" gorm:"column:discovered_data;type:text;not null;default:'{}'"`
	FSID             *string    `gorm:"column:fsid;size:36;index:idx_cluster_fsid"`
	CephVersion      *string    `gorm:"size:128"`
	Status           string     `gorm:"size:32;not null;default:unknown;index:idx_cluster_status_enabled,priority:1"`
	Enabled          bool       `gorm:"not null;default:true;index:idx_cluster_status_enabled,priority:2"`
	Generation       uint64     `gorm:"not null;default:0"`
	LastSeenAt       *time.Time `gorm:"index:idx_cluster_last_seen"`
	LastErrorCode    *string    `gorm:"size:64"`
	LastErrorMessage *string    `gorm:"type:text"`
	ObservedAt       *time.Time
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`
}

func (CephCluster) TableName() string { return "ceph_cluster" }

type UserRoleBinding struct {
	ID              uint64       `gorm:"primaryKey;autoIncrement"`
	UserID          uint64       `gorm:"not null;uniqueIndex:uq_user_role_scope,priority:1;index:idx_binding_user_cluster,priority:1"`
	User            User         `gorm:"constraint:OnDelete:CASCADE"`
	RoleID          uint64       `gorm:"not null;uniqueIndex:uq_user_role_scope,priority:2;index:idx_binding_role"`
	Role            Role         `gorm:"constraint:OnDelete:CASCADE"`
	ClusterID       *uint64      `gorm:"index:idx_binding_user_cluster,priority:2"`
	Cluster         *CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	ScopeKey        string       `gorm:"size:32;not null;uniqueIndex:uq_user_role_scope,priority:3"`
	CreatedByUserID *uint64      `gorm:"index:idx_binding_created_by"`
	CreatedBy       *User        `gorm:"foreignKey:CreatedByUserID;constraint:OnDelete:SET NULL"`
	CreatedAt       time.Time    `gorm:"not null"`
}

func (UserRoleBinding) TableName() string { return "user_role_binding" }

type CephClusterCredential struct {
	ID          uint64      `gorm:"primaryKey;autoIncrement"`
	ClusterID   uint64      `gorm:"not null;uniqueIndex:uq_cluster_credential,priority:1"`
	Cluster     CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Kind        string      `gorm:"size:64;not null;uniqueIndex:uq_cluster_credential,priority:2"`
	Credential  string      `gorm:"type:text;not null"`
	Fingerprint string      `gorm:"size:64;not null"`
	CreatedAt   time.Time   `gorm:"not null"`
	UpdatedAt   time.Time   `gorm:"not null"`
}

func (CephClusterCredential) TableName() string { return "ceph_cluster_credential" }

type CephClusterEndpoint struct {
	ID             uint64                 `gorm:"primaryKey;autoIncrement"`
	ClusterID      uint64                 `gorm:"not null;uniqueIndex:uq_cluster_endpoint,priority:1;index:idx_endpoint_enabled,priority:1"`
	Cluster        CephCluster            `gorm:"constraint:OnDelete:CASCADE"`
	Kind           string                 `gorm:"size:64;not null;uniqueIndex:uq_cluster_endpoint,priority:2"`
	Name           string                 `gorm:"size:128;not null;default:default;uniqueIndex:uq_cluster_endpoint,priority:3"`
	URL            string                 `gorm:"size:2048;not null"`
	TLSMode        string                 `gorm:"size:32;not null"`
	CACredentialID *uint64                `gorm:"index:idx_endpoint_ca"`
	CACredential   *CephClusterCredential `gorm:"foreignKey:CACredentialID;constraint:OnDelete:SET NULL"`
	ConfigJSON     *string                `gorm:"type:text"`
	Enabled        bool                   `gorm:"not null;default:true;check:endpoint_enabled_check,enabled IN (0,1);index:idx_endpoint_enabled,priority:2"`
	CreatedAt      time.Time              `gorm:"not null"`
	UpdatedAt      time.Time              `gorm:"not null"`
}

func (CephClusterEndpoint) TableName() string { return "ceph_cluster_endpoint" }

type CephClusterCapability struct {
	ID          uint64      `gorm:"primaryKey;autoIncrement"`
	ClusterID   uint64      `gorm:"not null;uniqueIndex:uq_cluster_capability,priority:1;index:idx_capability_supported,priority:1"`
	Cluster     CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Name        string      `gorm:"size:128;not null;uniqueIndex:uq_cluster_capability,priority:2"`
	Supported   bool        `gorm:"not null;check:capability_supported_check,supported IN (0,1);index:idx_capability_supported,priority:2"`
	Reason      *string     `gorm:"size:64"`
	Version     *string     `gorm:"size:128"`
	DetailsJSON *string     `gorm:"type:text"`
	ObservedAt  time.Time   `gorm:"not null;index:idx_capability_observed"`
	UpdatedAt   time.Time   `gorm:"not null"`
}

func (CephClusterCapability) TableName() string { return "ceph_cluster_capability" }

type CephHost struct {
	ID                     uint64      `gorm:"primaryKey;autoIncrement"`
	ClusterID              uint64      `gorm:"not null;uniqueIndex:uq_ceph_host,priority:1;index:idx_ceph_host_cluster,priority:1"`
	Cluster                CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Hostname               string      `gorm:"size:512;not null;uniqueIndex:uq_ceph_host,priority:2;index:idx_ceph_host_cluster,priority:2"`
	SSHAddress             string      `gorm:"column:ssh_address;size:255;not null"`
	SSHPort                uint16      `gorm:"column:ssh_port;not null;default:22"`
	SSHUser                string      `gorm:"column:ssh_user;size:128;not null"`
	SSHAuthMethod          string      `gorm:"column:ssh_auth_method;size:32;not null"`
	SSHPasswordSecret      *string     `gorm:"column:ssh_password_secret;type:text"`
	SSHPrivateKeySecret    *string     `gorm:"column:ssh_private_key_secret;type:text"`
	SSHKeyPassphraseSecret *string     `gorm:"column:ssh_key_passphrase_secret;type:text"`
	Notes                  *string     `gorm:"type:text"`
	Address                *string     `gorm:"size:255"`
	Status                 *string     `gorm:"size:64;index:idx_ceph_host_status"`
	ConfiguredData         *string     `json:"-" gorm:"column:configured_data;type:text"`
	DiscoveredData         string      `json:"-" gorm:"column:discovered_data;type:text;not null;default:'{}'"`
	Generation             uint64      `gorm:"not null;default:0"`
	ResourceVersion        uint64      `gorm:"not null;default:1"`
	Source                 string      `gorm:"size:32;not null;default:''"`
	SourceVersion          *string     `gorm:"size:128"`
	ObservedAt             *time.Time  `gorm:"index:idx_ceph_host_observed"`
	StaleAt                *time.Time  `gorm:"index:idx_ceph_host_stale"`
	CreatedAt              time.Time   `gorm:"not null"`
	UpdatedAt              time.Time   `gorm:"not null"`
}

func (CephHost) TableName() string { return "ceph_host" }

type CephEntityRecord struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`
	ClusterID       uint64    `gorm:"not null"`
	Kind            string    `gorm:"-"`
	NaturalKey      string    `gorm:"size:512;not null"`
	ParentKind      *string   `gorm:"size:64"`
	ParentKey       *string   `gorm:"size:512"`
	Name            *string   `gorm:"size:512"`
	Status          *string   `gorm:"size:64"`
	Generation      uint64    `gorm:"not null"`
	ResourceVersion uint64    `gorm:"not null;default:1"`
	Source          string    `gorm:"size:32;not null"`
	SourceVersion   *string   `gorm:"size:128"`
	ObservedAt      time.Time `gorm:"not null"`
	StaleAt         *time.Time
	ConfiguredData  *string   `json:"-" gorm:"column:configured_data;type:text"`
	DiscoveredData  string    `json:"-" gorm:"column:discovered_data;type:text;not null"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

type CephCollectionRun struct {
	ID           uint64      `gorm:"primaryKey;autoIncrement"`
	ClusterID    uint64      `gorm:"not null;uniqueIndex:uq_collection_run,priority:1;index:idx_collection_module_started,priority:1"`
	Cluster      CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Module       string      `gorm:"size:64;not null;uniqueIndex:uq_collection_run,priority:2;index:idx_collection_module_started,priority:2"`
	Generation   uint64      `gorm:"not null;uniqueIndex:uq_collection_run,priority:3"`
	Status       string      `gorm:"size:32;not null;index:idx_collection_status_started,priority:1"`
	Source       string      `gorm:"size:32;not null"`
	RecordCount  uint64      `gorm:"not null;default:0"`
	ErrorCode    *string     `gorm:"size:64"`
	ErrorMessage *string     `gorm:"type:text"`
	StartedAt    time.Time   `gorm:"not null;index:idx_collection_module_started,priority:3;index:idx_collection_status_started,priority:2"`
	FinishedAt   *time.Time  `gorm:"index:idx_collection_finished"`
	CreatedAt    time.Time   `gorm:"not null"`
}

func (CephCollectionRun) TableName() string { return "ceph_collection_run" }

type AuditEvent struct {
	ID               uint64       `gorm:"primaryKey;autoIncrement"`
	OccurredAt       time.Time    `gorm:"not null;index:idx_audit_occurred;index:idx_audit_actor_occurred,priority:2;index:idx_audit_cluster_occurred,priority:2;index:idx_audit_action_occurred,priority:2;index:idx_audit_resource_occurred,priority:3"`
	EventType        string       `gorm:"size:32;not null"`
	RequestID        string       `gorm:"size:64;not null;index:idx_audit_request"`
	ActorUserID      *uint64      `gorm:"index:idx_audit_actor_occurred,priority:1"`
	ActorUser        *User        `gorm:"constraint:OnDelete:SET NULL"`
	ActorUsername    string       `gorm:"size:128;not null"`
	ClusterID        *uint64      `gorm:"index:idx_audit_cluster_occurred,priority:1"`
	Cluster          *CephCluster `gorm:"constraint:OnDelete:SET NULL"`
	ClusterName      *string      `gorm:"size:128"`
	Action           string       `gorm:"size:128;not null;index:idx_audit_action_occurred,priority:1"`
	ResourceKind     *string      `gorm:"size:64;index:idx_audit_resource_occurred,priority:1"`
	ResourceKey      *string      `gorm:"size:512;index:idx_audit_resource_occurred,priority:2"`
	Risk             *string      `gorm:"size:16"`
	Outcome          string       `gorm:"size:32;not null"`
	HTTPStatus       *int
	ErrorCode        *string `gorm:"size:64"`
	SourceIP         *string `gorm:"size:64"`
	UserAgent        *string `gorm:"size:512"`
	BeforeGeneration *uint64
	AfterGeneration  *uint64
	ParametersJSON   *string `gorm:"type:text"`
	DetailsJSON      *string `gorm:"type:text"`
	PreviousHash     *string `gorm:"size:64"`
	EventHash        string  `gorm:"size:64;not null;uniqueIndex:uq_audit_hash"`
}

func (AuditEvent) TableName() string { return "audit_event" }
