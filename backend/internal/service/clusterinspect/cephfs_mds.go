package clusterinspect

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/security"
)

type CephFSMDSRank struct {
	Rank    string `json:"rank"`
	Name    string `json:"name"`
	GID     string `json:"gid"`
	State   string `json:"state"`
	Laggy   bool   `json:"laggy"`
	Version string `json:"version"`
}

type CephFSMDSStatus struct {
	Filesystem    string          `json:"filesystem"`
	Ranks         []CephFSMDSRank `json:"ranks"`
	Standbys      []CephFSMDSRank `json:"standbys"`
	MetadataError string          `json:"metadata_error,omitempty"`
	ObservedAt    time.Time       `json:"observed_at"`
}

type cephFSMDSIdentity struct {
	Name       string      `json:"name"`
	GID        json.Number `json:"gid"`
	Rank       *int        `json:"rank"`
	State      string      `json:"state"`
	LaggySince *string     `json:"laggy_since"`
}

func (s *Service) CephFSMDS(ctx context.Context, clusterID uint64, filesystem string) (CephFSMDSStatus, error) {
	if !cephFSNamePattern.MatchString(filesystem) {
		return CephFSMDSStatus{}, invalid("invalid filesystem name")
	}
	var dump struct {
		Filesystems []struct {
			MDSMap struct {
				Name string                       `json:"fs_name"`
				In   []int                        `json:"in"`
				Up   map[string]json.Number       `json:"up"`
				Info map[string]cephFSMDSIdentity `json:"info"`
			} `json:"mdsmap"`
		} `json:"filesystems"`
		Standbys []cephFSMDSIdentity `json:"standbys"`
	}
	bad := func(message string) (CephFSMDSStatus, error) {
		return CephFSMDSStatus{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: message}
	}
	if err := s.read(ctx, clusterID, "cephfs.mds.map", []string{"fs", "dump", "--format", "json"}, &dump); err != nil {
		return CephFSMDSStatus{}, err
	}
	if dump.Filesystems == nil || dump.Standbys == nil {
		return bad("Ceph returned an incomplete FS map")
	}
	selected := -1
	for i, fs := range dump.Filesystems {
		if fs.MDSMap.Name == filesystem {
			if selected >= 0 {
				return bad("Ceph returned duplicate filesystem identities")
			}
			selected = i
		}
	}
	if selected < 0 {
		return CephFSMDSStatus{}, &cephdomain.ActionError{Code: "resource_not_found", Message: "filesystem is absent from the current FS map"}
	}
	m := dump.Filesystems[selected].MDSMap
	if m.In == nil || m.Up == nil || m.Info == nil {
		return bad("Ceph returned an incomplete MDS rank map")
	}
	seenNames, seenGIDs := map[string]bool{}, map[string]bool{}
	validate := func(info cephFSMDSIdentity) bool {
		if !mdsDaemonNamePattern.MatchString(info.Name) || info.Rank == nil || info.State == "" || seenNames[info.Name] || seenGIDs[info.GID.String()] {
			return false
		}
		if _, err := strconv.ParseUint(info.GID.String(), 10, 64); err != nil {
			return false
		}
		seenNames[info.Name], seenGIDs[info.GID.String()] = true, true
		return true
	}
	for key, info := range m.Info {
		if key != "gid_"+info.GID.String() || !validate(info) {
			return bad("Ceph returned an invalid MDS identity")
		}
	}
	result := CephFSMDSStatus{Filesystem: filesystem, Ranks: []CephFSMDSRank{}, Standbys: []CephFSMDSRank{}, ObservedAt: time.Now().UTC()}
	seenRanks := map[int]bool{}
	row := func(rank string, info cephFSMDSIdentity) CephFSMDSRank {
		return CephFSMDSRank{Rank: rank, Name: info.Name, GID: info.GID.String(), State: strings.TrimPrefix(info.State, "up:"), Laggy: info.LaggySince != nil}
	}
	sort.Ints(m.In)
	for _, rank := range m.In {
		if rank < 0 || seenRanks[rank] {
			return bad("Ceph returned invalid or duplicate MDS ranks")
		}
		seenRanks[rank] = true
		gid, up := m.Up[fmt.Sprintf("mds_%d", rank)]
		if !up {
			result.Ranks = append(result.Ranks, CephFSMDSRank{Rank: strconv.Itoa(rank), State: "failed"})
			continue
		}
		info, exists := m.Info["gid_"+gid.String()]
		if !exists || *info.Rank != rank || info.State == "up:standby-replay" {
			return bad("Ceph returned inconsistent active MDS rank identities")
		}
		result.Ranks = append(result.Ranks, row(strconv.Itoa(rank), info))
	}
	for key := range m.Up {
		rank, err := strconv.Atoi(strings.TrimPrefix(key, "mds_"))
		if err != nil || key != fmt.Sprintf("mds_%d", rank) || !seenRanks[rank] {
			return bad("Ceph returned an unexpected active MDS rank")
		}
	}
	var replay []cephFSMDSIdentity
	for _, info := range m.Info {
		if info.State == "up:standby-replay" {
			replay = append(replay, info)
		}
	}
	sort.Slice(replay, func(i, j int) bool { return *replay[i].Rank < *replay[j].Rank })
	for _, info := range replay {
		if !seenRanks[*info.Rank] {
			return bad("Ceph returned an unassigned standby-replay rank")
		}
		result.Ranks = append(result.Ranks, row(strconv.Itoa(*info.Rank)+"-s", info))
	}
	for _, info := range dump.Standbys {
		if !validate(info) || *info.Rank != -1 {
			return bad("Ceph returned an invalid global standby identity")
		}
		result.Standbys = append(result.Standbys, row("—", info))
	}
	sort.Slice(result.Standbys, func(i, j int) bool { return result.Standbys[i].Name < result.Standbys[j].Name })
	var metadata []struct {
		Name    string `json:"name"`
		Version string `json:"ceph_version"`
	}
	if err := s.read(ctx, clusterID, "cephfs.mds.metadata", []string{"mds", "metadata", "--format", "json"}, &metadata); err != nil {
		if ctx.Err() != nil {
			return CephFSMDSStatus{}, ctx.Err()
		}
		result.MetadataError = security.Redact(err.Error())
	} else if metadata == nil {
		result.MetadataError = "Ceph returned an invalid MDS metadata list"
	} else {
		versions := map[string]string{}
		for _, item := range metadata {
			if item.Name == "" {
				result.MetadataError = "Ceph returned invalid MDS metadata identities"
				break
			}
			if _, exists := versions[item.Name]; exists {
				result.MetadataError = "Ceph returned duplicate MDS metadata identities"
				break
			}
			versions[item.Name] = item.Version
		}
		if result.MetadataError == "" {
			for i := range result.Ranks {
				result.Ranks[i].Version = versions[result.Ranks[i].Name]
			}
			for i := range result.Standbys {
				result.Standbys[i].Version = versions[result.Standbys[i].Name]
			}
		}
	}
	return result, nil
}
