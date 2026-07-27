package ceph

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type Observation struct {
	Kind, NaturalKey, Name, Status, Source, SourceVersion string
	ParentKind, ParentKey                                 string
	Payload                                               any
	ObservedAt                                            time.Time
}
type CollectorProvider interface {
	Collect(context.Context, ClusterAccess, string) ([]Observation, error)
}

// CollectionResult distinguishes a successful empty collection from resource
// kinds whose optional command was unavailable during this collection.
type CollectionResult struct {
	Observations     []Observation
	UnavailableKinds []string
}

type CollectionMetadataProvider interface {
	CollectWithMetadata(context.Context, ClusterAccess, string) (CollectionResult, error)
}

type collectionTraceKey struct{}
type collectionTrace struct{ unavailable map[string]struct{} }

var collectionFailureKinds = map[string][]string{
	"collect.mds_fs":                  {"mds"},
	"collect.upgrade":                 {"upgrade"},
	"collect.cephfs_subvolume":        {"subvolume"},
	"collect.rbd_image":               {"rbd_image"},
	"collect.rgw_status":              {"rgw_status"},
	"collect.nfs_cluster":             {"nfs_cluster", "nfs_export"},
	"collect.smb_cluster":             {"smb_cluster", "smb_share"},
	"collect.rbd_namespace":           {"rbd_namespace"},
	"collect.rbd_image_detail":        {"rbd_snapshot"},
	"collect.rbd_snapshot":            {"rbd_snapshot"},
	"collect.rbd_trash":               {"rbd_trash"},
	"collect.rbd_group":               {"rbd_group"},
	"collect.rbd_mirroring":           {"rbd_mirroring"},
	"collect.cephfs_group":            {"subvolume_group"},
	"collect.cephfs_subvolume_detail": {"cephfs_snapshot"},
	"collect.cephfs_snapshot":         {"cephfs_snapshot"},
	"collect.rgw_user":                {"rgw_user"},
	"collect.rgw_account":             {"rgw_account"},
	"collect.rgw_role":                {"rgw_role"},
	"collect.rgw_bucket":              {"rgw_bucket"},
	"collect.rgw_realm":               {"rgw_realm"},
	"collect.rgw_zonegroup":           {"rgw_zonegroup"},
	"collect.rgw_zone":                {"rgw_zone"},
	"collect.nfs_export_cluster":      {"nfs_export"},
	"collect.nfs_export":              {"nfs_export"},
	"collect.smb_share_cluster":       {"smb_share"},
	"collect.smb_share":               {"smb_share"},
	"collect.osd_removal":             {"osd_removal"},
	"collect.config_option":           {"config_option"},
	"collect.mgr_module":              {"mgr_module"},
	"collect.crush_rule":              {"crush_rule"},
	"collect.erasure_code_profile":    {"erasure_code_profile"},
}

