package mutation

import (
	"encoding/json"
	"reflect"
	"strings"
)

func zoneGroupUpdateMatches(p map[string]any, output []byte) bool {
	var group struct {
		Name       string           `json:"name"`
		MasterZone string           `json:"master_zone"`
		Zones      []map[string]any `json:"zones"`
	}
	if json.Unmarshal(output, &group) != nil || group.Name != optional(p, "zonegroup") {
		return false
	}
	for _, zone := range group.Zones {
		if zone["name"] != optional(p, "new_name") {
			continue
		}
		id, ok := zone["id"].(string)
		if !ok || id == "" {
			return false
		}
		for _, key := range []string{"read_only", "tier_type", "sync_from_all"} {
			if expected, present := p[key]; present && !reflect.DeepEqual(zone[key], expected) {
				return false
			}
		}
		if p["master"] == true && group.MasterZone != id {
			return false
		}
		if expected, present := p["sync_from"]; present {
			raw, ok := expected.(string)
			if !ok {
				return false
			}
			actual, ok := zone["sync_from"].([]any)
			if !ok {
				return false
			}
			sources := map[string]bool{}
			for _, item := range actual {
				source, ok := item.(string)
				if !ok {
					return false
				}
				sources[source] = true
			}
			removing := p["sync_from_all"] == true
			for _, item := range strings.Split(raw, ",") {
				source := strings.TrimSpace(item)
				if source == "" {
					continue
				}
				if sources[source] == removing {
					return false
				}
			}
		}
		if expected, present := p["endpoints"]; present {
			raw, ok := expected.(string)
			if !ok {
				return false
			}
			actual, ok := zone["endpoints"].([]any)
			if !ok {
				return false
			}
			endpoints := map[string]bool{}
			for _, item := range actual {
				value, ok := item.(string)
				if !ok {
					return false
				}
				endpoints[value] = true
			}
			wanted := map[string]bool{}
			for _, item := range strings.Split(raw, ",") {
				wanted[strings.TrimSpace(item)] = true
			}
			if len(endpoints) != len(wanted) {
				return false
			}
			for endpoint := range wanted {
				if !endpoints[endpoint] {
					return false
				}
			}
		}
		return true
	}
	return false
}

func zoneCredentialReadbackMatches(p map[string]any, output []byte) bool {
	expectedAccess, present := p["access_key"]
	if !present {
		return true
	}
	var zone struct {
		Name string `json:"name"`
		ID   string `json:"id"`
		Key  struct {
			Access string `json:"access_key"`
			Secret string `json:"secret_key"`
		} `json:"system_key"`
	}
	if json.Unmarshal(output, &zone) != nil || zone.ID == "" {
		return false
	}
	name := optional(p, "new_name")
	if name == "" {
		name = optional(p, "name")
	}
	return zone.Name == name && zone.Key.Access == expectedAccess && zone.Key.Secret == p["secret_key"]
}
