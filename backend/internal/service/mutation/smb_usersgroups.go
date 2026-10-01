package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"
)

func smbUsersGroupsJSON(p map[string]any) ([]byte, error) {
	name, _ := p["name"].(string)
	if !smbResourceIDPattern.MatchString(name) {
		return nil, invalid("invalid SMB user group resource ID")
	}
	encoded, err := json.Marshal(p["users"])
	var users []struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err != nil || json.Unmarshal(encoded, &users) != nil || users == nil {
		return nil, invalid("SMB users must be an array")
	}
	seen := map[string]bool{}
	for _, user := range users {
		if strings.TrimSpace(user.Name) == "" || user.Password == "" || seen[user.Name] {
			return nil, invalid("SMB users require unique names and nonempty passwords")
		}
		seen[user.Name] = true
	}
	encoded, err = json.Marshal(p["groups"])
	var names []string
	if err != nil || json.Unmarshal(encoded, &names) != nil || names == nil {
		return nil, invalid("SMB groups must be an array")
	}
	groups := make([]map[string]string, 0, len(names))
	seen = map[string]bool{}
	for _, name := range names {
		if strings.TrimSpace(name) == "" || seen[name] {
			return nil, invalid("SMB group names must be nonempty and unique")
		}
		seen[name] = true
		groups = append(groups, map[string]string{"name": name})
	}
	payload := map[string]any{"resource_type": "ceph.smb.usersgroups", "users_groups_id": name, "intent": "present", "values": map[string]any{"users": users, "groups": groups}}
	if linked, exists := p["linked_to_cluster"]; exists {
		id, ok := linked.(string)
		if !ok || !smbResourceIDPattern.MatchString(id) {
			return nil, invalid("invalid linked SMB cluster ID")
		}
		payload["linked_to_cluster"] = id
	}
	return json.Marshal(payload)
}

func smbUsersGroupsCreated(request Request, raw []byte) bool {
	var response struct {
		Resources []map[string]any `json:"resources"`
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	if d.Decode(&response) != nil || response.Resources == nil || d.Decode(new(any)) != io.EOF {
		return false
	}
	expectedJSON, err := smbUsersGroupsJSON(request.Parameters)
	if err != nil {
		return false
	}
	var expected map[string]any
	if json.Unmarshal(expectedJSON, &expected) != nil {
		return false
	}
	stripPasswords := func(item map[string]any) bool {
		values, ok := item["values"].(map[string]any)
		if !ok {
			return false
		}
		users, ok := values["users"].([]any)
		if !ok {
			return false
		}
		for _, entry := range users {
			user, ok := entry.(map[string]any)
			if !ok {
				return false
			}
			delete(user, "password")
		}
		return true
	}
	if !stripPasswords(expected) {
		return false
	}
	seen := map[string]bool{}
	found := false
	for _, item := range response.Resources {
		id, _ := item["users_groups_id"].(string)
		if item["resource_type"] != "ceph.smb.usersgroups" || !smbResourceIDPattern.MatchString(id) || seen[id] {
			return false
		}
		seen[id] = true
		if id != request.Parameters["name"] {
			continue
		}
		if !stripPasswords(item) || !reflect.DeepEqual(item["values"], expected["values"]) || item["linked_to_cluster"] != expected["linked_to_cluster"] || (item["intent"] != nil && item["intent"] != "present") {
			return false
		}
		found = true
	}
	return found
}
