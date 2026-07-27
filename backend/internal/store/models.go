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
	ID               uint64    `gorm:"primaryKey;autoIncrement"`
	Name             string    `gorm:"size:128;not null;uniqueIndex:uq_cluster_name"`
	MonitorAddresses string    `gorm:"size:4096;not null"`
	ClientUsername   string    `gorm:"size:128;not null"`
	ClientKey        string    `gorm:"type:text;not null"`
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

type CephClusterObservation struct {
	ClusterID        uint64      `gorm:"primaryKey;autoIncrement:false"`
	Cluster          CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	FSID             *string     `gorm:"column:fsid;size:36;uniqueIndex:uq_observation_fsid"`
	CephVersion      *string     `gorm:"size:128"`
	Status           string      `gorm:"size:32;not null;default:unknown;index:idx_observation_status_enabled,priority:1"`
	Enabled          bool        `gorm:"not null;default:true;check:observation_enabled_check,enabled IN (0,1);index:idx_observation_status_enabled,priority:2"`
	Generation       uint64      `gorm:"not null;default:0"`
	LastSeenAt       *time.Time  `gorm:"index:idx_observation_last_seen"`
	LastErrorCode    *string     `gorm:"size:64"`
	LastErrorMessage *string     `gorm:"type:text"`
	ObservedAt       *time.Time
	UpdatedAt        time.Time `gorm:"not null"`
}

func (CephClusterObservation) TableName() string { return "ceph_cluster_observation" }

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

type CephResourceRecord struct {
	ID                   uint64      `gorm:"primaryKey;autoIncrement"`
	ClusterID            uint64      `gorm:"not null;uniqueIndex:uq_resource,priority:1;index:idx_resource_name,priority:1;index:idx_resource_status,priority:1;index:idx_resource_generation,priority:1;index:idx_resource_observed,priority:1;index:idx_resource_parent,priority:1;index:idx_resource_stale,priority:1"`
	Cluster              CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Kind                 string      `gorm:"size:64;not null;uniqueIndex:uq_resource,priority:2;index:idx_resource_name,priority:2;index:idx_resource_status,priority:2;index:idx_resource_generation,priority:2;index:idx_resource_observed,priority:2"`
	NaturalKey           string      `gorm:"size:512;not null;uniqueIndex:uq_resource,priority:3"`
	ParentKind           *string     `gorm:"size:64;index:idx_resource_parent,priority:2"`
	ParentKey            *string     `gorm:"size:512;index:idx_resource_parent,priority:3"`
	Name                 *string     `gorm:"size:512;index:idx_resource_name,priority:3"`
	Status               *string     `gorm:"size:64;index:idx_resource_status,priority:3"`
	Generation           uint64      `gorm:"not null;index:idx_resource_generation,priority:3"`
	ResourceVersion      uint64      `gorm:"not null;default:1"`
	Source               string      `gorm:"size:32;not null"`
	SourceVersion        *string     `gorm:"size:128"`
	ObservedAt           time.Time   `gorm:"not null;index:idx_resource_observed,priority:3"`
	StaleAt              *time.Time  `gorm:"index:idx_resource_stale,priority:2"`
	PayloadSchemaVersion int         `gorm:"not null"`
	PayloadJSON          string      `gorm:"type:text;not null"`
	CreatedAt            time.Time   `gorm:"not null"`
	UpdatedAt            time.Time   `gorm:"not null"`
}

func (CephResourceRecord) TableName() string { return "ceph_resource_record" }

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

type CephActionPlan struct {
	ID                 string      `gorm:"primaryKey;size:36"`
	ClusterID          uint64      `gorm:"not null;index:idx_plan_resource,priority:1"`
	Cluster            CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	ActorUserID        *uint64     `gorm:"index:idx_plan_actor_created,priority:1"`
	ActorUser          *User       `gorm:"constraint:OnDelete:SET NULL"`
	ActorUsername      string      `gorm:"size:128;not null"`
	RequestID          string      `gorm:"size:64;not null;index:idx_plan_request"`
	Action             string      `gorm:"size:128;not null"`
	ResourceKind       string      `gorm:"size:64;not null;index:idx_plan_resource,priority:2"`
	ResourceKey        string      `gorm:"size:512;not null;index:idx_plan_resource,priority:3"`
	ResourceGeneration uint64      `gorm:"not null"`
	Risk               string      `gorm:"size:16;not null"`
	Status             string      `gorm:"size:32;not null;index:idx_plan_status_expiry,priority:1"`
	RequestJSON        string      `gorm:"type:text;not null"`
	BlockersJSON       string      `gorm:"type:text;not null;default:'[]'"`
	WarningsJSON       string      `gorm:"type:text;not null;default:'[]'"`
	ExpiresAt          time.Time   `gorm:"not null;index:idx_plan_status_expiry,priority:2"`
	ConsumedAt         *time.Time
	CreatedAt          time.Time `gorm:"not null;index:idx_plan_actor_created,priority:2"`
}

func (CephActionPlan) TableName() string { return "ceph_action_plan" }

