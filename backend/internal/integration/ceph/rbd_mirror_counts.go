package ceph

import (
	"encoding/json"
	"sort"
	"strconv"
)

func mirrorLeaderCounts(status map[string]any, pool poolWire, daemons any) map[string]any {
	list, ok := daemons.([]any)
	if !ok {
		return nil
	}
	var leader map[string]any
	for _, item := range list {
		daemon, ok := item.(map[string]any)
		if !ok {
			return nil
		}
		if daemon["leader"] == true {
			if leader != nil {
				return nil
			}
			leader = daemon
		}
	}
	if leader == nil || textField(leader, "instance_id") == "" || textField(leader, "service_id") == "" {
		return nil
	}
	services, _ := status["rbd-mirror"].(map[string]any)
	daemon, _ := services[textField(leader, "service_id")].(map[string]any)
	report, _ := daemon["status"].(map[string]any)
	var pools map[string]struct {
		Name       string                                `json:"name"`
		Leader     bool                                  `json:"leader"`
		Instance   string                                `json:"instance_id"`
		Namespaces map[string]map[string]json.RawMessage `json:"namespaces"`
	}
	if json.Unmarshal([]byte(textField(report, "json")), &pools) != nil {
		return nil
	}
	data, ok := pools[strconv.FormatInt(pool.Pool, 10)]
	if !ok || !data.Leader || data.Name != pool.PoolName || data.Instance != textField(leader, "instance_id") || data.Namespaces == nil {
		return nil
	}
	count := func(raw json.RawMessage) string {
		value := string(raw)
		if _, err := strconv.ParseUint(value, 10, 64); err != nil {
			return ""
		}
		return value
	}
	names := make([]string, 0, len(data.Namespaces))
	for name := range data.Namespaces {
		names = append(names, name)
	}
	sort.Strings(names)
	rows := make([]map[string]any, 0, len(names))
	for _, name := range names {
		row := map[string]any{"namespace": name}
		for _, field := range []string{"image_local_count", "image_remote_count"} {
			if value := count(data.Namespaces[name][field]); value != "" {
				row[field] = value
			}
		}
		rows = append(rows, row)
	}
	return map[string]any{"instance_id": data.Instance, "service_id": textField(leader, "service_id"), "namespaces": rows}
}
