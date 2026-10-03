package ceph

type Host struct {
	Hostname         string            `json:"hostname"`
	Address          *string           `json:"address"`
	Status           *string           `json:"status"`
	Labels           []string          `json:"labels"`
	ServiceType      *string           `json:"service_type,omitempty"`
	Services         []HostService     `json:"services,omitempty"`
	ServiceInstances []ServiceInstance `json:"service_instances,omitempty"`
	System           *string           `json:"system,omitempty"`
	Platform         *string           `json:"platform,omitempty"`
	Distro           *string           `json:"distro,omitempty"`
	KernelRelease    *string           `json:"kernel_release,omitempty"`
	KernelBuild      *string           `json:"kernel_build,omitempty"`
	Arch             *string           `json:"arch,omitempty"`
	CPUModel         *string           `json:"cpu_model,omitempty"`
	CPUCores         *int              `json:"cpu_cores,omitempty"`
	MemoryBytes      *uint64           `json:"memory_bytes,omitempty"`
}

type HostService struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type ServiceInstance struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type Daemon struct {
	DaemonRuntime
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Hostname       *string `json:"hostname"`
	Status         *string `json:"status"`
	Version        *string `json:"version"`
	ContainerImage *string `json:"container_image"`
	CPUPercentage  *string `json:"cpu_percentage,omitempty"`
	MemoryUsage    *uint64 `json:"memory_usage,omitempty"`
	LastRefresh    *string `json:"last_refresh,omitempty"`
}
type DaemonRuntime struct {
	DaemonID              *string  `json:"daemon_id,omitempty"`
	ContainerID           *string  `json:"container_id,omitempty"`
	ContainerImageID      *string  `json:"container_image_id,omitempty"`
	ContainerImageDigests []string `json:"container_image_digests,omitempty"`
	IP                    *string  `json:"ip,omitempty"`
	Ports                 []int    `json:"ports,omitempty"`
	SystemdUnit           *string  `json:"systemd_unit,omitempty"`
	IsActive              *bool    `json:"is_active,omitempty"`
	OSDSpecAffinity       *string  `json:"osdspec_affinity,omitempty"`
	Created               *string  `json:"created,omitempty"`
	Started               *string  `json:"started,omitempty"`
	LastDeployed          *string  `json:"last_deployed,omitempty"`
	LastConfigured        *string  `json:"last_configured,omitempty"`
	Events                []string `json:"events,omitempty"`
}
type Service struct {
	Networks           []string `json:"networks"`
	ContainerImageName *string  `json:"container_image_name"`
	ContainerImageID   *string  `json:"container_image_id"`
	ServiceURL         *string  `json:"service_url"`
	VirtualIP          *string  `json:"virtual_ip"`
	CephCreatedAt      *string  `json:"ceph_created_at"`
	Name               string   `json:"name"`
	Type               string   `json:"type"`
	Running            *int     `json:"running"`
	Size               *int     `json:"size"`
	Placement          any      `json:"placement,omitempty"`
	Unmanaged          bool     `json:"unmanaged"`
	LastRefresh        *string  `json:"last_refresh"`
	Ports              []int    `json:"ports"`
	Events             []string `json:"events"`
}
type Monitor struct {
	Name         string `json:"name"`
	Rank         int    `json:"rank"`
	Address      string `json:"address"`
	InQuorum     bool   `json:"in_quorum"`
	OpenSessions any    `json:"open_sessions,omitempty"`
}
type MonitorStatus struct {
	FSID        string   `json:"fsid"`
	Modified    string   `json:"modified"`
	Epoch       int      `json:"epoch"`
	QuorumCon   string   `json:"quorum_con"`
	QuorumMon   []string `json:"quorum_mon"`
	RequiredCon string   `json:"required_con"`
	RequiredMon []string `json:"required_mon"`
}
type MonitorPerfCounter struct {
	Monitor     string `json:"monitor"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Value       any    `json:"value"`
	RawValue    any    `json:"raw_value,omitempty"`
	Unit        string `json:"unit,omitempty"`
	MetricType  string `json:"metric_type,omitempty"`
	ValueType   string `json:"value_type,omitempty"`
	Priority    int    `json:"priority"`
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
	Who           string  `json:"who"`
	Name          string  `json:"name"`
	Value         string  `json:"value"`
	Level         *string `json:"level,omitempty"`
	Section       *string `json:"section,omitempty"`
	LocationType  *string `json:"location_type,omitempty"`
	LocationValue *string `json:"location_value,omitempty"`
	DeviceClass   *string `json:"device_class,omitempty"`
	Mask          *string `json:"mask,omitempty"`
}
