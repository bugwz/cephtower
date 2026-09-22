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
	"collect.health_detail":           {"health_check"},
	"collect.mds_fs":                  {"mds"},
	"collect.upgrade":                 {"upgrade"},
	"collect.cephfs_subvolume":        {"subvolume"},
	"collect.rbd_image_status":        {"rbd_image"},
	"collect.rbd_image_config":        {"rbd_image"},
	"collect.rbd_image_info":          {"rbd_image"},
	"collect.rbd_image":               {"rbd_image"},
	"collect.rgw_global_ratelimit":    {"rgw_status"},
	"collect.rgw_status":              {"rgw_status"},
	"collect.nfs_cluster":             {"nfs_cluster", "nfs_export"},
	"collect.smb_cluster":             {"smb_cluster", "smb_share"},
	"collect.rbd_namespace":           {"rbd_namespace", "rbd_trash", "rbd_snapshot", "rbd_image", "rbd_group"},
	"collect.rbd_image_detail":        {"rbd_snapshot", "rbd_image"},
	"collect.rbd_snapshot":            {"rbd_snapshot"},
	"collect.rbd_trash":               {"rbd_trash"},
	"collect.rbd_group_images":        {"rbd_group"},
	"collect.rbd_group_snapshots":     {"rbd_group"},
	"collect.rbd_group_info":          {"rbd_group"},
	"collect.rbd_group":               {"rbd_group"},
	"collect.rbd_mirroring_status":    {"rbd_mirroring"},
	"collect.rbd_mirroring":           {"rbd_mirroring"},
	"collect.cephfs_group":            {"subvolume_group"},
	"collect.cephfs_subvolume_detail": {"cephfs_snapshot"},
	"collect.cephfs_snapshot":         {"cephfs_snapshot"},
	"collect.rgw_user_ratelimit":      {"rgw_user"},
	"collect.rgw_user_stats":          {"rgw_user"},
	"collect.rgw_user_detail":         {"rgw_user"},
	"collect.rgw_user":                {"rgw_user"},
	"collect.rgw_account_stats":       {"rgw_account"},
	"collect.rgw_account_detail":      {"rgw_account", "rgw_role"},
	"collect.rgw_account":             {"rgw_account", "rgw_role"},
	"collect.rgw_role_detail":         {"rgw_role"},
	"collect.rgw_role":                {"rgw_role"},
	"collect.rgw_bucket_ratelimit":    {"rgw_bucket"},
	"collect.rgw_bucket_detail":       {"rgw_bucket"},
	"collect.rgw_bucket":              {"rgw_bucket"},
	"collect.rgw_zone_detail":         {"rgw_zone"},
	"collect.rgw_zonegroup_detail":    {"rgw_zonegroup"},
	"collect.rgw_realm_detail":        {"rgw_realm"},
	"collect.rgw_realm":               {"rgw_realm"},
	"collect.rgw_zonegroup":           {"rgw_zonegroup"},
	"collect.rgw_zone":                {"rgw_zone"},
	"collect.nfs_export_cluster":      {"nfs_export"},
	"collect.nfs_export":              {"nfs_export"},
	"collect.smb_share_cluster":       {"smb_share"},
	"collect.smb_share":               {"smb_share"},
	"collect.osd_removal":             {"osd_removal"},
	"collect.config_option":           {"config_option"},
	"collect.mgr_module_metadata":     {"mgr_module"},
	"collect.mgr_module":              {"mgr_module"},
	"collect.crush_rule":              {"crush_rule"},
	"collect.erasure_code_profile":    {"erasure_code_profile"},
	"collect.mon_perf":                {"mon_perf_counter"},
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
	case "ceph_auth":
		return p.collectCephUsers(ctx, access)
	case "configuration":
		return p.collectConfiguration(ctx, access)
	default:
		return nil, fmt.Errorf("unknown collection module %q", module)
	}
}