type CephOperation struct {
	ID                   string          `gorm:"primaryKey;size:36"`
	ClusterID            *uint64         `gorm:"index:idx_operation_cluster_status_created,priority:1"`
	Cluster              *CephCluster    `gorm:"constraint:OnDelete:SET NULL"`
	ClusterName          string          `gorm:"size:128;not null"`
	ActorUserID          *uint64         `gorm:"index:idx_operation_actor_created,priority:1"`
	ActorUser            *User           `gorm:"constraint:OnDelete:SET NULL"`
	ActorUsername        string          `gorm:"size:128;not null"`
	PlanID               *string         `gorm:"size:36;index:idx_operation_plan"`
	Plan                 *CephActionPlan `gorm:"constraint:OnDelete:SET NULL"`
	RetryOfID            *string         `gorm:"size:36;index:idx_operation_retry_of"`
	RetryOf              *CephOperation  `gorm:"foreignKey:RetryOfID;constraint:OnDelete:SET NULL"`
	RequestID            string          `gorm:"size:64;not null;index:idx_operation_request"`
	Action               string          `gorm:"size:128;not null"`
	ResourceKind         string          `gorm:"size:64;not null;index:idx_operation_resource_created,priority:1"`
	ResourceKey          string          `gorm:"size:512;not null;index:idx_operation_resource_created,priority:2"`
	ResourceGeneration   *uint64
	Risk                 string  `gorm:"size:16;not null"`
	Status               string  `gorm:"size:32;not null;default:queued;index:idx_operation_cluster_status_created,priority:2;index:idx_operation_status_scheduled,priority:1"`
	Stage                string  `gorm:"size:64;not null;default:queued"`
	Progress             int     `gorm:"not null;default:0;check:operation_progress_check,progress >= 0 AND progress <= 100"`
	Attempt              int     `gorm:"not null;default:0"`
	MaxAttempts          int     `gorm:"not null;default:1"`
	IdempotencyKeyHash   *string `gorm:"size:64"`
	IdempotencyScopeHash *string `gorm:"size:64;uniqueIndex:uq_operation_idempotency"`
	RequestJSON          string  `gorm:"type:text;not null"`
	ResultJSON           *string `gorm:"type:text"`
	ErrorCode            *string `gorm:"size:64"`
	ErrorMessage         *string `gorm:"type:text"`
	ErrorDetailsJSON     *string `gorm:"type:text"`
	Retryable            bool    `gorm:"not null;default:false;check:operation_retryable_check,retryable IN (0,1)"`
	CancelRequestedAt    *time.Time
	ScheduledAt          time.Time `gorm:"not null;index:idx_operation_status_scheduled,priority:2"`
	StartedAt            *time.Time
	HeartbeatAt          *time.Time `gorm:"index:idx_operation_heartbeat"`
	CompletedAt          *time.Time
	CreatedAt            time.Time `gorm:"not null;index:idx_operation_cluster_status_created,priority:3;index:idx_operation_actor_created,priority:2;index:idx_operation_resource_created,priority:3"`
	UpdatedAt            time.Time `gorm:"not null"`
}

func (CephOperation) TableName() string { return "ceph_operation" }

type CephOperationEvent struct {
	ID          uint64        `gorm:"primaryKey;autoIncrement"`
	OperationID string        `gorm:"size:36;not null;uniqueIndex:uq_operation_event,priority:1;index:idx_event_operation_created,priority:1"`
	Operation   CephOperation `gorm:"constraint:OnDelete:CASCADE"`
	Sequence    uint64        `gorm:"not null;uniqueIndex:uq_operation_event,priority:2"`
	EventType   string        `gorm:"size:32;not null"`
	Stage       string        `gorm:"size:64;not null"`
	Progress    *int
	Message     string    `gorm:"type:text;not null"`
	DataJSON    *string   `gorm:"type:text"`
	ErrorCode   *string   `gorm:"size:64"`
	CreatedAt   time.Time `gorm:"not null;index:idx_event_operation_created,priority:2;index:idx_event_created"`
}

func (CephOperationEvent) TableName() string { return "ceph_operation_event" }

type CephOperationLock struct {
	LockKey        string        `gorm:"primaryKey;size:191"`
	ClusterID      uint64        `gorm:"not null;index:idx_lock_resource,priority:1"`
	Cluster        CephCluster   `gorm:"constraint:OnDelete:CASCADE"`
	ResourceKind   string        `gorm:"size:64;not null;index:idx_lock_resource,priority:2"`
	ResourceKey    string        `gorm:"size:512;not null;index:idx_lock_resource,priority:3"`
	OperationID    string        `gorm:"size:36;not null;index:idx_lock_operation"`
	Operation      CephOperation `gorm:"constraint:OnDelete:CASCADE"`
	FencingToken   uint64        `gorm:"not null"`
	LeaseExpiresAt time.Time     `gorm:"not null;index:idx_lock_expiry"`
	AcquiredAt     time.Time     `gorm:"not null"`
	UpdatedAt      time.Time     `gorm:"not null"`
}

func (CephOperationLock) TableName() string { return "ceph_operation_lock" }

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
	PlanID           *string `gorm:"size:36"`
	OperationID      *string `gorm:"size:36;index:idx_audit_operation"`
	BeforeGeneration *uint64
	AfterGeneration  *uint64
	ParametersJSON   *string `gorm:"type:text"`
	DetailsJSON      *string `gorm:"type:text"`
	PreviousHash     *string `gorm:"size:64"`
	EventHash        string  `gorm:"size:64;not null;uniqueIndex:uq_audit_hash"`
}

func (AuditEvent) TableName() string { return "audit_event" }
