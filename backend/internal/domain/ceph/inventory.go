package ceph

type Host struct {
	Hostname    string   `json:"hostname"`
	Address     *string  `json:"address"`
	Status      *string  `json:"status"`
	Labels      []string `json:"labels"`
	CephVersion *string  `json:"ceph_version"`
}
type Daemon struct {
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Hostname       *string `json:"hostname"`
	Status         *string `json:"status"`
	Version        *string `json:"version"`
	ContainerImage *string `json:"container_image"`
}
type Service struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Running   *int   `json:"running"`
	Size      *int   `json:"size"`
	Placement any    `json:"placement,omitempty"`
}
type Monitor struct {
	Name     string `json:"name"`
	Rank     int    `json:"rank"`
	Address  string `json:"address"`
	InQuorum bool   `json:"in_quorum"`
}
type Manager struct {
	Name      string  `json:"name"`
	Active    bool    `json:"active"`
	Address   *string `json:"address"`
	Available bool    `json:"available"`
}
type MetadataServer struct {
	Name       string `json:"name"`
	Filesystem string `json:"filesystem,omitempty"`
	Rank       *int   `json:"rank"`
	State      string `json:"state"`
	Standby    bool   `json:"standby"`
}
type Device struct {
	ID              string            `json:"device_id"`
	Hostname        string            `json:"hostname"`
	Path            string            `json:"path"`
	Available       bool              `json:"available"`
	RejectedReasons []string          `json:"rejected_reasons"`
	DeviceType      *string           `json:"device_type,omitempty"`
	Model           *string           `json:"model,omitempty"`
	Vendor          *string           `json:"vendor,omitempty"`
	Serial          *string           `json:"serial,omitempty"`
	SizeBytes       *uint64           `json:"size_bytes,omitempty"`
	Rotational      *bool             `json:"rotational,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}
type ConfigValue struct {
	Who     string  `json:"who"`
	Name    string  `json:"name"`
	Value   string  `json:"value"`
	Level   *string `json:"level,omitempty"`
	Section *string `json:"section,omitempty"`
}