type statusWire struct {
	FSID   string `json:"fsid"`
	Health struct {
		Status string `json:"status"`
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
		NumPGs     uint64  `json:"num_pgs"`
		NumPools   *uint64 `json:"num_pools"`
		NumObjects *uint64 `json:"num_objects"`
		PGsByState []struct {
			StateName string `json:"state_name"`
			Count     uint64 `json:"count"`
		} `json:"pgs_by_state"`
		ReadBytesSec          *uint64 `json:"read_bytes_sec"`
		WriteBytesSec         *uint64 `json:"write_bytes_sec"`
		ReadOpPerSec          *uint64 `json:"read_op_per_sec"`
		WriteOpPerSec         *uint64 `json:"write_op_per_sec"`
		RecoveringBytesPerSec *uint64 `json:"recovering_bytes_per_sec"`
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

type pgDumpSummaryWire struct {
	PGMap struct {
		PGStatsSum struct {
			StatSum struct {
				NumObjects          *uint64 `json:"num_objects"`
				NumObjectCopies     *uint64 `json:"num_object_copies"`
				NumObjectsDegraded  *uint64 `json:"num_objects_degraded"`
				NumObjectsMisplaced *uint64 `json:"num_objects_misplaced"`
				NumObjectsUnfound   *uint64 `json:"num_objects_unfound"`
			} `json:"stat_sum"`
		} `json:"pg_stats_sum"`
	} `json:"pg_map"`
}

type osdDFWire struct {
	Nodes []struct {
		ID   *int    `json:"id"`
		PGs  *uint64 `json:"pgs"`
		Type string  `json:"type"`
	} `json:"nodes"`
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
	overview := cephdomain.Overview{FSID: status.FSID, HealthStatus: status.Health.Status, Capacity: cephdomain.Capacity{TotalBytes: df.Stats.TotalBytes, UsedBytes: df.Stats.TotalUsedBytes, AvailableBytes: df.Stats.TotalAvailBytes}, Services: map[string]cephdomain.ServiceCount{"mon": {Total: &status.MonMap.NumMons, InQuorum: intPointer(len(status.MonMap.Quorum))}, "mgr": {Active: intPointer(boolInt(status.MgrMap.Available)), Standby: &status.MgrMap.NumStandbys}, "osd": {Total: &status.OSDMap.NumOSDs, Up: &status.OSDMap.NumUpOSDs, In: &status.OSDMap.NumInOSDs}, "mds": {Active: &status.FSMap.Up, Standby: intPointer(len(status.FSMap.Standbys))}}, PoolCount: status.PGMap.NumPools, ObjectStats: cephdomain.ObjectStats{Objects: status.PGMap.NumObjects}, ClientIO: cephdomain.ClientIO{ReadBytesPerSecond: status.PGMap.ReadBytesSec, WriteBytesPerSecond: status.PGMap.WriteBytesSec, ReadOpsPerSecond: status.PGMap.ReadOpPerSec, WriteOpsPerSecond: status.PGMap.WriteOpPerSec, RecoveringBytesPerSecond: status.PGMap.RecoveringBytesPerSec}, ObservedAt: now}
	if versions, err := p.run(ctx, access, "collect.versions", 30*time.Second, "versions", "--format", "json"); err == nil {
		overview.CephVersion = cephVersionFromVersions(versions)
	}
	for _, state := range status.PGMap.PGsByState {
		overview.PlacementGroups = append(overview.PlacementGroups, cephdomain.PGState{Name: state.StateName, Count: state.Count})
	}
	p.collectOverviewDetails(ctx, access, &overview)
	rows := []Observation{{Kind: "overview", NaturalKey: "overview", Name: "overview", Status: status.Health.Status, Source: "ceph_cli", SourceVersion: overview.CephVersion, Payload: overview, ObservedAt: now}}
	var health struct {
		Status string                     `json:"status"`
		Checks map[string]healthCheckWire `json:"checks"`
		Mutes  []struct {
			Code    string `json:"code"`
			TTL     string `json:"ttl"`
			Sticky  bool   `json:"sticky"`
			Summary string `json:"summary"`
			Count   *int64 `json:"count"`
		} `json:"mutes"`
	}
	// Status contains summaries only. Keep cached checks stale if the detailed
	// command fails instead of replacing them with an apparently empty list.
	if !p.optional(ctx, access, executor.BinaryCeph, "collect.health_detail", []string{"health", "detail", "--format", "json"}, &health) {
		return rows, nil
	}
	if health.Status == "" || health.Checks == nil {
		return nil, fmt.Errorf("parse collect.health_detail response: status and checks are required")
	}
	checks := make(map[string]cephdomain.HealthCheck, len(health.Checks))
	for code, check := range health.Checks {
		details := make([]string, 0, len(check.Detail))
		for _, detail := range check.Detail {
			details = append(details, detail.Message)
		}
		payload := cephdomain.HealthCheck{Code: code, Severity: check.Severity, Summary: check.Summary.Message, Detail: details, Count: check.Summary.Count, Muted: check.Muted, Active: true}
		checks[code] = payload
	}
	for _, mute := range health.Mutes {
		if mute.Code == "" {
			continue
		}
		check, exists := checks[mute.Code]
		if !exists {
			check = cephdomain.HealthCheck{Code: mute.Code, Severity: "HEALTH_OK", Summary: mute.Summary, Count: mute.Count, Detail: []string{}}
		}
		check.Muted, check.Sticky, check.MuteUntil = true, mute.Sticky, mute.TTL
		checks[mute.Code] = check
	}
	for code, check := range checks {
		rows = append(rows, Observation{Kind: "health_check", NaturalKey: code, Name: code, Status: check.Severity, Source: "ceph_cli", Payload: check, ObservedAt: now})
	}
	return rows, nil
}

func (p *NativeProvider) collectOverviewDetails(ctx context.Context, access ClusterAccess, overview *cephdomain.Overview) {
	var pgSummary pgDumpSummaryWire
	if p.optional(ctx, access, executor.BinaryCeph, "collect.overview_pg_summary", []string{"pg", "dump", "summary", "--format", "json"}, &pgSummary) {
		stats := pgSummary.PGMap.PGStatsSum.StatSum
		if stats.NumObjects != nil && stats.NumObjectCopies != nil && stats.NumObjectsDegraded != nil && stats.NumObjectsMisplaced != nil && stats.NumObjectsUnfound != nil {
			overview.ObjectStats = cephdomain.ObjectStats{
				Objects: stats.NumObjects, Copies: stats.NumObjectCopies, Degraded: stats.NumObjectsDegraded,
				Misplaced: stats.NumObjectsMisplaced, Unfound: stats.NumObjectsUnfound,
			}
		}
	}

	var osdDF osdDFWire
	if p.optional(ctx, access, executor.BinaryCeph, "collect.overview_osd_df", []string{"osd", "df", "--format", "json"}, &osdDF) {
		var total uint64
		count := 0
		for _, node := range osdDF.Nodes {
			if node.ID == nil || *node.ID < 0 || node.PGs == nil || (node.Type != "" && node.Type != "osd") {
				continue
			}
			total += *node.PGs
			count++
		}
		if count > 0 {
			average := float64(total) / float64(count)
			overview.PGsPerOSD = &average
		}
	}

	var osdDump osdDumpWire
	flagsKnown := p.optional(ctx, access, executor.BinaryCeph, "collect.overview_osd_flags", []string{"osd", "dump", "--format", "json"}, &osdDump) && osdDump.Flags != nil
	if status := overviewScrubStatus(overview.PlacementGroups, splitOSDFlags(value(osdDump.Flags)), flagsKnown); status != "" {
		overview.ScrubStatus = &status
	}
}

func overviewScrubStatus(states []cephdomain.PGState, flags []string, flagsKnown bool) string {
	if flagsKnown {
		for _, flag := range flags {
			if flag == "noscrub" || flag == "nodeep-scrub" {
				return "disabled"
			}
		}
	}
	for _, state := range states {
		if strings.Contains(state.Name, "scrubbing") || strings.Contains(state.Name, "deep") {
			return "active"
		}
	}
	if flagsKnown {
		return "inactive"
	}
	return ""
}

type hostWire struct {
	Hostname         string                       `json:"hostname"`
	Addr             *string                      `json:"addr"`
	Status           *string                      `json:"status"`
	Labels           []string                     `json:"labels"`
	ServiceType      *string                      `json:"service_type"`
	Services         []cephdomain.HostService     `json:"services"`
	ServiceInstances []cephdomain.ServiceInstance `json:"service_instances"`
	Facts            map[string]any               `json:"facts"`
}

type hostFacts struct {
	System        *string
	Platform      *string
	Distro        *string
	KernelRelease *string
	KernelBuild   *string
	Arch          *string
	CPUModel      *string
	CPUCores      *int
	MemoryBytes   *uint64
}

type daemonWire struct {
	DaemonName         string  `json:"daemon_name"`
	DaemonType         string  `json:"daemon_type"`
	Hostname           *string `json:"hostname"`
	StatusDesc         *string `json:"status_desc"`
	Version            *string `json:"version"`
	ContainerImageName *string `json:"container_image_name"`
	CPUPercentage      *string `json:"cpu_percentage"`
	MemoryUsage        *uint64 `json:"memory_usage"`
	LastRefresh        *string `json:"last_refresh"`
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
	Features    struct {
		QuorumCon   string        `json:"quorum_con"`
		QuorumMon   []interface{} `json:"quorum_mon"`
		RequiredCon string        `json:"required_con"`
		RequiredMon []interface{} `json:"required_mon"`
	} `json:"features"`
	MonMap struct {
		FSID     string `json:"fsid"`
		Modified string `json:"modified"`
		Epoch    int    `json:"epoch"`
	} `json:"monmap"`
}

type perfCounterSchema struct {
	Description string `json:"description"`
	MetricType  string `json:"metric_type"`
	ValueType   string `json:"value_type"`
	Units       string `json:"units"`
	Type        int    `json:"type"`
	Priority    int    `json:"priority"`
}

const (
	perfCounterTime       = 0x1
	perfCounterLongRunAvg = 0x4
	perfCounterCounter    = 0x8
	defaultMgrStatsLimit  = 5
)

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
	factsByHost := p.collectDaemonHostFacts(ctx, access)
	rows := make([]Observation, 0, len(hosts)+len(daemons)+len(services)+len(mons.Mons)+1+len(managers.Standbys))
	for _, wire := range hosts {
		if strings.TrimSpace(wire.Hostname) == "" {
			return nil, fmt.Errorf("parse collect.host response: hostname is required")
		}
		facts := mergeHostFacts(hostFactsFromMap(wire.Facts), factsByHost[wire.Hostname])
		payload := cephdomain.Host{
			Hostname:         wire.Hostname,
			Address:          wire.Addr,
			Status:           wire.Status,
			Labels:           wire.Labels,
			ServiceType:      wire.ServiceType,
			Services:         wire.Services,
			ServiceInstances: wire.ServiceInstances,
			System:           facts.System,
			Platform:         facts.Platform,
			Distro:           facts.Distro,
			KernelRelease:    facts.KernelRelease,
			KernelBuild:      facts.KernelBuild,
			Arch:             facts.Arch,
			CPUModel:         facts.CPUModel,
			CPUCores:         facts.CPUCores,
			MemoryBytes:      facts.MemoryBytes,
		}
		rows = append(rows, Observation{Kind: "host", NaturalKey: wire.Hostname, Name: wire.Hostname, Status: value(wire.Status), Source: "ceph_cli", Payload: payload, ObservedAt: now})
	}
	for _, wire := range daemons {
		if strings.TrimSpace(wire.DaemonName) == "" || strings.TrimSpace(wire.DaemonType) == "" {
			return nil, fmt.Errorf("parse collect.daemon response: daemon_name and daemon_type are required")
		}
		payload := cephdomain.Daemon{
			Name:           wire.DaemonName,
			Type:           wire.DaemonType,
			Hostname:       wire.Hostname,
			Status:         wire.StatusDesc,
			Version:        cephdomain.NormalizeVersionPointer(wire.Version),
			ContainerImage: wire.ContainerImageName,
			CPUPercentage:  wire.CPUPercentage,
			MemoryUsage:    wire.MemoryUsage,
			LastRefresh:    wire.LastRefresh,
		}
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
	statusPayload := cephdomain.MonitorStatus{
		FSID: quorum.MonMap.FSID, Modified: quorum.MonMap.Modified, Epoch: quorum.MonMap.Epoch,
		QuorumCon: quorum.Features.QuorumCon, QuorumMon: interfaceStrings(quorum.Features.QuorumMon),
		RequiredCon: quorum.Features.RequiredCon, RequiredMon: interfaceStrings(quorum.Features.RequiredMon),
	}
	rows = append(rows, Observation{Kind: "mon_status", NaturalKey: "status", Name: "status", Source: "ceph_cli", Payload: statusPayload, ObservedAt: now})
	perfPriority, perfAvailable := p.collectMgrStatsThreshold(ctx, access)
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
		var counters []Observation
		var openSessions any
		if perfAvailable {
			counters, openSessions = p.collectMonitorPerfCounters(ctx, access, wire.Name, perfPriority, now)
			rows = append(rows, counters...)
		}
		payload := cephdomain.Monitor{Name: wire.Name, Rank: wire.Rank, Address: address, InQuorum: inQuorum, OpenSessions: openSessions}
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

func (p *NativeProvider) collectMgrStatsThreshold(ctx context.Context, access ClusterAccess) (int, bool) {
	if p.Executor == nil {
		markCollectionUnavailable(ctx, "collect.mon_perf.threshold")
		return 0, false
	}
	result, err := p.run(ctx, access, "collect.mon_perf.threshold", 15*time.Second, "config", "get", "mgr", "mgr_stats_threshold")
	if err != nil {
		markCollectionUnavailable(ctx, "collect.mon_perf.threshold")
		return 0, false
	}
	value := strings.TrimSpace(string(result))
	if value == "" {
		return defaultMgrStatsLimit, true
	}
	threshold, err := strconv.Atoi(value)
	if err != nil || threshold < 0 || threshold > 11 {
		markCollectionUnavailable(ctx, "collect.mon_perf.threshold")
		return 0, false
	}
	return threshold, true
}

func (p *NativeProvider) collectMonitorPerfCounters(ctx context.Context, access ClusterAccess, monitor string, minimumPriority int, observedAt time.Time) ([]Observation, any) {
	var schema map[string]map[string]perfCounterSchema
	if err := p.runInto(ctx, access, "collect.mon_perf.schema", []string{"tell", "mon." + monitor, "perf", "schema", "--format", "json"}, &schema); err != nil {
		return nil, nil
	}
	var values map[string]map[string]any
	if err := p.runInto(ctx, access, "collect.mon_perf.dump", []string{"tell", "mon." + monitor, "perf", "dump", "--format", "json"}, &values); err != nil {
		return nil, nil
	}
	groups := make([]string, 0, len(schema))
	for group := range schema {
		groups = append(groups, group)
	}
	sort.Strings(groups)
	rows := make([]Observation, 0)
	var openSessions any
	for _, group := range groups {
		names := make([]string, 0, len(schema[group]))
		for name := range schema[group] {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			definition := schema[group][name]
			if definition.Priority < minimumPriority {
				continue
			}
			metricType := dashboardMetricType(definition.Type)
			rawValue := dashboardPerfRawValue(values[group][name], definition)
			qualifiedName := group + "." + name
			if qualifiedName == "mon.num_sessions" {
				openSessions = rawValue
			}
			payload := cephdomain.MonitorPerfCounter{Monitor: monitor, Name: qualifiedName, Description: definition.Description, Value: rawValue, RawValue: rawValue, Unit: dashboardPerfUnit(definition.Units, metricType), MetricType: metricType, ValueType: definition.ValueType, Priority: definition.Priority}
			rows = append(rows, Observation{Kind: "mon_perf_counter", NaturalKey: monitor + ":" + qualifiedName, Name: qualifiedName, ParentKind: "mon", ParentKey: monitor, Source: "ceph_cli", Payload: payload, ObservedAt: observedAt})
		}
	}
	return rows, openSessions
}

func dashboardPerfRawValue(value any, definition perfCounterSchema) any {
	pair, ok := value.(map[string]any)
	if !ok {
		return value
	}
	sum, ok := numberFloat(pair["sum"])
	if !ok {
		return value
	}
	if definition.Type&perfCounterTime != 0 || strings.HasPrefix(definition.ValueType, "real-") {
		return sum * float64(time.Second)
	}
	return sum
}

func dashboardMetricType(counterType int) string {
	if counterType&perfCounterLongRunAvg != 0 || counterType&perfCounterCounter != 0 {
		return "counter"
	}
	return "gauge"
}

func dashboardPerfUnit(unit, metricType string) string {
	if metricType != "counter" {
		return ""
	}
	if unit == "bytes" {
		return "B/s"
	}
	return "/s"
}

func numberFloat(value any) (float64, bool) {
	switch typed := value.(type) {
	case json.Number:
		parsed, err := typed.Float64()
		return parsed, err == nil
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	default:
		return 0, false
	}
}

func interfaceStrings(values []interface{}) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			result = append(result, typed)
		case json.Number:
			result = append(result, typed.String())
		case float64:
			result = append(result, strconv.FormatFloat(typed, 'f', -1, 64))
		}
	}
	return result
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
	Flags *string `json:"flags"`
	OSDs  []struct {
		OSD int `json:"osd"`
		Up  int `json:"up"`
		In  int `json:"in"`
	} `json:"osds"`
}
type poolWire struct {
	Raw                 map[string]any  `json:"-"`
	Pool                int64           `json:"pool"`
	PoolID              int64           `json:"pool_id"`
	PoolName            string          `json:"pool_name"`
	Type                int             `json:"type"`
	Size                *int64          `json:"size"`
	MinSize             *int64          `json:"min_size"`
	PGNum               *int64          `json:"pg_num"`
	PGPNum              *int64          `json:"pg_placement_num"`
	PGAutoscaleMode     *string         `json:"pg_autoscale_mode"`
	ApplicationMetadata map[string]any  `json:"application_metadata"`
	CrushRule           json.RawMessage `json:"crush_rule"`
	FlagsNames          string          `json:"flags_names"`
	Options             map[string]any  `json:"options"`
	QuotaMaxBytes       *int64          `json:"quota_max_bytes"`
	QuotaMaxObjects     *int64          `json:"quota_max_objects"`
	Quotas              poolQuotaWire   `json:"quotas"`
}
type poolQuotaWire struct {
	MaxBytes   *int64 `json:"max_bytes"`
	MaxObjects *int64 `json:"max_objects"`
}
type poolGetQuotaWire struct {
	QuotaMaxBytes   *int64 `json:"quota_max_bytes"`
	QuotaMaxObjects *int64 `json:"quota_max_objects"`
}

