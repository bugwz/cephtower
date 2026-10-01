package ceph

import (
	"context"
	"regexp"
	"time"
)

var smbInventoryID = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,16}[a-zA-Z0-9])?$`)

// Passwords and raw authentication objects must never enter inventory, even if
// a server ignores the filter. Only explicitly allowed display fields remain.
func (p *NativeProvider) collectSMBAuthResources(ctx context.Context, access ClusterAccess, now time.Time) []Observation {
	var rows []Observation
	for _, spec := range []struct{ kind, resourceType, idField string }{
		{"smb_join_auth", "ceph.smb.join.auth", "auth_id"},
		{"smb_usersgroups", "ceph.smb.usersgroups", "users_groups_id"},
	} {
		commandID := "collect." + spec.kind
		var response struct {
			Resources []map[string]any `json:"resources"`
		}
		if err := p.runInto(ctx, access, commandID, []string{"smb", "show", spec.resourceType, "--results=full", "--password-filter=hidden", "--format", "json"}, &response); err != nil {
			continue
		}
		valid := response.Resources != nil
		seen := map[string]bool{}
		var batch []Observation
		for _, item := range response.Resources {
			id, _ := item[spec.idField].(string)
			if !smbInventoryID.MatchString(id) || seen[id] || item["resource_type"] != spec.resourceType {
				valid = false
				break
			}
			seen[id] = true
			payload := map[string]any{"resource_type": spec.resourceType, spec.idField: id}
			if auth, ok := item["auth"].(map[string]any); ok && spec.kind == "smb_join_auth" {
				if username, ok := auth["username"].(string); ok {
					payload["username"] = username
				}
			}
			if values, ok := item["values"].(map[string]any); ok && spec.kind == "smb_usersgroups" {
				if users, ok := values["users"].([]any); ok {
					payload["user_count"] = len(users)
					names := make([]string, 0, len(users))
					for _, user := range users {
						if object, ok := user.(map[string]any); ok {
							if name, ok := object["name"].(string); ok {
								names = append(names, name)
							}
						}
					}
					if len(names) == len(users) {
						payload["user_names"] = names
					}
				}
				if groups, ok := values["groups"].([]any); ok {
					names := make([]string, 0, len(groups))
					for _, group := range groups {
						if object, ok := group.(map[string]any); ok {
							if name, ok := object["name"].(string); ok {
								names = append(names, name)
							}
						}
					}
					payload["group_names"] = names
				}
			}
			for _, key := range []string{"intent", "linked_to_cluster"} {
				if value, exists := item[key]; exists && value != nil {
					text, ok := value.(string)
					if !ok || (key == "intent" && text != "present") || (key == "linked_to_cluster" && text != "" && !smbInventoryID.MatchString(text)) {
						valid = false
						break
					}
					payload[key] = text
				}
			}
			batch = append(batch, Observation{Kind: spec.kind, NaturalKey: id, Name: id, Status: "available", Source: "ceph_cli", Payload: payload, ObservedAt: now})
		}
		if !valid {
			markCollectionUnavailable(ctx, commandID)
			continue
		}
		rows = append(rows, batch...)
	}
	return rows
}

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
