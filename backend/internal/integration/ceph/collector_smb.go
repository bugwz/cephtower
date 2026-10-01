package ceph

import "context"

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