func (wire *poolWire) UnmarshalJSON(data []byte) error {
	type poolWireAlias poolWire
	raw, err := rawMap(data)
	if err != nil {
		return err
	}
	var decoded poolWireAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*wire = poolWire(decoded)
	if wire.Pool == 0 {
		wire.Pool = wire.PoolID
	}
	wire.Raw = raw
	return nil
}

type fsDumpWire struct {
	Standbys []struct {
		Name string `json:"name"`
	} `json:"standbys"`
	Filesystems []struct {
		MDSMap struct {
			FSName       string           `json:"fs_name"`
			ID           int64            `json:"id"`
			MaxMDS       *int64           `json:"max_mds"`
			MetadataPool *int64           `json:"metadata_pool"`
			DataPools    []int64          `json:"data_pools"`
			In           []int64          `json:"in"`
			Up           map[string]int64 `json:"up"`
			Info         map[string]struct {
				Name  string `json:"name"`
				Rank  int    `json:"rank"`
				State string `json:"state"`
			} `json:"info"`
		} `json:"mdsmap"`
	} `json:"filesystems"`
}
type rbdImageWire struct {
	Snapshot *string `json:"snapshot"`
	Name     string  `json:"image"`
	Size     *uint64 `json:"size"`
	Format   *int    `json:"format"`
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
	hosts := osdHosts(tree)
	crushPaths := osdCrushPaths(tree)
	var pools []poolWire
	if err := p.runInto(ctx, access, "collect.pool", []string{"osd", "pool", "ls", "detail", "--format", "json"}, &pools); err != nil {
		return nil, err
	}
	var fs fsDumpWire
	if err := p.runInto(ctx, access, "collect.fs", []string{"fs", "dump", "--format", "json"}, &fs); err != nil {
		return nil, err
	}
	rows := []Observation{}
	rows = append(rows, Observation{Kind: "osd_flag", NaturalKey: "flags", Name: "flags", Source: "ceph_cli", Payload: map[string]any{"flags": splitOSDFlags(value(dump.Flags))}, ObservedAt: now})
	for _, node := range tree.Nodes {
		if node.Type != "osd" {
			continue
		}
		if node.ID < 0 || strings.TrimSpace(node.Name) == "" {
			return nil, fmt.Errorf("parse collect.osd_tree response: OSD id and name are required")
		}
		state := states[node.ID]
		up, in := state[0], state[1]
		payload := cephdomain.OSD{ID: node.ID, Name: node.Name, Status: node.Status, Up: &up, In: &in, Weight: node.CrushWeight, DeviceClass: node.DeviceClass, Host: hosts[node.ID], CrushPath: crushPaths[node.ID]}
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
		quota := p.collectPoolQuota(ctx, access, wire.PoolName)
		payload := cephdomain.Pool{
			Name: wire.PoolName, ID: wire.Pool, Type: kind, Size: wire.Size, MinSize: wire.MinSize, PGNum: wire.PGNum, PGPNum: wire.PGPNum,
			PGAutoscaleMode: wire.PGAutoscaleMode, Applications: poolApplications(wire.ApplicationMetadata), ApplicationMetadata: wire.ApplicationMetadata,
			CrushRule: rawTextPointer(wire.CrushRule), Flags: poolFlagNames(wire.FlagsNames), CompressionMode: poolCompressionMode(wire.Options), CompressionAlgorithm: poolOptionStringPointer(wire.Options, "compression_algorithm"),
			CompressionMinBlobSize: poolOptionInt64Pointer(wire.Options, "compression_min_blob_size"), CompressionMaxBlobSize: poolOptionInt64Pointer(wire.Options, "compression_max_blob_size"), CompressionRequiredRatio: poolOptionFloat64Pointer(wire.Options, "compression_required_ratio"),
			QuotaMaxBytes: firstInt64Pointer(quota.QuotaMaxBytes, wire.QuotaMaxBytes, wire.Quotas.MaxBytes), QuotaMaxObjects: firstInt64Pointer(quota.QuotaMaxObjects, wire.QuotaMaxObjects, wire.Quotas.MaxObjects),
			RBDMirroring: p.collectPoolMirroringMode(ctx, access, wire.PoolName),
			RawDetail:    poolRawDetail(wire.Raw, kind), Configuration: p.collectPoolConfiguration(ctx, access, wire.PoolName),
		}
		rows = append(rows, Observation{Kind: "pool", NaturalKey: wire.PoolName, Name: wire.PoolName, Status: "available", Source: "ceph_cli", Payload: payload, ObservedAt: now})
	}
	for _, wire := range fs.Filesystems {
		m := wire.MDSMap
		if strings.TrimSpace(m.FSName) == "" {
			return nil, fmt.Errorf("parse collect.fs response: fs_name is required")
		}
		payload := cephdomain.Filesystem{
			Name: m.FSName, ID: m.ID, MaxMDS: m.MaxMDS, MetadataPool: m.MetadataPool,
			DataPools: m.DataPools, In: m.In, Up: m.Up,
		}
		rows = append(rows, Observation{Kind: "filesystem", NaturalKey: m.FSName, Name: m.FSName, Status: "available", Source: "ceph_cli", Payload: payload, ObservedAt: now})
		var subvolumes []namedWire
		if err := p.runBinaryInto(ctx, access, executor.BinaryCeph, "collect.cephfs_subvolume", []string{"fs", "subvolume", "ls", m.FSName, "--format", "json"}, &subvolumes); err == nil {
			for _, subvolume := range subvolumes {
				payload := map[string]any{"filesystem": m.FSName, "name": subvolume.Name}
				var details map[string]any
				if err := p.runBinaryInto(ctx, access, executor.BinaryCeph, "collect.cephfs_subvolume", []string{"fs", "subvolume", "info", m.FSName, subvolume.Name, "--format", "json"}, &details); err == nil {
					for key, value := range details {
						payload[key] = value
					}
				}
				rows = append(rows, Observation{Kind: "subvolume", NaturalKey: m.FSName + "/" + subvolume.Name, ParentKind: "filesystem", ParentKey: m.FSName, Name: subvolume.Name, Status: "available", Source: "ceph_cli", Payload: payload, ObservedAt: now})
			}
		}
	}
	for _, pool := range pools {
		var images []rbdImageWire
		if err := p.runBinaryInto(ctx, access, executor.BinaryRBD, "collect.rbd_image", []string{"ls", "--long", pool.PoolName, "--format", "json"}, &images); err != nil {
			continue
		}
		if images == nil {
			markCollectionUnavailable(ctx, "collect.rbd_image")
		}
		for _, image := range images {
			if image.Snapshot != nil {
				continue
			}
			if strings.TrimSpace(image.Name) == "" {
				return nil, fmt.Errorf("parse collect.rbd_image response: image name is required")
			}
			spec := pool.PoolName + "/" + image.Name
			key := base64.RawURLEncoding.EncodeToString([]byte(spec))
			payload := cephdomain.RBDImage{ImagePath: spec, ImageSpec: key, Pool: pool.PoolName, Name: image.Name, SizeBytes: image.Size, Format: image.Format}
			p.enrichRBDImage(ctx, access, spec, &payload)
			rows = append(rows, Observation{Kind: "rbd_image", NaturalKey: key, Name: image.Name, Status: "available", Source: "rbd_cli", Payload: payload, ObservedAt: now})
		}
	}
	var realms rgwRealmWire
	if err := p.runBinaryInto(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_status", []string{"realm", "list", "--format", "json"}, &realms); err == nil {
		payload := cephdomain.RGWStatus{Realms: realms.Realms}
		var limits map[string]any
		if p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_global_ratelimit", []string{"global", "ratelimit", "get", "--format", "json"}, &limits) {
			valid := true
			for _, scope := range []string{"user_ratelimit", "bucket_ratelimit", "anonymous_ratelimit"} {
				if _, ok := limits[scope].(map[string]any); !ok {
					valid = false
				}
			}
			if valid {
				payload.GlobalRateLimit = limits
			} else {
				markCollectionUnavailable(ctx, "collect.rgw_global_ratelimit")
			}
		}
		rows = append(rows, Observation{Kind: "rgw_status", NaturalKey: "status", Name: "status", Status: "available", Source: "rgw_admin", Payload: payload, ObservedAt: now})
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
	var rawDevices []json.RawMessage
	if err := p.runInto(ctx, access, "collect.device", []string{"orch", "device", "ls", "--wide", "--refresh", "--format", "json"}, &rawDevices); err != nil {
		return nil, err
	}
	devices, err := normalizeDeviceWires(rawDevices)
	if err != nil {
		return nil, err
	}
	rows := make([]Observation, 0, len(devices))
	for _, wire := range devices {
		key := wire.Hostname + ":" + wire.Path
		payload := cephdomain.Device{ID: wire.DeviceID, Hostname: wire.Hostname, Path: wire.Path, Available: wire.Available, RejectedReasons: wire.RejectedReasons, DeviceType: wire.DeviceType, Model: wire.Model, Vendor: wire.Vendor, Serial: wire.Serial, SizeBytes: wire.SizeBytes, Rotational: wire.Rotational, Metadata: wire.Metadata}
		rows = append(rows, Observation{Kind: "device", NaturalKey: key, ParentKind: "host", ParentKey: wire.Hostname, Name: key, Source: "ceph_cli", Payload: payload, ObservedAt: now})
	}
	return rows, nil
}

