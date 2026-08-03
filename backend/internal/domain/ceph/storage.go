package ceph

type OSD struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	Up          *bool    `json:"up"`
	In          *bool    `json:"in"`
	Weight      *float64 `json:"weight"`
	DeviceClass *string  `json:"device_class"`
	Host        *string  `json:"host"`
}
type Pool struct {
	Name                string         `json:"name"`
	ID                  int64          `json:"id"`
	Type                string         `json:"type"`
	Size                *int64         `json:"size"`
	MinSize             *int64         `json:"min_size"`
	PGNum               *int64         `json:"pg_num"`
	PGPNum              *int64         `json:"pgp_num"`
	PGAutoscaleMode     *string        `json:"pg_autoscale_mode"`
	Applications        []string       `json:"applications,omitempty"`
	ApplicationMetadata map[string]any `json:"application_metadata,omitempty"`
	CrushRule           *string        `json:"crush_rule"`
	CompressionMode     *string        `json:"compression_mode"`
	QuotaMaxBytes       *int64         `json:"quota_max_bytes"`
	QuotaMaxObjects     *int64         `json:"quota_max_objects"`
}
type Filesystem struct {
	Name   string           `json:"name"`
	ID     int64            `json:"id"`
	MaxMDS *int64           `json:"max_mds"`
	In     []int64          `json:"in"`
	Up     map[string]int64 `json:"up"`
}
type RBDImage struct {
	ImageSpec string  `json:"image_spec"`
	Pool      string  `json:"pool"`
	Name      string  `json:"name"`
	SizeBytes *uint64 `json:"size_bytes"`
	Format    *int    `json:"format"`
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
	Realms []string `json:"realms"`
}
