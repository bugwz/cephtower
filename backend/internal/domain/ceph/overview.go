package ceph

import "time"

type HealthCheck struct {
	Code     string   `json:"code"`
	Severity string   `json:"severity"`
	Summary  string   `json:"summary"`
	Detail   []string `json:"detail"`
	Count    *int64   `json:"count"`
	Muted    bool     `json:"muted"`
}
type Capacity struct {
	TotalBytes     *uint64 `json:"total_bytes"`
	UsedBytes      *uint64 `json:"used_bytes"`
	AvailableBytes *uint64 `json:"available_bytes"`
}
type ServiceCount struct {
	Total    *int `json:"total,omitempty"`
	Active   *int `json:"active,omitempty"`
	Standby  *int `json:"standby,omitempty"`
	Up       *int `json:"up,omitempty"`
	In       *int `json:"in,omitempty"`
	InQuorum *int `json:"in_quorum,omitempty"`
}
type PGState struct {
	Name  string `json:"name"`
	Count uint64 `json:"count"`
}
type ClientIO struct {
	ReadBytesPerSecond  *uint64 `json:"read_bytes_per_second"`
	WriteBytesPerSecond *uint64 `json:"write_bytes_per_second"`
	ReadOpsPerSecond    *uint64 `json:"read_ops_per_second"`
	WriteOpsPerSecond   *uint64 `json:"write_ops_per_second"`
}
type Overview struct {
	FSID            string                  `json:"fsid"`
	CephVersion     string                  `json:"ceph_version,omitempty"`
	HealthStatus    string                  `json:"health_status"`
	Capacity        Capacity                `json:"capacity"`
	Services        map[string]ServiceCount `json:"services"`
	PlacementGroups []PGState               `json:"placement_groups"`
	ClientIO        ClientIO                `json:"client_io"`
	ObservedAt      time.Time               `json:"observed_at"`
}