func normalizeDeviceWires(rawDevices []json.RawMessage) ([]deviceWire, error) {
	devices := make([]deviceWire, 0, len(rawDevices))
	for _, raw := range rawDevices {
		record, err := rawMap(raw)
		if err != nil {
			return nil, fmt.Errorf("parse collect.device response: %w", err)
		}
		devices = append(devices, normalizeDeviceRecord(record, "")...)
	}
	return devices, nil
}

func normalizeDeviceRecord(record map[string]any, inheritedHost string) []deviceWire {
	host := firstString(record, "hostname", "host")
	if host == "" && hasArray(record, "devices") {
		host = firstString(record, "name")
	}
	if host == "" {
		host = inheritedHost
	}
	if children, ok := record["devices"].([]any); ok {
		devices := []deviceWire{}
		for _, child := range children {
			childRecord, ok := child.(map[string]any)
			if !ok {
				continue
			}
			devices = append(devices, normalizeDeviceRecord(childRecord, host)...)
		}
		return devices
	}
	device, ok := buildDeviceWire(record, host)
	if !ok {
		return nil
	}
	return []deviceWire{device}
}

func buildDeviceWire(record map[string]any, host string) (deviceWire, bool) {
	sysAPI, _ := record["sys_api"].(map[string]any)
	hostname := firstNonEmpty(firstString(record, "hostname", "host"), host)
	path := firstNonEmpty(firstString(record, "path"), firstString(sysAPI, "path"))
	if strings.TrimSpace(hostname) == "" || strings.TrimSpace(path) == "" {
		return deviceWire{}, false
	}
	size := firstUintPointer(record, "size_bytes", "size")
	if size == nil {
		size = firstUintPointer(sysAPI, "size_bytes", "size")
	}
	rotational := firstBoolPointer(record, "rotational")
	if rotational == nil {
		rotational = firstBoolPointer(sysAPI, "rotational")
	}
	metadata := stringMap(record["metadata"])
	for key, value := range stringMap(sysAPI) {
		if metadata == nil {
			metadata = map[string]string{}
		}
		metadata["sys_api."+key] = value
	}
	deviceType := firstStringPointer(record, "device_type", "human_readable_type", "crush_device_class", "type")
	if deviceType == nil {
		deviceType = firstStringPointer(sysAPI, "human_readable_type", "crush_device_class", "type")
	}
	model := firstStringPointer(record, "model")
	if model == nil {
		model = firstStringPointer(sysAPI, "model")
	}
	vendor := firstStringPointer(record, "vendor")
	if vendor == nil {
		vendor = firstStringPointer(sysAPI, "vendor")
	}
	serial := firstStringPointer(record, "serial")
	if serial == nil {
		serial = firstStringPointer(sysAPI, "serial")
	}
	return deviceWire{
		DeviceID:        firstNonEmpty(firstString(record, "device_id", "id", "devid"), hostname+":"+path),
		Hostname:        hostname,
		Path:            path,
		Available:       boolValue(record["available"]),
		RejectedReasons: stringSlice(record["rejected_reasons"]),
		DeviceType:      deviceType,
		Model:           model,
		Vendor:          vendor,
		Serial:          serial,
		SizeBytes:       size,
		Rotational:      rotational,
		Metadata:        metadata,
	}, true
}

