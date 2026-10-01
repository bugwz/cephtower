package mutation

import (
	"bytes"
	"encoding/json"
	"io"
)

func isSMBAuthDelete(action string) bool {
	return action == "smb_join_auth.delete" || action == "smb_usersgroups.delete"
}

func smbAuthResourceType(action string) (string, string) {
	if action == "smb_join_auth.delete" {
		return "ceph.smb.join.auth", "auth_id"
	}
	return "ceph.smb.usersgroups", "users_groups_id"
}

func smbAuthDeleted(request Request, raw []byte) bool {
	var response struct {
		Resources []map[string]any `json:"resources"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if decoder.Decode(&response) != nil || response.Resources == nil || decoder.Decode(new(any)) != io.EOF {
		return false
	}
	resourceType, idField := smbAuthResourceType(request.Action)
	target := last(resourceTail(request.ResourceKey))
	seen := map[string]bool{}
	for _, item := range response.Resources {
		id, _ := item[idField].(string)
		if item["resource_type"] != resourceType || !smbResourceIDPattern.MatchString(id) || seen[id] || id == target {
			return false
		}
		seen[id] = true
	}
	return true
}
