package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"reflect"
	"regexp"
	"strings"
)

var smbResourceIDPattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,16}[a-zA-Z0-9])?$`)
var smbPlacementHostPattern = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_.-]{0,252}$`)

func smbDNSServers(value any) ([]string, error) {
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
	return servers, nil
}

func smbPublicAddresses(value any) ([]string, error) {
	data, err := json.Marshal(value)
	var addresses []string
	if err != nil || json.Unmarshal(data, &addresses) != nil || addresses == nil {
		return nil, invalid("smb_public_addresses must be an array")
	}
	for _, address := range addresses {
		parts := strings.Split(address, "%")
		if len(parts) > 2 {
			return nil, invalid("public address must be IP/prefix or IP/prefix%network")
		}
		if _, _, err := net.ParseCIDR(parts[0]); err != nil {
			return nil, invalid("public address requires an IP address and prefix length")
		}
		if len(parts) == 2 {
			ip, network, err := net.ParseCIDR(parts[1])
			if err != nil || !ip.Equal(network.IP) {
				return nil, invalid("public address destination must be a network CIDR without host bits")
			}
		}
	}
	return addresses, nil
}

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
	if record["auth_mode"] != mode {
		if mode == "active-directory" {
			_, hasRealm := request.Parameters["domain_realm"]
			_, hasJoin := request.Parameters["domain_join_ref"]
			if !hasRealm || !hasJoin {
				return nil, invalid("switching to active-directory requires domain_realm and domain_join_ref")
			}
		} else if _, exists := request.Parameters["user_group_ref"]; !exists {
			return nil, invalid("switching to user authentication requires user_group_ref")
		}
	}
	record["auth_mode"] = mode
	if _, exists := request.Parameters["clustering"]; exists {
		clustering, err := enum(request.Parameters, "clustering", "default", "always", "never")
		if err != nil {
			return nil, err
		}
		record["clustering"] = clustering
	}
	if value, exists := request.Parameters["smb_label"]; exists {
		label, ok := value.(string)
		if !ok || !smbPlacementHostPattern.MatchString(label) {
			return nil, invalid("smb_label must be a plain non-empty label")
		}
		if _, hosts := request.Parameters["smb_hosts"]; hosts {
			return nil, invalid("smb_label and smb_hosts are mutually exclusive")
		}
		placement, ok := record["placement"].(map[string]any)
		if !ok {
			if record["placement"] != nil {
				return nil, invalid("existing SMB placement is invalid")
			}
			placement = map[string]any{}
		}
		placement["label"] = label
		delete(placement, "hosts")
		delete(placement, "host_pattern")
		record["placement"] = placement
	}
	if value, exists := request.Parameters["smb_hosts"]; exists {
		data, err := json.Marshal(value)
		var hosts []string
		if err != nil || json.Unmarshal(data, &hosts) != nil || len(hosts) == 0 {
			return nil, invalid("smb_hosts must be a non-empty list of hostnames")
		}
		seen := map[string]bool{}
		for _, host := range hosts {
			if !smbPlacementHostPattern.MatchString(host) || strings.Trim(host, "0123456789") == "" || seen[host] {
				return nil, invalid("smb_hosts requires unique plain hostnames")
			}
			seen[host] = true
		}
		placement, ok := record["placement"].(map[string]any)
		if !ok {
			if record["placement"] != nil {
				return nil, invalid("existing SMB placement is invalid")
			}
			placement = map[string]any{}
		}
		placement["hosts"] = hosts
		delete(placement, "label")
		delete(placement, "host_pattern")
		record["placement"] = placement
	}
	if _, exists := request.Parameters["count"]; exists {
		count, err := optionalPositiveInteger(request.Parameters, "count")
		if err != nil || count == "" {
			return nil, invalid("count must be a positive integer")
		}
		placement, ok := record["placement"].(map[string]any)
		if !ok {
			if record["placement"] != nil {
				return nil, invalid("existing SMB placement is invalid")
			}
			placement = map[string]any{}
		}
		placement["count"] = json.Number(count)
		record["placement"] = placement
	}
	_, hasRealm := request.Parameters["domain_realm"]
	_, hasJoin := request.Parameters["domain_join_ref"]
	if hasRealm || hasJoin {
		realm, ok := request.Parameters["domain_realm"].(string)
		if mode != "active-directory" || !hasRealm || !hasJoin || !ok || strings.TrimSpace(realm) == "" || strings.ContainsAny(realm, "\x00\r\n") {
			return nil, invalid("AD configuration requires domain_realm and domain_join_ref in active-directory mode")
		}
		data, err := json.Marshal(request.Parameters["domain_join_ref"])
		var refs []string
		if err != nil || json.Unmarshal(data, &refs) != nil || len(refs) == 0 {
			return nil, invalid("domain_join_ref must be a non-empty list")
		}
		sources := make([]map[string]string, 0, len(refs))
		seen := map[string]bool{}
		for _, ref := range refs {
			if !smbResourceIDPattern.MatchString(ref) || seen[ref] {
				return nil, invalid("domain_join_ref requires unique valid SMB resource IDs")
			}
			seen[ref] = true
			sources = append(sources, map[string]string{"source_type": "resource", "ref": ref})
		}
		record["domain_settings"] = map[string]any{"realm": realm, "join_sources": sources}
		delete(record, "user_group_settings")
	}
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
		servers, err := smbDNSServers(value)
		if err != nil {
			return nil, err
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
	if _, changed := request.Parameters["domain_join_ref"]; changed {
		if sources, ok := found["user_group_settings"].([]any); found["user_group_settings"] != nil && (!ok || len(sources) != 0) {
			return false
		}
	}
	for key, value := range expected {
		if !reflect.DeepEqual(found[key], value) {
			return false
		}
	}
	return true
}