func splitOSDFlags(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	flags := make([]string, 0, len(parts))
	for _, part := range parts {
		flag := strings.TrimSpace(part)
		if flag != "" {
			flags = append(flags, flag)
		}
	}
	return flags
}

func osdHosts(tree osdTreeWire) map[int]*string {
	hosts := map[int]*string{}
	for _, node := range tree.Nodes {
		if node.Type != "host" || strings.TrimSpace(node.Name) == "" {
			continue
		}
		host := node.Name
		for _, child := range node.Children {
			hosts[child] = &host
		}
	}
	return hosts
}

func osdCrushPaths(tree osdTreeWire) map[int]map[string]string {
	nodes := make(map[int]struct {
		name     string
		kind     string
		children []int
	}, len(tree.Nodes))
	for _, node := range tree.Nodes {
		nodes[node.ID] = struct {
			name     string
			kind     string
			children []int
		}{name: strings.TrimSpace(node.Name), kind: strings.TrimSpace(node.Type), children: node.Children}
	}
	paths := map[int]map[string]string{}
	var walk func(int, map[string]string, map[int]struct{})
	walk = func(id int, inherited map[string]string, visiting map[int]struct{}) {
		if _, cycle := visiting[id]; cycle {
			return
		}
		node, ok := nodes[id]
		if !ok {
			return
		}
		path := make(map[string]string, len(inherited)+1)
		for kind, name := range inherited {
			path[kind] = name
		}
		if node.kind != "" && node.kind != "osd" && node.name != "" {
			path[node.kind] = node.name
		}
		if node.kind == "osd" {
			paths[id] = path
			return
		}
		nextVisiting := make(map[int]struct{}, len(visiting)+1)
		for current := range visiting {
			nextVisiting[current] = struct{}{}
		}
		nextVisiting[id] = struct{}{}
		for _, child := range node.children {
			walk(child, path, nextVisiting)
		}
	}
	for _, node := range tree.Nodes {
		if node.Type == "root" {
			walk(node.ID, nil, map[int]struct{}{})
		}
	}
	return paths
}

