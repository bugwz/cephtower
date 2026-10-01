package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"reflect"
	"regexp"
)

var smbResourceIDPattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,16}[a-zA-Z0-9])?$`)

func smbClusterRecord(data []byte, request Request) (map[string]any, error) {
	var record map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if decoder.Decode(&record) != nil || record == nil {
		return nil, invalid("invalid SMB cluster response")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF || record["resource_type"] != "ceph.smb.cluster" || record["cluster_id"] != last(resourceTail(request.ResourceKey)) {
		return nil, invalid("SMB cluster response identity mismatch")
	}
	return record, nil
}

func smbClusterUpdateJSON(data []byte, request Request) ([]byte, error) {
	record, err := smbClusterRecord(data, request)
	if err != nil {
		return nil, err
	}
	mode, err := enum(request.Parameters, "auth_mode", "user", "active-directory")
	if err != nil {
		return nil, err
	}
	record["auth_mode"] = mode
	if value, exists := request.Parameters["user_group_ref"]; exists {
		data, err := json.Marshal(value)
		var refs []string
		if err != nil || json.Unmarshal(data, &refs) != nil || len(refs) == 0 || mode != "user" {
			return nil, invalid("user_group_ref requires a non-empty list in user authentication mode")
		}
		sources := make([]map[string]string, 0, len(refs))
		seen := map[string]bool{}
		for _, ref := range refs {
			if !smbResourceIDPattern.MatchString(ref) || seen[ref] {
				return nil, invalid("user_group_ref requires unique valid SMB resource IDs")
			}
			seen[ref] = true
			sources = append(sources, map[string]string{"source_type": "resource", "ref": ref})
		}
		record["user_group_settings"] = sources
		delete(record, "domain_settings")
	}
	if value, exists := request.Parameters["custom_dns"]; exists {
		data, err := json.Marshal(value)
		var servers []string
		if err != nil || json.Unmarshal(data, &servers) != nil || servers == nil {
			return nil, invalid("custom_dns must be an array of IP addresses")
		}
		for _, server := range servers {
			if net.ParseIP(server) == nil {
				return nil, invalid("custom_dns entries must be IP addresses")
			}
		}
		record["custom_dns"] = servers
	}
	return json.Marshal(record)
}

func smbClusterUpdateMatches(wanted, actual []byte, request Request) bool {
	expected, err := smbClusterRecord(wanted, request)
	if err != nil {
		return false
	}
	found, err := smbClusterRecord(actual, request)
	if err != nil {
		return false
	}
	if _, changed := request.Parameters["user_group_ref"]; changed && found["domain_settings"] != nil {
		return false
	}
	for key, value := range expected {
		if !reflect.DeepEqual(found[key], value) {
			return false
		}
	}
	return true
}
