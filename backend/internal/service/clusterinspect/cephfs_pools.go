package clusterinspect

import (
	"context"
	"encoding/json"
	"math/big"
	"strconv"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

type CephFSPoolUsage struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Stored    *string `json:"stored"`
	Available *string `json:"available"`
	Size      *string `json:"size"`
	BytesUsed *string `json:"bytes_used"`
	Error     string  `json:"error,omitempty"`
}

type CephFSPoolList struct {
	Filesystem string            `json:"filesystem"`
	Items      []CephFSPoolUsage `json:"items"`
	ObservedAt time.Time         `json:"observed_at"`
}

func (s *Service) CephFSPools(ctx context.Context, clusterID uint64, filesystem string) (CephFSPoolList, error) {
	if !cephFSNamePattern.MatchString(filesystem) {
		return CephFSPoolList{}, invalid("invalid filesystem name")
	}
	var fs struct {
		MDSMap struct {
			Name     string  `json:"fs_name"`
			Metadata *int64  `json:"metadata_pool"`
			Data     []int64 `json:"data_pools"`
		} `json:"mdsmap"`
	}
	if err := s.read(ctx, clusterID, "cephfs.pools.map", []string{"fs", "get", filesystem, "--format", "json"}, &fs); err != nil {
		return CephFSPoolList{}, err
	}
	if fs.MDSMap.Name != filesystem || fs.MDSMap.Metadata == nil || *fs.MDSMap.Metadata < 0 || fs.MDSMap.Data == nil {
		return CephFSPoolList{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned an invalid filesystem pool map"}
	}
	ids := append([]int64{*fs.MDSMap.Metadata}, fs.MDSMap.Data...)
	seen := map[int64]bool{}
	for _, id := range ids {
		if id < 0 || seen[id] {
			return CephFSPoolList{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned invalid or duplicate filesystem pool IDs"}
		}
		seen[id] = true
	}
	type poolStats struct {
		Stored    json.Number `json:"stored"`
		Available json.Number `json:"max_avail"`
		BytesUsed json.Number `json:"bytes_used"`
	}
	type pool struct {
		ID    *int64    `json:"id"`
		Name  string    `json:"name"`
		Stats poolStats `json:"stats"`
	}
	var df struct {
		Pools []pool `json:"pools"`
	}
	if err := s.read(ctx, clusterID, "cephfs.pools.df", []string{"df", "detail", "--format", "json"}, &df); err != nil {
		return CephFSPoolList{}, err
	}
	if df.Pools == nil {
		return CephFSPoolList{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned an invalid pool usage list"}
	}
	byID := map[int64]pool{}
	for _, row := range df.Pools {
		if row.ID == nil || *row.ID < 0 {
			return CephFSPoolList{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned an invalid pool usage identity"}
		}
		if _, exists := byID[*row.ID]; exists {
			return CephFSPoolList{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned duplicate pool usage identities"}
		}
		byID[*row.ID] = row
	}
	items := make([]CephFSPoolUsage, 0, len(ids))
	for index, id := range ids {
		item := CephFSPoolUsage{ID: strconv.FormatInt(id, 10), Type: "data"}
		if index == 0 {
			item.Type = "metadata"
		}
		row, exists := byID[id]
		if !exists || row.Name == "" {
			item.Error = "pool is absent from the current usage response"
		} else {
			item.Name = row.Name
			item.Stored = unsignedPoolStat(row.Stats.Stored)
			item.Available = unsignedPoolStat(row.Stats.Available)
			item.BytesUsed = unsignedPoolStat(row.Stats.BytesUsed)
			if item.Stored == nil || item.Available == nil {
				item.Error = "logical pool usage statistics are missing or invalid"
			} else {
				stored, _ := new(big.Int).SetString(*item.Stored, 10)
				available, _ := new(big.Int).SetString(*item.Available, 10)
				total := new(big.Int).Add(stored, available).String()
				item.Size = &total
			}
		}
		items = append(items, item)
	}
	return CephFSPoolList{Filesystem: filesystem, Items: items, ObservedAt: time.Now().UTC()}, nil
}

func unsignedPoolStat(value json.Number) *string {
	if _, err := strconv.ParseUint(value.String(), 10, 64); err != nil {
		return nil
	}
	text := value.String()
	return &text
}