func rawTextPointer(raw json.RawMessage) *string {
	value := strings.TrimSpace(string(raw))
	if value == "" || value == "null" {
		return nil
	}
	if strings.HasPrefix(value, `"`) {
		var decoded string
		if err := json.Unmarshal(raw, &decoded); err == nil {
			value = strings.TrimSpace(decoded)
		}
	}
	if value == "" {
		return nil
	}
	return &value
}

func poolApplications(metadata map[string]any) []string {
	if len(metadata) == 0 {
		return nil
	}
	applications := make([]string, 0, len(metadata))
	for name := range metadata {
		if strings.TrimSpace(name) != "" {
			applications = append(applications, name)
		}
	}
	sort.Strings(applications)
	return applications
}

func poolCompressionMode(options map[string]any) *string {
	return poolOptionStringPointer(options, "compression_mode")
}

func poolOptionStringPointer(options map[string]any, key string) *string {
	if len(options) == 0 {
		return nil
	}
	value, ok := options[key].(string)
	if !ok || strings.TrimSpace(value) == "" {
		return nil
	}
	value = strings.TrimSpace(value)
	return &value
}

func poolOptionInt64Pointer(options map[string]any, key string) *int64 {
	if len(options) == 0 {
		return nil
	}
	switch value := options[key].(type) {
	case float64:
		integer := int64(value)
		if float64(integer) == value {
			return &integer
		}
	case json.Number:
		if integer, err := value.Int64(); err == nil {
			return &integer
		}
	case string:
		if integer, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64); err == nil {
			return &integer
		}
	}
	return nil
}

func poolOptionFloat64Pointer(options map[string]any, key string) *float64 {
	if len(options) == 0 {
		return nil
	}
	switch value := options[key].(type) {
	case float64:
		return &value
	case json.Number:
		if number, err := value.Float64(); err == nil {
			return &number
		}
	case string:
		if number, err := strconv.ParseFloat(strings.TrimSpace(value), 64); err == nil {
			return &number
		}
	}
	return nil
}

