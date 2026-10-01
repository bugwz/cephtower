package ceph

import (
	"context"
	"encoding/json"
	"net"
)

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
	cleanBackends := make([]map[string]any, 0, len(backends))
	for _, backend := range backends {
		entry, ok := backend.(map[string]any)
		if !ok {
			return result
		}
		host, ok := entry["hostname"].(string)
		ip, ipOK := entry["ip"].(string)
		if !ok || host == "" || !ipOK || net.ParseIP(ip) == nil || !nfsInfoPort(entry["port"]) {
			return result
		}
		cleanBackends = append(cleanBackends, map[string]any{"hostname": host, "ip": ip, "port": entry["port"]})
	}
	if value := detail["virtual_ip"]; value != nil {
		ip, ok := value.(string)
		if !ok || net.ParseIP(ip) == nil {
			return result
		}
	}
	if value := detail["ingress_mode"]; value != nil {
		if _, ok := value.(string); !ok {
			return result
		}
	}
	if !nfsInfoPort(detail["port"]) || !nfsInfoPort(detail["monitor_port"]) {
		return result
	}
	for _, key := range []string{"virtual_ip", "port", "monitor_port", "ingress_mode"} {
		if value, exists := detail[key]; exists {
			result[key] = value
		}
	}
	result["backend"] = cleanBackends
	result["info_available"] = true
	return result
}

func nfsInfoPort(value any) bool {
	if value == nil {
		return true
	} // Native daemons may not report a port yet.
	number, ok := value.(json.Number)
	if !ok {
		return false
	}
	port, err := number.Int64()
	return err == nil && port > 0 && port <= 65535
}
