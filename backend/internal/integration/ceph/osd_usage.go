package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"strings"
)

func (p *NativeProvider) collectOSDUsage(ctx context.Context, access ClusterAccess) map[int]*cephdomain.OSDUsage {
	var wire struct {
		Nodes []struct {
			ID          *int     `json:"id"`
			Type        string   `json:"type"`
			KB          *uint64  `json:"kb"`
			Used        *uint64  `json:"kb_used"`
			Available   *uint64  `json:"kb_avail"`
			Data        *uint64  `json:"kb_used_data"`
			Omap        *uint64  `json:"kb_used_omap"`
			Metadata    *uint64  `json:"kb_used_meta"`
			PGs         *uint64  `json:"pgs"`
			Utilization *float64 `json:"utilization"`
		} `json:"nodes"`
	}
	if !p.optional(ctx, access, executor.BinaryCeph, "collect.osd_usage", []string{"osd", "df", "--format", "json"}, &wire) {
		return nil
	}
	invalid := func() map[int]*cephdomain.OSDUsage { markCollectionUnavailable(ctx, "collect.osd_usage"); return nil }
	if wire.Nodes == nil {
		return invalid()
	}
	result := map[int]*cephdomain.OSDUsage{}
	for _, node := range wire.Nodes {
		if node.ID == nil || node.Type == "" || node.Type != strings.TrimSpace(node.Type) {
			return invalid()
		}
		if node.Type != "osd" {
			continue
		}
		if node.ID == nil || *node.ID < 0 || result[*node.ID] != nil {
			return invalid()
		}
		if node.Utilization != nil && (*node.Utilization < 0 || *node.Utilization > 100) {
			return invalid()
		}
		result[*node.ID] = &cephdomain.OSDUsage{KB: managerGID(node.KB), KBUsed: managerGID(node.Used), KBAvailable: managerGID(node.Available), KBData: managerGID(node.Data), KBOmap: managerGID(node.Omap), KBMetadata: managerGID(node.Metadata), PGs: managerGID(node.PGs), Utilization: node.Utilization}
	}
	return result
}
