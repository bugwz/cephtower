package clusterinspect

import (
	"context"
	"encoding/json"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

type CrushMap struct {
	Nodes []map[string]any `json:"nodes"`
	Roots []int64          `json:"roots"`
}

func (s *Service) CrushMap(ctx context.Context, clusterID uint64) (CrushMap, error) {
	var result CrushMap
	if err := s.read(ctx, clusterID, "crush.map", []string{"osd", "tree", "--format", "json"}, &result); err != nil {
		return result, err
	}
	bad := func() (CrushMap, error) {
		return CrushMap{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned an invalid CRUSH tree"}
	}
	if result.Nodes == nil {
		return bad()
	}
	ids := map[int64]bool{}
	edges := map[int64][]int64{}
	parents := map[int64]bool{}
	order := []int64{}
	for _, node := range result.Nodes {
		number, ok := node["id"].(json.Number)
		if !ok {
			return bad()
		}
		id, err := number.Int64()
		if err != nil || ids[id] {
			return bad()
		}
		name, nameOK := node["name"].(string)
		kind, kindOK := node["type"].(string)
		if !nameOK || name == "" || !kindOK || kind == "" {
			return bad()
		}
		ids[id] = true
		order = append(order, id)
		if children, exists := node["children"]; exists {
			list, ok := children.([]any)
			if !ok {
				return bad()
			}
			for _, child := range list {
				n, ok := child.(json.Number)
				if !ok {
					return bad()
				}
				childID, err := n.Int64()
				if err != nil {
					return bad()
				}
				edges[id] = append(edges[id], childID)
				parents[childID] = true
			}
		}
	}
	state := map[int64]int{}
	var visit func(int64) bool
	visit = func(id int64) bool {
		if !ids[id] || state[id] == 1 {
			return false
		}
		if state[id] == 2 {
			return true
		}
		state[id] = 1
		for _, child := range edges[id] {
			if !visit(child) {
				return false
			}
		}
		state[id] = 2
		return true
	}
	result.Roots = []int64{}
	for _, id := range order {
		if !visit(id) {
			return bad()
		}
		if !parents[id] {
			result.Roots = append(result.Roots, id)
		}
	}
	return result, nil
}
