package ceph

import "context"

func (p *NativeProvider) collectNFSClusterInfo(ctx context.Context, access ClusterAccess, name string) map[string]any {
	result := map[string]any{"name": name, "info_available": false}
	var info map[string]map[string]any
	if err := p.runInto(ctx, access, "collect.nfs_cluster_info", []string{"nfs", "cluster", "info", name, "--format", "json"}, &info); err != nil {
		return result
	}
	detail, ok := info[name]
	if !ok || detail == nil {
		return result
	}
	backends, ok := detail["backend"].([]any)
	if !ok {
		return result
	}
	for _, backend := range backends {
		if _, ok := backend.(map[string]any); !ok {
			return result
		}
	}
	for _, key := range []string{"virtual_ip", "port", "monitor_port", "ingress_mode", "backend"} {
		if value, exists := detail[key]; exists {
			result[key] = value
		}
	}
	result["info_available"] = true
	return result
}
