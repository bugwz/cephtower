package clusterinspect

import (
	"context"
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/security"
)

var mdsCounterNames = []string{
	"mds_server.handle_client_request", "mds_log.ev", "mds_cache.num_strays",
	"mds.exported", "mds.exported_inodes", "mds.imported", "mds.imported_inodes",
	"mds.inodes", "mds.caps", "mds.subtrees", "mds_mem.ino",
	"mds_mem.dn", "mds_mem.dir", "mds_mem.cap", "mds_sessions.session_count", "mds_log.replay",
}

var mdsDaemonNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,255}$`)

type MDSCounterSample struct {
	Name  string  `json:"name"`
	Value *string `json:"value"`
}

type MDSPerformance struct {
	Name       string             `json:"name"`
	GID        string             `json:"gid"`
	Rank       int                `json:"rank"`
	State      string             `json:"state"`
	Counters   []MDSCounterSample `json:"counters"`
	Error      string             `json:"error,omitempty"`
	ObservedAt time.Time          `json:"observed_at"`
}

type CephFSPerformance struct {
	Filesystem string           `json:"filesystem"`
	Items      []MDSPerformance `json:"items"`
	ObservedAt time.Time        `json:"observed_at"`
}

func (s *Service) CephFSCounters(ctx context.Context, clusterID uint64, filesystem string) (CephFSPerformance, error) {
	if !cephFSNamePattern.MatchString(filesystem) {
		return CephFSPerformance{}, invalid("invalid filesystem name")
	}
	var fs struct {
		MDSMap struct {
			Name string `json:"fs_name"`
			Info map[string]struct {
				Name  string      `json:"name"`
				GID   json.Number `json:"gid"`
				Rank  *int        `json:"rank"`
				State string      `json:"state"`
			} `json:"info"`
		} `json:"mdsmap"`
	}
	if err := s.read(ctx, clusterID, "cephfs.performance.map", []string{"fs", "get", filesystem, "--format", "json"}, &fs); err != nil {
		return CephFSPerformance{}, err
	}
	if fs.MDSMap.Name != filesystem || fs.MDSMap.Info == nil || len(fs.MDSMap.Info) > 256 {
		return CephFSPerformance{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned an invalid filesystem MDS map"}
	}
	items := make([]MDSPerformance, 0, len(fs.MDSMap.Info))
	seen := map[string]bool{}
	for _, daemon := range fs.MDSMap.Info {
		if !mdsDaemonNamePattern.MatchString(daemon.Name) || seen[daemon.Name] || daemon.GID == "" || daemon.Rank == nil || daemon.State == "" {
			return CephFSPerformance{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned an invalid MDS identity"}
		}
		if _, err := strconv.ParseUint(daemon.GID.String(), 10, 64); err != nil {
			return CephFSPerformance{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned an invalid MDS gid"}
		}
		seen[daemon.Name] = true
		items = append(items, MDSPerformance{Name: daemon.Name, GID: daemon.GID.String(), Rank: *daemon.Rank, State: daemon.State, Counters: []MDSCounterSample{}})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	for i := range items {
		var counters map[string]map[string]json.RawMessage
		err := s.read(ctx, clusterID, "cephfs.performance.dump", []string{"tell", "mds." + items[i].Name, "perf", "dump", "--format", "json"}, &counters)
		items[i].ObservedAt = time.Now().UTC()
		if err != nil {
			if ctx.Err() != nil {
				return CephFSPerformance{}, ctx.Err()
			}
			items[i].Error = security.Redact(err.Error())
			continue
		}
		if len(counters) == 0 {
			items[i].Error = "Ceph returned empty MDS counters"
			continue
		}
		for _, name := range mdsCounterNames {
			parts := strings.SplitN(name, ".", 2)
			var value *string
			if raw := counters[parts[0]][parts[1]]; len(raw) > 0 && string(raw) != "null" {
				var number json.Number
				if json.Unmarshal(raw, &number) != nil || len(raw) == 0 || raw[0] == '"' {
					items[i].Error = "Ceph returned a non-numeric MDS counter"
					break
				}
				if _, err := strconv.ParseUint(number.String(), 10, 64); err != nil {
					items[i].Error = "Ceph returned an invalid unsigned MDS counter"
					break
				}
				text := number.String()
				value = &text
			}
			items[i].Counters = append(items[i].Counters, MDSCounterSample{Name: name, Value: value})
		}
		if items[i].Error != "" {
			items[i].Counters = []MDSCounterSample{}
		}
	}
	return CephFSPerformance{Filesystem: filesystem, Items: items, ObservedAt: time.Now().UTC()}, nil
}
