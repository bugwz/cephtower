package ceph

type OSD struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Status      string            `json:"status"`
	Up          *bool             `json:"up"`
	In          *bool             `json:"in"`
	Weight      *float64          `json:"weight"`
	DeviceClass *string           `json:"device_class"`
	Host        *string           `json:"host"`
	CrushPath   map[string]string `json:"crush_path,omitempty"`
}
type Pool struct {
	ClientIORate             *PoolClientIORate    `json:"client_io_rate"`
	AutoscaleStatus          *PoolAutoscaleStatus `json:"autoscale_status"`
	Name                     string               `json:"name"`
	ID                       int64                `json:"id"`
	Type                     string               `json:"type"`
	Size                     *int64               `json:"size"`
	MinSize                  *int64               `json:"min_size"`
	PGNum                    *int64               `json:"pg_num"`
	PGNumTarget              *int64               `json:"pg_num_target"`
	PGPNumTarget             *int64               `json:"pgp_num_target"`
	PGStatus                 map[string]uint64    `json:"pg_status"`
	UsedPercent              *float64             `json:"used_percent"`
	Stored                   *uint64              `json:"stored"`
	Objects                  *uint64              `json:"objects"`
	ReadBytes                *uint64              `json:"read_bytes"`
	ReadOperations           *uint64              `json:"read_operations"`
	WriteOperations          *uint64              `json:"write_operations"`
	WriteBytes               *uint64              `json:"write_bytes"`
	CompressBytesUsed        *uint64              `json:"compress_bytes_used"`
	CompressUnderBytes       *uint64              `json:"compress_under_bytes"`
	BytesUsed                *uint64              `json:"bytes_used"`
	MaxAvail                 *uint64              `json:"max_avail"`
	ErasureCodeProfile       *string              `json:"erasure_code_profile"`
	PGPNum                   *int64               `json:"pgp_num"`
	PGAutoscaleMode          *string              `json:"pg_autoscale_mode"`
	Applications             []string             `json:"applications,omitempty"`
	ApplicationMetadata      map[string]any       `json:"application_metadata,omitempty"`
	CrushRule                *int64               `json:"crush_rule"`
	Flags                    []string             `json:"flags,omitempty"`
	CompressionMode          *string              `json:"compression_mode"`
	CompressionAlgorithm     *string              `json:"compression_algorithm"`
	CompressionMinBlobSize   *int64               `json:"compression_min_blob_size"`
	CompressionMaxBlobSize   *int64               `json:"compression_max_blob_size"`
	CompressionRequiredRatio *float64             `json:"compression_required_ratio"`
	QuotaMaxBytes            *int64               `json:"quota_max_bytes"`
	QuotaMaxObjects          *int64               `json:"quota_max_objects"`
	RBDMirroring             *string              `json:"rbd_mirroring,omitempty"`
	RawDetail                map[string]any       `json:"raw_detail,omitempty"`
	Configuration            []PoolConfig         `json:"configuration,omitempty"`
}

type PoolAutoscaleStatus struct {
	LogicalUsed          *float64 `json:"logical_used"`
	RawUsedRate          *float64 `json:"raw_used_rate"`
	ActualCapacityRatio  *float64 `json:"actual_capacity_ratio"`
	CapacityRatio        *float64 `json:"capacity_ratio"`
	PGNumFinal           *uint64  `json:"pg_num_final"`
	WouldAdjust          *bool    `json:"would_adjust"`
	TargetBytes          *uint64  `json:"target_bytes"`
	SubtreeCapacity      *uint64  `json:"subtree_capacity"`
	TargetRatio          *float64 `json:"target_ratio"`
	EffectiveTargetRatio *float64 `json:"effective_target_ratio"`
	Bias                 *float64 `json:"bias"`
	Bulk                 *bool    `json:"bulk"`
}

type PoolClientIORate struct {
	ReadBytesSec  *uint64 `json:"read_bytes_sec"`
	WriteBytesSec *uint64 `json:"write_bytes_sec"`
	ReadOpsSec    *uint64 `json:"read_op_per_sec"`
	WriteOpsSec   *uint64 `json:"write_op_per_sec"`
}
type PoolConfig struct {
	Name        string `json:"name"`
	Value       any    `json:"value"`
	Source      string `json:"source,omitempty"`
	Description string `json:"description,omitempty"`
}
type Filesystem struct {
	Name         string           `json:"name"`
	ID           int64            `json:"id"`
	Enabled      *bool            `json:"enabled"`
	Created      *string          `json:"created"`
	MaxMDS       *int64           `json:"max_mds"`
	MetadataPool *int64           `json:"metadata_pool,omitempty"`
	DataPools    []int64          `json:"data_pools,omitempty"`
	In           []int64          `json:"in"`
	Up           map[string]int64 `json:"up"`
}
type RBDImage struct {
	Configuration  []PoolConfig               `json:"configuration,omitempty"`
	RuntimeStatus  map[string]any             `json:"runtime_status,omitempty"`
	ScheduleInfo   *RBDMirrorSnapshotSchedule `json:"schedule_info,omitempty"`
	Features       []string                   `json:"features,omitempty"`
	Parent         map[string]any             `json:"parent,omitempty"`
	Details        map[string]any             `json:"details,omitempty"`
	MirrorMode     string                     `json:"mirror_mode,omitempty"`
	MirrorState    string                     `json:"mirror_state,omitempty"`
	MirrorGlobalID string                     `json:"mirror_global_id,omitempty"`
	Primary        *bool                      `json:"primary,omitempty"`
	CreatedAt      string                     `json:"created_at,omitempty"`
	DataPool       string                     `json:"data_pool,omitempty"`
	BlockPrefix    string                     `json:"block_name_prefix,omitempty"`
	ImagePath      string                     `json:"image_path"`
	Namespace      string                     `json:"namespace"`
	ImageSpec      string                     `json:"image_spec"`
	Pool           string                     `json:"pool"`
	Name           string                     `json:"name"`
	SizeBytes      *uint64                    `json:"size_bytes"`
	UsedBytes      *uint64                    `json:"used_bytes,omitempty"`
	TotalUsedBytes *uint64                    `json:"total_used_bytes,omitempty"`
	ObjectCount    *uint64                    `json:"object_count,omitempty"`
	ObjectSize     *uint64                    `json:"object_size_bytes,omitempty"`
	StripeUnit     *uint64                    `json:"stripe_unit_bytes,omitempty"`
	StripeCount    *uint64                    `json:"stripe_count,omitempty"`
	Order          *uint64                    `json:"order,omitempty"`
	Format         *int                       `json:"format"`
	SnapshotUsage  map[string]uint64          `json:"-"`
}

type RBDMirrorSnapshotSchedule struct {
	Name          string                          `json:"name"`
	InheritedFrom string                          `json:"inherited_from,omitempty"`
	NextRun       string                          `json:"schedule_time,omitempty"`
	Status        string                          `json:"schedule_status"`
	Intervals     []RBDMirrorSnapshotScheduleItem `json:"schedule_interval"`
}

type RBDMirrorSnapshotScheduleItem struct {
	Interval  string `json:"interval"`
	StartTime string `json:"start_time,omitempty"`
}
type CephFSSubvolume struct {
	Filesystem string `json:"filesystem"`
	Name       string `json:"name"`
	Group      string `json:"group,omitempty"`
}
type GatewayCluster struct {
	Name string `json:"name"`
}
type RGWStatus struct {
	GlobalRateLimit map[string]any `json:"global_rate_limit,omitempty"`
	Realms          []string       `json:"realms"`
}
