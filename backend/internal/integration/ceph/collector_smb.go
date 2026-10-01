package ceph

import (
	"context"
	"time"
)

func (p *NativeProvider) collectSMBClusterInfo(ctx context.Context, access ClusterAccess, name string) map[string]any {
	result := map[string]any{"name": name, "info_available": false}
	var detail map[string]any
	if err := p.runInto(ctx, access, "collect.smb_cluster_info", []string{"smb", "show", "ceph.smb.cluster." + name, "--format", "json"}, &detail); err != nil {
		return result
	}
	if detail["resource_type"] != "ceph.smb.cluster" || detail["cluster_id"] != name {
		return result
	}
	auth, ok := detail["auth_mode"].(string)
	if !ok || (auth != "user" && auth != "active-directory") {
		return result
	}
	for _, key := range []string{"auth_mode", "intent", "domain_settings", "user_group_settings", "custom_dns", "public_addrs", "placement", "clustering"} {
		if value, exists := detail[key]; exists {
			result[key] = value
		}
	}
	result["info_available"] = true
	return result
}

func (p *NativeProvider) collectSMBShares(ctx context.Context, access ClusterAccess, cluster string, now time.Time) []Observation {
	var response struct {
		Resources []map[string]any `json:"resources"`
	}
	if err := p.runInto(ctx, access, "collect.smb_share", []string{"smb", "show", "ceph.smb.share." + cluster, "--results=full", "--format", "json"}, &response); err != nil {
		return nil
	}
	if response.Resources == nil {
		markCollectionUnavailable(ctx, "collect.smb_share")
		return nil
	}
	rows := make([]Observation, 0, len(response.Resources))
	seen := map[string]bool{}
	for _, item := range response.Resources {
		id, ok := item["share_id"].(string)
		if !ok || id == "" || seen[id] || item["resource_type"] != "ceph.smb.share" || item["cluster_id"] != cluster {
			markCollectionUnavailable(ctx, "collect.smb_share")
			return nil
		}
		seen[id] = true
		rows = append(rows, Observation{Kind: "smb_share", NaturalKey: opaquePair(cluster, id), ParentKind: "smb_cluster", ParentKey: cluster, Name: id, Status: "available", Source: "ceph_cli", Payload: item, ObservedAt: now})
	}
	return rows
}