func poolRawDetail(raw map[string]any, poolType string) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	detail := make(map[string]any, len(raw)+2)
	for key, value := range raw {
		detail[key] = value
	}
	if poolType != "" {
		detail["type"] = poolType
	}
	if metadata, ok := detail["application_metadata"].(map[string]any); ok && len(metadata) > 0 {
		detail["application_metadata"] = strings.Join(poolApplications(metadata), ", ")
	}
	return detail
}

func poolFlagNames(value string) []string {
	var flags []string
	for _, flag := range strings.Split(value, ",") {
		if flag = strings.TrimSpace(flag); flag != "" {
			flags = append(flags, flag)
		}
	}
	return flags
}

func (p *NativeProvider) collectPoolConfiguration(ctx context.Context, access ClusterAccess, pool string) []cephdomain.PoolConfig {
	if strings.TrimSpace(pool) == "" {
		return nil
	}
	var raw any
	if !p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_pool_config", []string{"config", "pool", "list", pool, "--format", "json"}, &raw) {
		return nil
	}
	rows := poolConfigRows(raw)
	configs := make([]cephdomain.PoolConfig, 0, len(rows))
	for _, row := range rows {
		name := firstNonEmpty(textField(row, "name"), textField(row, "key"), textField(row, "option"))
		if name == "" {
			continue
		}
		configs = append(configs, cephdomain.PoolConfig{
			Name:        name,
			Value:       firstConfigValue(row, "value", "val", "default"),
			Source:      firstNonEmpty(textField(row, "source"), textField(row, "level"), textField(row, "who")),
			Description: firstNonEmpty(textField(row, "description"), textField(row, "desc"), textField(row, "help")),
		})
	}
	sort.Slice(configs, func(i, j int) bool { return configs[i].Name < configs[j].Name })
	return configs
}

func (p *NativeProvider) collectPoolQuota(ctx context.Context, access ClusterAccess, pool string) poolGetQuotaWire {
	if strings.TrimSpace(pool) == "" {
		return poolGetQuotaWire{}
	}
	var quota poolGetQuotaWire
	if !p.optional(ctx, access, executor.BinaryCeph, "collect.pool_quota", []string{"osd", "pool", "get-quota", pool, "--format", "json"}, &quota) {
		return poolGetQuotaWire{}
	}
	return quota
}

func poolConfigRows(raw any) []map[string]any {
	rows := objectList(raw)
	if len(rows) > 0 {
		return rows
	}
	record, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	rows = make([]map[string]any, 0, len(record))
	for name, value := range record {
		if nested, ok := value.(map[string]any); ok {
			row := make(map[string]any, len(nested)+1)
			for key, item := range nested {
				row[key] = item
			}
			row["name"] = name
			rows = append(rows, row)
		} else {
			rows = append(rows, map[string]any{"name": name, "value": value})
		}
	}
	return rows
}

func firstConfigValue(record map[string]any, keys ...string) any {
	for _, key := range keys {
		value, ok := record[key]
		if !ok || value == nil {
			continue
		}
		if text, ok := value.(string); ok && strings.TrimSpace(text) == "" {
			continue
		}
		return value
	}
	return ""
}

func rawMap(raw json.RawMessage) (map[string]any, error) {
	var value map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstString(record map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := stringValue(record[key]); value != "" {
			return value
		}
	}
	return ""
}

func firstStringPointer(record map[string]any, keys ...string) *string {
	value := firstString(record, keys...)
	if value == "" {
		return nil
	}
	return &value
}

func firstInt64Pointer(values ...*int64) *int64 {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return ""
	}
}

func hasArray(record map[string]any, key string) bool {
	value, ok := record[key].([]any)
	return ok && len(value) > 0
}

func boolValue(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case json.Number:
		return typed.String() != "0"
	case string:
		normalized := strings.TrimSpace(strings.ToLower(typed))
		return normalized == "true" || normalized == "1" || normalized == "yes"
	default:
		return false
	}
}

func firstBoolPointer(record map[string]any, keys ...string) *bool {
	for _, key := range keys {
		switch record[key].(type) {
		case bool, json.Number, string:
			value := boolValue(record[key])
			return &value
		}
	}
	return nil
}

func firstUintPointer(record map[string]any, keys ...string) *uint64 {
	for _, key := range keys {
		if value, ok := uintValue(record[key]); ok {
			return &value
		}
	}
	return nil
}

func uintValue(value any) (uint64, bool) {
	switch typed := value.(type) {
	case json.Number:
		if parsed, err := strconv.ParseUint(typed.String(), 10, 64); err == nil {
			return parsed, true
		}
		if parsed, err := strconv.ParseFloat(typed.String(), 64); err == nil && parsed >= 0 {
			return uint64(parsed), true
		}
	case float64:
		if typed >= 0 {
			return uint64(typed), true
		}
	case string:
		if parsed, err := strconv.ParseUint(strings.TrimSpace(typed), 10, 64); err == nil {
			return parsed, true
		}
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64); err == nil && parsed >= 0 {
			return uint64(parsed), true
		}
	}
	return 0, false
}

func stringSlice(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if text := stringValue(item); text != "" {
			result = append(result, text)
		}
	}
	return result
}

func stringMap(value any) map[string]string {
	record, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	result := map[string]string{}
	for key, item := range record {
		if text := stringValue(item); text != "" {
			result[key] = text
		}
	}
	return result
}

type configValueWire struct {
	Who           string  `json:"who"`
	Name          string  `json:"name"`
	Value         string  `json:"value"`
	Level         *string `json:"level"`
	Section       *string `json:"section"`
	LocationType  *string `json:"location_type"`
	LocationValue *string `json:"location_value"`
	Mask          *string `json:"mask"`
}

func (p *NativeProvider) collectConfiguration(ctx context.Context, access ClusterAccess) ([]Observation, error) {
	now := time.Now().UTC()
	var values []configValueWire
	if err := p.runInto(ctx, access, "collect.config", []string{"config", "dump", "--format", "json"}, &values); err != nil {
		return nil, err
	}
	rows := make([]Observation, 0, len(values))
	for _, wire := range values {
		who := configValueWho(wire)
		name := strings.TrimSpace(wire.Name)
		if who == "" || name == "" {
			return nil, fmt.Errorf("parse collect.config response: who and name are required")
		}
		key := who + ":" + name
		payload := cephdomain.ConfigValue{Who: who, Name: name, Value: wire.Value, Level: wire.Level, Section: wire.Section, LocationType: wire.LocationType, LocationValue: wire.LocationValue, Mask: wire.Mask}
		rows = append(rows, Observation{Kind: "config_value", NaturalKey: key, Name: name, Source: "ceph_cli", Payload: payload, ObservedAt: now})
	}
	rows = append(rows, p.collectConfigurationOptional(ctx, access, now)...)
	return rows, nil
}