func (p *NativeProvider) CollectWithMetadata(ctx context.Context, access ClusterAccess, module string) (CollectionResult, error) {
	trace := &collectionTrace{unavailable: map[string]struct{}{}}
	rows, err := p.Collect(context.WithValue(ctx, collectionTraceKey{}, trace), access, module)
	kinds := make([]string, 0, len(trace.unavailable))
	for kind := range trace.unavailable {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return CollectionResult{Observations: rows, UnavailableKinds: kinds}, err
}

func (p *NativeProvider) Collect(ctx context.Context, access ClusterAccess, module string) ([]Observation, error) {
	switch module {
	case "fast":
		return p.collectFast(ctx, access)
	case "topology":
		return p.collectTopology(ctx, access)
	case "storage":
		return p.collectStorage(ctx, access)
	case "inventory":
		return p.collectInventory(ctx, access)
	case "configuration":
		return p.collectConfiguration(ctx, access)
	default:
		return nil, fmt.Errorf("unknown collection module %q", module)
	}
}

type statusWire struct {
	FSID   string `json:"fsid"`
	Health struct {
		Status string                     `json:"status"`
		Checks map[string]healthCheckWire `json:"checks"`
	} `json:"health"`
	MonMap struct {
		NumMons int   `json:"num_mons"`
		Quorum  []int `json:"quorum"`
	} `json:"monmap"`
	OSDMap struct {
		NumOSDs   int `json:"num_osds"`
		NumUpOSDs int `json:"num_up_osds"`
		NumInOSDs int `json:"num_in_osds"`
	} `json:"osdmap"`
	PGMap struct {
		NumPGs     uint64 `json:"num_pgs"`
		PGsByState []struct {
			StateName string `json:"state_name"`
			Count     uint64 `json:"count"`
		} `json:"pgs_by_state"`
		ReadBytesSec  *uint64 `json:"read_bytes_sec"`
		WriteBytesSec *uint64 `json:"write_bytes_sec"`
		ReadOpPerSec  *uint64 `json:"read_op_per_sec"`
		WriteOpPerSec *uint64 `json:"write_op_per_sec"`
	} `json:"pgmap"`
	MgrMap struct {
		Available   bool `json:"available"`
		NumStandbys int  `json:"num_standbys"`
	} `json:"mgrmap"`
	FSMap struct {
		Up       int               `json:"up"`
		ByRank   []json.RawMessage `json:"by_rank"`
		Standbys []json.RawMessage `json:"standbys"`
	} `json:"fsmap"`
}
type healthCheckWire struct {
	Severity string `json:"severity"`
	Summary  struct {
		Message string `json:"message"`
		Count   *int64 `json:"count"`
	} `json:"summary"`
	Detail []struct {
		Message string `json:"message"`
	} `json:"detail"`
	Muted bool `json:"muted"`
}
type dfWire struct {
	Stats struct {
		TotalBytes      *uint64 `json:"total_bytes"`
		TotalUsedBytes  *uint64 `json:"total_used_bytes"`
		TotalAvailBytes *uint64 `json:"total_avail_bytes"`
	} `json:"stats"`
}

func (p *NativeProvider) collectFast(ctx context.Context, access ClusterAccess) ([]Observation, error) {
	now := time.Now().UTC()
	var status statusWire
	if err := p.runInto(ctx, access, "collect.status", []string{"status", "--format", "json"}, &status); err != nil {
		return nil, err
	}
	var df dfWire
	if err := p.runInto(ctx, access, "collect.df", []string{"df", "detail", "--format", "json"}, &df); err != nil {
		return nil, err
	}
	if strings.TrimSpace(status.FSID) == "" || strings.TrimSpace(status.Health.Status) == "" {
		return nil, fmt.Errorf("parse collect.status response: fsid and health.status are required")
	}
	if df.Stats.TotalBytes == nil || df.Stats.TotalUsedBytes == nil || df.Stats.TotalAvailBytes == nil {
		return nil, fmt.Errorf("parse collect.df response: total byte fields are required")
	}
	overview := cephdomain.Overview{FSID: status.FSID, HealthStatus: status.Health.Status, Capacity: cephdomain.Capacity{TotalBytes: df.Stats.TotalBytes, UsedBytes: df.Stats.TotalUsedBytes, AvailableBytes: df.Stats.TotalAvailBytes}, Services: map[string]cephdomain.ServiceCount{"mon": {Total: &status.MonMap.NumMons, InQuorum: intPointer(len(status.MonMap.Quorum))}, "mgr": {Active: intPointer(boolInt(status.MgrMap.Available)), Standby: &status.MgrMap.NumStandbys}, "osd": {Total: &status.OSDMap.NumOSDs, Up: &status.OSDMap.NumUpOSDs, In: &status.OSDMap.NumInOSDs}, "mds": {Active: &status.FSMap.Up, Standby: intPointer(len(status.FSMap.Standbys))}}, ClientIO: cephdomain.ClientIO{ReadBytesPerSecond: status.PGMap.ReadBytesSec, WriteBytesPerSecond: status.PGMap.WriteBytesSec, ReadOpsPerSecond: status.PGMap.ReadOpPerSec, WriteOpsPerSecond: status.PGMap.WriteOpPerSec}, ObservedAt: now}
	for _, state := range status.PGMap.PGsByState {
		overview.PlacementGroups = append(overview.PlacementGroups, cephdomain.PGState{Name: state.StateName, Count: state.Count})
	}
	rows := []Observation{{Kind: "overview", NaturalKey: "overview", Name: "overview", Status: status.Health.Status, Source: "ceph_cli", Payload: overview, ObservedAt: now}}
	for code, check := range status.Health.Checks {
		details := make([]string, 0, len(check.Detail))
		for _, detail := range check.Detail {
			details = append(details, detail.Message)
		}
		payload := cephdomain.HealthCheck{Code: code, Severity: check.Severity, Summary: check.Summary.Message, Detail: details, Count: check.Summary.Count, Muted: check.Muted}
		rows = append(rows, Observation{Kind: "health_check", NaturalKey: code, Name: code, Status: check.Severity, Source: "ceph_cli", Payload: payload, ObservedAt: now})
	}
	return rows, nil
}

type hostWire struct {
	Hostname    string   `json:"hostname"`
	Addr        *string  `json:"addr"`
	Status      *string  `json:"status"`
	Labels      []string `json:"labels"`
	CephVersion *string  `json:"ceph_version"`
}
type daemonWire struct {
	DaemonName         string  `json:"daemon_name"`
	DaemonType         string  `json:"daemon_type"`
	Hostname           *string `json:"hostname"`
	StatusDesc         *string `json:"status_desc"`
	Version            *string `json:"version"`
	ContainerImageName *string `json:"container_image_name"`
}
type serviceWire struct {
	ServiceName string `json:"service_name"`
	ServiceType string `json:"service_type"`
	Running     *int   `json:"running"`
	Size        *int   `json:"size"`
	Placement   any    `json:"placement"`
}
type monDumpWire struct {
	Mons []struct {
		Name        string `json:"name"`
		Rank        int    `json:"rank"`
		PublicAddrs struct {
			AddrVec []struct {
				Addr string `json:"addr"`
			} `json:"addrvec"`
		} `json:"public_addrs"`
		PublicAddr string `json:"public_addr"`
		Addr       string `json:"addr"`
	} `json:"mons"`
}
type quorumWire struct {
	QuorumNames []string `json:"quorum_names"`
}
type mgrDumpWire struct {
	Available  bool   `json:"available"`
	ActiveName string `json:"active_name"`
	ActiveAddr string `json:"active_addr"`
	Standbys   []struct {
		Name string `json:"name"`
	} `json:"standbys"`
}

func (p *NativeProvider) collectTopology(ctx context.Context, access ClusterAccess) ([]Observation, error) {
	now := time.Now().UTC()
	var hosts []hostWire
	if err := p.runInto(ctx, access, "collect.host", []string{"orch", "host", "ls", "--detail", "--format", "json"}, &hosts); err != nil {
		return nil, err
	}
	var daemons []daemonWire
	if err := p.runInto(ctx, access, "collect.daemon", []string{"orch", "ps", "--refresh", "--format", "json"}, &daemons); err != nil {
		return nil, err
	}
	var services []serviceWire
	if err := p.runInto(ctx, access, "collect.service", []string{"orch", "ls", "--export", "--format", "json"}, &services); err != nil {
		return nil, err
	}
	var mons monDumpWire
	if err := p.runInto(ctx, access, "collect.mon", []string{"mon", "dump", "--format", "json"}, &mons); err != nil {
		return nil, err
	}
	var quorum quorumWire
	if err := p.runInto(ctx, access, "collect.quorum", []string{"quorum_status", "--format", "json"}, &quorum); err != nil {
		return nil, err
	}
	var managers mgrDumpWire
	if err := p.runInto(ctx, access, "collect.mgr", []string{"mgr", "dump", "--format", "json"}, &managers); err != nil {
		return nil, err
	}
	rows := make([]Observation, 0, len(hosts)+len(daemons)+len(services)+len(mons.Mons)+1+len(managers.Standbys))
	for _, wire := range hosts {
		if strings.TrimSpace(wire.Hostname) == "" {
			return nil, fmt.Errorf("parse collect.host response: hostname is required")
		}
		payload := cephdomain.Host{Hostname: wire.Hostname, Address: wire.Addr, Status: wire.Status, Labels: wire.Labels, CephVersion: wire.CephVersion}
		rows = append(rows, Observation{Kind: "host", NaturalKey: wire.Hostname, Name: wire.Hostname, Status: value(wire.Status), Source: "ceph_cli", Payload: payload, ObservedAt: now})
	}
	for _, wire := range daemons {
		if strings.TrimSpace(wire.DaemonName) == "" || strings.TrimSpace(wire.DaemonType) == "" {
			return nil, fmt.Errorf("parse collect.daemon response: daemon_name and daemon_type are required")
		}
		payload := cephdomain.Daemon{Name: wire.DaemonName, Type: wire.DaemonType, Hostname: wire.Hostname, Status: wire.StatusDesc, Version: wire.Version, ContainerImage: wire.ContainerImageName}
		rows = append(rows, Observation{Kind: "daemon", NaturalKey: wire.DaemonName, Name: wire.DaemonName, Status: value(wire.StatusDesc), Source: "ceph_cli", Payload: payload, ObservedAt: now})
	}
	for _, wire := range services {
		if strings.TrimSpace(wire.ServiceName) == "" || strings.TrimSpace(wire.ServiceType) == "" {
			return nil, fmt.Errorf("parse collect.service response: service_name and service_type are required")
		}
		payload := cephdomain.Service{Name: wire.ServiceName, Type: wire.ServiceType, Running: wire.Running, Size: wire.Size, Placement: wire.Placement}
		rows = append(rows, Observation{Kind: "service", NaturalKey: wire.ServiceName, Name: wire.ServiceName, Source: "ceph_cli", Payload: payload, ObservedAt: now})
	}
	quorumSet := make(map[string]struct{}, len(quorum.QuorumNames))
	for _, name := range quorum.QuorumNames {
		quorumSet[name] = struct{}{}
	}
	for _, wire := range mons.Mons {
		if strings.TrimSpace(wire.Name) == "" {
			return nil, fmt.Errorf("parse collect.mon response: monitor name is required")
		}
		address := wire.PublicAddr
		if address == "" {
			address = wire.Addr
		}
		if len(wire.PublicAddrs.AddrVec) > 0 {
			address = wire.PublicAddrs.AddrVec[0].Addr
		}
		_, inQuorum := quorumSet[wire.Name]
		payload := cephdomain.Monitor{Name: wire.Name, Rank: wire.Rank, Address: address, InQuorum: inQuorum}
		status := "out_of_quorum"
		if inQuorum {
			status = "in_quorum"
		}
		rows = append(rows, Observation{Kind: "mon", NaturalKey: wire.Name, Name: wire.Name, Status: status, Source: "ceph_cli", Payload: payload, ObservedAt: now})
	}
	if managers.ActiveName != "" {
		address := managers.ActiveAddr
		payload := cephdomain.Manager{Name: managers.ActiveName, Active: true, Address: &address, Available: managers.Available}
		rows = append(rows, Observation{Kind: "mgr", NaturalKey: managers.ActiveName, Name: managers.ActiveName, Status: "active", Source: "ceph_cli", Payload: payload, ObservedAt: now})
	}
	for _, wire := range managers.Standbys {
		payload := cephdomain.Manager{Name: wire.Name, Available: managers.Available}
		rows = append(rows, Observation{Kind: "mgr", NaturalKey: wire.Name, Name: wire.Name, Status: "standby", Source: "ceph_cli", Payload: payload, ObservedAt: now})
	}
	var mds fsDumpWire
	if err := p.runInto(ctx, access, "collect.mds_fs", []string{"fs", "dump", "--format", "json"}, &mds); err == nil {
		for _, filesystem := range mds.Filesystems {
			for _, info := range filesystem.MDSMap.Info {
				rank := info.Rank
				payload := cephdomain.MetadataServer{Name: info.Name, Filesystem: filesystem.MDSMap.FSName, Rank: &rank, State: info.State}
				rows = append(rows, Observation{Kind: "mds", NaturalKey: info.Name, Name: info.Name, Status: info.State, Source: "ceph_cli", Payload: payload, ObservedAt: now})
			}
		}
		for _, standby := range mds.Standbys {
			payload := cephdomain.MetadataServer{Name: standby.Name, State: "standby", Standby: true}
			rows = append(rows, Observation{Kind: "mds", NaturalKey: standby.Name, Name: standby.Name, Status: "standby", Source: "ceph_cli", Payload: payload, ObservedAt: now})
		}
	}
	var upgrade map[string]any
	if err := p.runInto(ctx, access, "collect.upgrade", []string{"orch", "upgrade", "status", "--format", "json"}, &upgrade); err == nil {
		status := textValue(upgrade["in_progress"])
		rows = append(rows, Observation{Kind: "upgrade", NaturalKey: "upgrade", Name: "upgrade", Status: status, Source: "ceph_cli", Payload: upgrade, ObservedAt: now})
	}
	return rows, nil
}

type osdTreeWire struct {
	Nodes []struct {
		ID          int      `json:"id"`
		Name        string   `json:"name"`
		Type        string   `json:"type"`
		Status      string   `json:"status"`
		CrushWeight *float64 `json:"crush_weight"`
		DeviceClass *string  `json:"device_class"`
		Children    []int    `json:"children"`
	} `json:"nodes"`
}
type osdDumpWire struct {
	Flags string `json:"flags"`
	OSDs  []struct {
		OSD int `json:"osd"`
		Up  int `json:"up"`
		In  int `json:"in"`
	} `json:"osds"`
}
type poolWire struct {
	Pool      int64   `json:"pool"`
	PoolName  string  `json:"pool_name"`
	Type      int     `json:"type"`
	Size      *int64  `json:"size"`
	MinSize   *int64  `json:"min_size"`
	PGNum     *int64  `json:"pg_num"`
	PGPNum    *int64  `json:"pg_placement_num"`
	CrushRule *string `json:"crush_rule"`
}
type fsDumpWire struct {
	Standbys []struct {
		Name string `json:"name"`
	} `json:"standbys"`
	Filesystems []struct {
		MDSMap struct {
			FSName string           `json:"fs_name"`
			ID     int64            `json:"id"`
			MaxMDS *int64           `json:"max_mds"`
			In     []int64          `json:"in"`
			Up     map[string]int64 `json:"up"`
			Info   map[string]struct {
				Name  string `json:"name"`
				Rank  int    `json:"rank"`
				State string `json:"state"`
			} `json:"info"`
		} `json:"mdsmap"`
	} `json:"filesystems"`
}
type rbdImageWire struct {
	Name   string  `json:"name"`
	Size   *uint64 `json:"size"`
	Format *int    `json:"format"`
}
type namedWire struct {
	Name string `json:"name"`
}
type rgwRealmWire struct {
	Realms []string `json:"realms"`
}

func (p *NativeProvider) collectStorage(ctx context.Context, access ClusterAccess) ([]Observation, error) {
	now := time.Now().UTC()
	var tree osdTreeWire
	if err := p.runInto(ctx, access, "collect.osd_tree", []string{"osd", "tree", "--format", "json"}, &tree); err != nil {
		return nil, err
	}
	var dump osdDumpWire
	if err := p.runInto(ctx, access, "collect.osd_dump", []string{"osd", "dump", "--format", "json"}, &dump); err != nil {
		return nil, err
	}
	states := map[int][2]bool{}
	for _, osd := range dump.OSDs {
		states[osd.OSD] = [2]bool{osd.Up == 1, osd.In == 1}
	}
	var pools []poolWire
	if err := p.runInto(ctx, access, "collect.pool", []string{"osd", "pool", "ls", "detail", "--format", "json"}, &pools); err != nil {
		return nil, err
	}
	var fs fsDumpWire
	if err := p.runInto(ctx, access, "collect.fs", []string{"fs", "dump", "--format", "json"}, &fs); err != nil {
		return nil, err
	}
	rows := []Observation{}
	if dump.Flags != "" {
		rows = append(rows, Observation{Kind: "osd_flag", NaturalKey: "flags", Name: "flags", Source: "ceph_cli", Payload: map[string]any{"flags": strings.Split(dump.Flags, ",")}, ObservedAt: now})
	}
	for _, node := range tree.Nodes {
		if node.Type != "osd" {
			continue
		}
		if node.ID < 0 || strings.TrimSpace(node.Name) == "" {
			return nil, fmt.Errorf("parse collect.osd_tree response: OSD id and name are required")
		}
		state := states[node.ID]
		up, in := state[0], state[1]
		payload := cephdomain.OSD{ID: node.ID, Name: node.Name, Status: node.Status, Up: &up, In: &in, Weight: node.CrushWeight, DeviceClass: node.DeviceClass}
		rows = append(rows, Observation{Kind: "osd", NaturalKey: strconv.Itoa(node.ID), Name: node.Name, Status: node.Status, Source: "ceph_cli", Payload: payload, ObservedAt: now})
	}
	for _, wire := range pools {
		if strings.TrimSpace(wire.PoolName) == "" {
			return nil, fmt.Errorf("parse collect.pool response: pool_name is required")
		}
		kind := "replicated"
		if wire.Type == 3 {
			kind = "erasure"
		}
		payload := cephdomain.Pool{Name: wire.PoolName, ID: wire.Pool, Type: kind, Size: wire.Size, MinSize: wire.MinSize, PGNum: wire.PGNum, PGPNum: wire.PGPNum, CrushRule: wire.CrushRule}
		rows = append(rows, Observation{Kind: "pool", NaturalKey: wire.PoolName, Name: wire.PoolName, Status: "available", Source: "ceph_cli", Payload: payload, ObservedAt: now})
	}
	for _, wire := range fs.Filesystems {
		m := wire.MDSMap
		if strings.TrimSpace(m.FSName) == "" {
			return nil, fmt.Errorf("parse collect.fs response: fs_name is required")
		}
		payload := cephdomain.Filesystem{Name: m.FSName, ID: m.ID, MaxMDS: m.MaxMDS, In: m.In, Up: m.Up}
		rows = append(rows, Observation{Kind: "filesystem", NaturalKey: m.FSName, Name: m.FSName, Status: "available", Source: "ceph_cli", Payload: payload, ObservedAt: now})
		var subvolumes []namedWire
		if err := p.runBinaryInto(ctx, access, executor.BinaryCeph, "collect.cephfs_subvolume", []string{"fs", "subvolume", "ls", m.FSName, "--format", "json"}, &subvolumes); err == nil {
			for _, subvolume := range subvolumes {
				payload := cephdomain.CephFSSubvolume{Filesystem: m.FSName, Name: subvolume.Name}
				rows = append(rows, Observation{Kind: "subvolume", NaturalKey: m.FSName + "/" + subvolume.Name, ParentKind: "filesystem", ParentKey: m.FSName, Name: subvolume.Name, Status: "available", Source: "ceph_cli", Payload: payload, ObservedAt: now})
			}
		}
	}
	for _, pool := range pools {
		var images []rbdImageWire
		if err := p.runBinaryInto(ctx, access, executor.BinaryRBD, "collect.rbd_image", []string{"ls", "--long", pool.PoolName, "--format", "json"}, &images); err != nil {
			continue
		}
		for _, image := range images {
			if strings.TrimSpace(image.Name) == "" {
				return nil, fmt.Errorf("parse collect.rbd_image response: image name is required")
			}
			spec := pool.PoolName + "/" + image.Name
			key := base64.RawURLEncoding.EncodeToString([]byte(spec))
			payload := cephdomain.RBDImage{ImageSpec: key, Pool: pool.PoolName, Name: image.Name, SizeBytes: image.Size, Format: image.Format}
			rows = append(rows, Observation{Kind: "rbd_image", NaturalKey: key, Name: image.Name, Status: "available", Source: "rbd_cli", Payload: payload, ObservedAt: now})
		}
	}
	var realms rgwRealmWire
	if err := p.runBinaryInto(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_status", []string{"realm", "list", "--format", "json"}, &realms); err == nil {
		rows = append(rows, Observation{Kind: "rgw_status", NaturalKey: "status", Name: "status", Status: "available", Source: "rgw_admin", Payload: cephdomain.RGWStatus{Realms: realms.Realms}, ObservedAt: now})
	}
	for _, gateway := range []struct {
		id, kind, prefix string
	}{{"collect.nfs_cluster", "nfs_cluster", "nfs"}, {"collect.smb_cluster", "smb_cluster", "smb"}} {
		var names []string
		if err := p.runBinaryInto(ctx, access, executor.BinaryCeph, gateway.id, []string{gateway.prefix, "cluster", "ls", "--format", "json"}, &names); err != nil {
			continue
		}
		for _, name := range names {
			rows = append(rows, Observation{Kind: gateway.kind, NaturalKey: name, Name: name, Status: "available", Source: "ceph_cli", Payload: cephdomain.GatewayCluster{Name: name}, ObservedAt: now})
		}
	}
	rows = append(rows, p.collectStorageOptional(ctx, access, pools, fs, now)...)
	return rows, nil
}

type deviceWire struct {
	DeviceID        string            `json:"device_id"`
	Hostname        string            `json:"hostname"`
	Path            string            `json:"path"`
	Available       bool              `json:"available"`
	RejectedReasons []string          `json:"rejected_reasons"`
	DeviceType      *string           `json:"device_type"`
	Model           *string           `json:"model"`
	Vendor          *string           `json:"vendor"`
	Serial          *string           `json:"serial"`
	SizeBytes       *uint64           `json:"size_bytes"`
	Rotational      *bool             `json:"rotational"`
	Metadata        map[string]string `json:"metadata"`
}

func (p *NativeProvider) collectInventory(ctx context.Context, access ClusterAccess) ([]Observation, error) {
	now := time.Now().UTC()
	var devices []deviceWire
	if err := p.runInto(ctx, access, "collect.device", []string{"orch", "device", "ls", "--wide", "--refresh", "--format", "json"}, &devices); err != nil {
		return nil, err
	}
	rows := make([]Observation, 0, len(devices))
	for _, wire := range devices {
		if strings.TrimSpace(wire.Hostname) == "" || strings.TrimSpace(wire.Path) == "" {
			return nil, fmt.Errorf("parse collect.device response: hostname and path are required")
		}
		key := wire.Hostname + ":" + wire.Path
		payload := cephdomain.Device{ID: wire.DeviceID, Hostname: wire.Hostname, Path: wire.Path, Available: wire.Available, RejectedReasons: wire.RejectedReasons, DeviceType: wire.DeviceType, Model: wire.Model, Vendor: wire.Vendor, Serial: wire.Serial, SizeBytes: wire.SizeBytes, Rotational: wire.Rotational, Metadata: wire.Metadata}
		rows = append(rows, Observation{Kind: "device", NaturalKey: key, Name: key, Source: "ceph_cli", Payload: payload, ObservedAt: now})
	}
	return rows, nil
}

type configValueWire struct {
	Who     string  `json:"who"`
	Name    string  `json:"name"`
	Value   string  `json:"value"`
	Level   *string `json:"level"`
	Section *string `json:"section"`
}

func (p *NativeProvider) collectConfiguration(ctx context.Context, access ClusterAccess) ([]Observation, error) {
	now := time.Now().UTC()
	var values []configValueWire
	if err := p.runInto(ctx, access, "collect.config", []string{"config", "dump", "--format", "json"}, &values); err != nil {
		return nil, err
	}
	rows := make([]Observation, 0, len(values))
	for _, wire := range values {
		if strings.TrimSpace(wire.Who) == "" || strings.TrimSpace(wire.Name) == "" {
			return nil, fmt.Errorf("parse collect.config response: who and name are required")
		}
		key := wire.Who + ":" + wire.Name
		payload := cephdomain.ConfigValue{Who: wire.Who, Name: wire.Name, Value: wire.Value, Level: wire.Level, Section: wire.Section}
		rows = append(rows, Observation{Kind: "config_value", NaturalKey: key, Name: wire.Name, Source: "ceph_cli", Payload: payload, ObservedAt: now})
	}
	rows = append(rows, p.collectConfigurationOptional(ctx, access, now)...)
	return rows, nil
}
func (p *NativeProvider) runInto(ctx context.Context, access ClusterAccess, id string, args []string, out any) error {
	return p.runBinaryInto(ctx, access, executor.BinaryCeph, id, args, out)
}
func (p *NativeProvider) runBinaryInto(ctx context.Context, access ClusterAccess, binary executor.Binary, id string, args []string, out any) error {
	result, err := p.Executor.Run(ctx, access, executor.CommandSpec{ID: id, Binary: binary, Args: args, Timeout: 45 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	if err != nil {
		markCollectionUnavailable(ctx, id)
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(result.Stdout))
	decoder.UseNumber()
	if err := decoder.Decode(out); err != nil {
		markCollectionUnavailable(ctx, id)
		return fmt.Errorf("parse %s response: %w", id, err)
	}
	return nil
}

func markCollectionUnavailable(ctx context.Context, commandID string) {
	trace, _ := ctx.Value(collectionTraceKey{}).(*collectionTrace)
	if trace == nil {
		return
	}
	for _, kind := range collectionFailureKinds[commandID] {
		trace.unavailable[kind] = struct{}{}
	}
}
func value(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
func intPointer(value int) *int { return &value }
func textValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case bool:
		return strconv.FormatBool(typed)
	default:
		return ""
	}
}
