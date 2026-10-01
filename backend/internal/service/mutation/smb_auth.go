package mutation

import (
	"bytes"
	"encoding/json"
	"io"
)

func isSMBAuthDelete(action string) bool {
	return action == "smb_join_auth.delete" || action == "smb_usersgroups.delete"
}

func isSMBJoinAuthWrite(action string) bool {
	return action == "smb_join_auth.create" || action == "smb_join_auth.update"
}

func isSMBAuthWrite(action string) bool {
	return isSMBJoinAuthWrite(action) || action == "smb_usersgroups.create"
}

func smbAuthResourceType(action string) (string, string) {
	if action == "smb_join_auth.delete" || isSMBJoinAuthWrite(action) {
		return "ceph.smb.join.auth", "auth_id"
	}
	return "ceph.smb.usersgroups", "users_groups_id"
}

func smbJoinAuthCreateJSON(p map[string]any) ([]byte, error) {
	name, _ := p["name"].(string)
	username, _ := p["username"].(string)
	password, _ := p["password"].(string)
	if !smbResourceIDPattern.MatchString(name) || username == "" || password == "" {
		return nil, invalid("valid SMB credential ID, username and password are required")
	}
	payload := map[string]any{"resource_type": "ceph.smb.join.auth", "auth_id": name, "intent": "present", "auth": map[string]any{"username": username, "password": password}}
	if linked, exists := p["linked_to_cluster"]; exists {
		id, ok := linked.(string)
		if !ok || !smbResourceIDPattern.MatchString(id) {
			return nil, invalid("invalid linked SMB cluster ID")
		}
		payload["linked_to_cluster"] = id
	}
	return json.Marshal(payload)
}

func smbJoinAuthCreated(request Request, raw []byte) bool {
	var response struct {
		Resources []map[string]any `json:"resources"`
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	if d.Decode(&response) != nil || response.Resources == nil || d.Decode(new(any)) != io.EOF {
		return false
	}
	seen := map[string]bool{}
	found := false
	for _, item := range response.Resources {
		id, _ := item["auth_id"].(string)
		if item["resource_type"] != "ceph.smb.join.auth" || !smbResourceIDPattern.MatchString(id) || seen[id] {
			return false
		}
		seen[id] = true
		if id != request.Parameters["name"] {
			continue
		}
		auth, ok := item["auth"].(map[string]any)
		if !ok || auth["username"] != request.Parameters["username"] {
			return false
		}
		want, _ := request.Parameters["linked_to_cluster"].(string)
		actual, _ := item["linked_to_cluster"].(string)
		if actual != want || (item["intent"] != nil && item["intent"] != "present") {
			return false
		}
		found = true
	}
	return found
}

func smbAuthDeleted(request Request, raw []byte) bool {
	return smbAuthPresenceMatches(request, raw, false)
}

func smbAuthPresenceMatches(request Request, raw []byte, present bool) bool {
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
	found := false
	for _, item := range response.Resources {
		id, _ := item[idField].(string)
		if item["resource_type"] != resourceType || !smbResourceIDPattern.MatchString(id) || seen[id] {
			return false
		}
		seen[id] = true
		found = found || id == target
	}
	return found == present
}