func configValueWho(wire configValueWire) string {
	who := strings.TrimSpace(wire.Who)
	if who == "" && wire.Section != nil {
		who = strings.TrimSpace(*wire.Section)
	}
	locationType := trimStringPointer(wire.LocationType)
	locationValue := trimStringPointer(wire.LocationValue)
	scope := trimStringPointer(wire.Mask)
	if locationType != "" && locationValue != "" {
		scope = locationType + ":" + locationValue
	}
	if scope == "" {
		return who
	}
	if who == "" {
		return scope
	}
	return who + "/" + scope
}

func (p *NativeProvider) collectDaemonHostFacts(ctx context.Context, access ClusterAccess) map[string]hostFacts {
	result := map[string]hostFacts{}
	for _, command := range []struct {
		id   string
		args []string
	}{
		{"collect.osd_metadata", []string{"osd", "metadata", "--format", "json"}},
		{"collect.mon_metadata", []string{"mon", "metadata", "--format", "json"}},
		{"collect.mgr_metadata", []string{"mgr", "metadata", "--format", "json"}},
		{"collect.mds_metadata", []string{"mds", "metadata", "--format", "json"}},
	} {
		var rows []map[string]any
		if !p.optional(ctx, access, executor.BinaryCeph, command.id, command.args, &rows) {
			continue
		}
		for _, row := range rows {
			hostname := firstString(row, "hostname", "host")
			if hostname == "" {
				continue
			}
			result[hostname] = mergeHostFacts(result[hostname], hostFactsFromDaemonMetadata(row))
		}
	}
	return result
}

func hostFactsFromMap(facts map[string]any) hostFacts {
	return hostFacts{
		System:        hostFactString(facts, "system", "os", "distro", "distribution"),
		Platform:      hostFactString(facts, "platform"),
		Distro:        hostFactString(facts, "distro"),
		KernelRelease: hostFactString(facts, "kernel_release", "kernel"),
		KernelBuild:   hostFactString(facts, "kernel_build"),
		Arch:          hostFactString(facts, "arch", "machine"),
		CPUModel:      hostFactString(facts, "cpu_model", "processor_model", "model_name"),
		CPUCores:      hostFactInt(facts, "cpu_cores", "cpu_count", "processor_count", "cpus"),
		MemoryBytes:   hostFactBytes(facts, "memory_bytes", "memory_total", "mem_total", "mem_total_kb"),
	}
}

func hostFactsFromDaemonMetadata(metadata map[string]any) hostFacts {
	return hostFacts{
		System:        hostFactString(metadata, "system", "distro_description", "distro", "os"),
		Platform:      hostFactString(metadata, "platform", "os"),
		Distro:        hostFactString(metadata, "distro"),
		KernelRelease: hostFactString(metadata, "kernel_release", "kernel_version", "kernel", "kernel_description"),
		KernelBuild:   hostFactString(metadata, "kernel_build", "kernel_description"),
		Arch:          hostFactString(metadata, "arch"),
		CPUModel:      hostFactString(metadata, "cpu_model", "cpu"),
		CPUCores:      hostFactInt(metadata, "cpu_cores", "cpu_count", "cpus"),
		MemoryBytes:   hostFactBytes(metadata, "memory_bytes", "mem_total_kb", "memory_total", "mem_total"),
	}
}

func mergeHostFacts(primary, fallback hostFacts) hostFacts {
	if primary.System == nil {
		primary.System = fallback.System
	}
	if primary.Platform == nil {
		primary.Platform = fallback.Platform
	}
	if primary.Distro == nil {
		primary.Distro = fallback.Distro
	}
	if primary.KernelRelease == nil {
		primary.KernelRelease = fallback.KernelRelease
	}
	if primary.KernelBuild == nil {
		primary.KernelBuild = fallback.KernelBuild
	}
	if primary.Arch == nil {
		primary.Arch = fallback.Arch
	}
	if primary.CPUModel == nil {
		primary.CPUModel = fallback.CPUModel
	}
	if primary.CPUCores == nil {
		primary.CPUCores = fallback.CPUCores
	}
	if primary.MemoryBytes == nil {
		primary.MemoryBytes = fallback.MemoryBytes
	}
	return primary
}

func trimStringPointer(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func hostFactString(facts map[string]any, keys ...string) *string {
	for _, key := range keys {
		value := strings.TrimSpace(textValue(hostFactValue(facts, key)))
		if value != "" {
			return &value
		}
	}
	return nil
}

func hostFactInt(facts map[string]any, keys ...string) *int {
	for _, key := range keys {
		switch value := hostFactValue(facts, key).(type) {
		case int:
			return &value
		case int64:
			parsed := int(value)
			return &parsed
		case float64:
			parsed := int(value)
			return &parsed
		case json.Number:
			if parsed, err := value.Int64(); err == nil {
				result := int(parsed)
				return &result
			}
		case string:
			if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
				return &parsed
			}
		}
	}
	return nil
}

func hostFactBytes(facts map[string]any, keys ...string) *uint64 {
	for _, key := range keys {
		raw := hostFactValue(facts, key)
		value, ok := hostFactUint64(raw)
		if !ok {
			continue
		}
		if strings.HasSuffix(strings.ToLower(key), "_kb") {
			value *= 1024
		}
		return &value
	}
	return nil
}

func hostFactValue(facts map[string]any, key string) any {
	if facts == nil {
		return nil
	}
	if value, ok := facts[key]; ok {
		return value
	}
	for existing, value := range facts {
		if strings.EqualFold(existing, key) {
			return value
		}
	}
	return nil
}

func hostFactUint64(raw any) (uint64, bool) {
	switch value := raw.(type) {
	case uint64:
		return value, true
	case int:
		if value >= 0 {
			return uint64(value), true
		}
	case int64:
		if value >= 0 {
			return uint64(value), true
		}
	case float64:
		if value >= 0 {
			return uint64(value), true
		}
	case json.Number:
		parsed, err := strconv.ParseUint(value.String(), 10, 64)
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
		return parsed, err == nil
	}
	return 0, false
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
	kinds := collectionFailureKinds[commandID]
	if strings.HasPrefix(commandID, "collect.mon_perf.") {
		kinds = collectionFailureKinds["collect.mon_perf"]
	}
	for _, kind := range kinds {
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
