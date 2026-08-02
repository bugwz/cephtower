package ceph

type ClusterSummary struct {
	HealthStatus      string                 `json:"health_status"`
	Version           string                 `json:"version,omitempty"`
	MgrID             string                 `json:"mgr_id,omitempty"`
	MgrHost           string                 `json:"mgr_host,omitempty"`
	HaveMonConnection string                 `json:"have_mon_connection,omitempty"`
	ExecutingTasks    []string               `json:"executing_tasks,omitempty"`
	FinishedTasks     []TaskSummary          `json:"finished_tasks,omitempty"`
	RBDMirroring      map[string]int         `json:"rbd_mirroring,omitempty"`
	Raw               map[string]interface{} `json:"-"`
}

type TaskSummary struct {
	Name      string         `json:"name"`
	Metadata  map[string]any `json:"metadata"`
	BeginTime string         `json:"begin_time"`
	EndTime   string         `json:"end_time"`
	Duration  int            `json:"duration"`
	Progress  int            `json:"progress"`
	Success   bool           `json:"success"`
	RetValue  string         `json:"ret_value"`
	Exception string         `json:"exception"`
}

type Host struct {
	Hostname         string            `json:"hostname"`
	Addr             string            `json:"addr"`
	Labels           []string          `json:"labels"`
	ServiceType      string            `json:"service_type"`
	Status           string            `json:"status"`
	Services         []HostService     `json:"services"`
	ServiceInstances []ServiceInstance `json:"service_instances"`
	System           string            `json:"system,omitempty"`
	Platform         string            `json:"platform,omitempty"`
	Distro           string            `json:"distro,omitempty"`
	KernelRelease    string            `json:"kernel_release,omitempty"`
	KernelBuild      string            `json:"kernel_build,omitempty"`
	Arch             string            `json:"arch,omitempty"`
	CPUModel         string            `json:"cpu_model,omitempty"`
	CPUCores         int               `json:"cpu_cores,omitempty"`
	MemoryBytes      uint64            `json:"memory_bytes,omitempty"`
}

type HostService struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type ServiceInstance struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type ListHostsOptions struct {
	Sources                 string
	Facts                   *bool
	Offset                  *int
	Limit                   *int
	Search                  string
	Sort                    string
	IncludeServiceInstances *bool
}

type ListOSDsOptions struct {
	Offset *int
	Limit  *int
	Search string
	Sort   string
}

type HostRequest struct {
	Hostname string   `json:"hostname,omitempty"`
	Addr     string   `json:"addr,omitempty"`
	Labels   []string `json:"labels,omitempty"`
	Status   string   `json:"status,omitempty"`
}

type UpdateHostRequest struct {
	Drain        bool     `json:"drain,omitempty"`
	Force        bool     `json:"force,omitempty"`
	Labels       []string `json:"labels,omitempty"`
	Maintenance  bool     `json:"maintenance,omitempty"`
	UpdateLabels bool     `json:"update_labels,omitempty"`
}

type DaemonActionRequest struct {
	Action         string `json:"action,omitempty"`
	ContainerImage string `json:"container_image,omitempty"`
	Force          bool   `json:"force,omitempty"`
}
