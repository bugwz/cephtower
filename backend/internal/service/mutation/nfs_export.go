package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"path"
	"strconv"
	"strings"
)

func nfsTransports(value any) (int, error) {
	data, err := json.Marshal(value)
	var transports []string
	if err != nil || json.Unmarshal(data, &transports) != nil || len(transports) == 0 {
		return 0, invalid("transports must contain unique TCP or UDP values")
	}
	mask := 0
	for _, transport := range transports {
		bit := map[string]int{"TCP": 1, "UDP": 2}[transport]
		if bit == 0 || mask&bit != 0 {
			return 0, invalid("transports must contain unique TCP or UDP values")
		}
		mask |= bit
	}
	return mask, nil
}

func nfsProtocols(value any) (int, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return 0, invalid("protocols must contain unique NFS versions 3 or 4")
	}
	var versions []int
	if json.Unmarshal(data, &versions) != nil || len(versions) == 0 {
		return 0, invalid("protocols must contain unique NFS versions 3 or 4")
	}
	mask := 0
	for _, version := range versions {
		if (version != 3 && version != 4) || mask&(1<<version) != 0 {
			return 0, invalid("protocols must contain unique NFS versions 3 or 4")
		}
		mask |= 1 << version
	}
	return mask, nil
}

func nfsExportCreateAvailable(data []byte, cluster, pseudo string) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var exports []map[string]any
	if decoder.Decode(&exports) != nil || exports == nil {
		return false
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return false
	}
	for _, export := range exports {
		existing, ok := export["pseudo"].(string)
		if !ok || !strings.HasPrefix(existing, "/") || path.Clean(existing) == path.Clean(pseudo) {
			return false
		}
		if value, exists := export["cluster_id"]; exists && value != cluster {
			return false
		}
	}
	return true
}

func nfsExportRecord(data []byte, cluster, exportID string) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var exports []map[string]any
	if decoder.Decode(&exports) != nil {
		return nil, invalid("invalid native NFS export list")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return nil, invalid("invalid native NFS export list")
	}
	var found map[string]any
	for _, export := range exports {
		id, ok := export["export_id"].(json.Number)
		if !ok || id.String() != exportID {
			continue
		}
		path, ok := export["pseudo"].(string)
		if !ok || !strings.HasPrefix(path, "/") || strings.ContainsAny(path, "\x00\r\n") || found != nil {
			return nil, invalid("invalid or ambiguous NFS export identity")
		}
		if value, exists := export["cluster_id"]; exists && value != cluster {
			return nil, invalid("NFS export cluster mismatch")
		}
		found = export
	}
	if found == nil {
		return nil, invalid("NFS export no longer exists")
	}
	return found, nil
}

func nfsExportDeleted(request Request, data []byte) bool {
	cluster, id, err := decodePair(last(resourceTail(request.ResourceKey)))
	if err != nil {
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var exports []map[string]any
	if decoder.Decode(&exports) != nil || exports == nil {
		return false
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return false
	}
	for _, export := range exports {
		current, ok := export["export_id"].(json.Number)
		if !ok || current.String() == id {
			return false
		}
		if _, err := strconv.ParseUint(current.String(), 10, 64); err != nil {
			return false
		}
		if value, exists := export["cluster_id"]; exists && value != cluster {
			return false
		}
	}
	return true
}

func nfsExportUpdateJSON(export, parameters map[string]any) ([]byte, error) {
	fsal, ok := export["fsal"].(map[string]any)
	if !ok || fsal["name"] != "CEPH" {
		return nil, invalid("this update form requires a CephFS export")
	}
	if export["access_type"] != "RO" && export["access_type"] != "RW" {
		return nil, invalid("this update form requires RO or RW access")
	}
	export["pseudo"] = parameters["pseudo"]
	export["path"] = parameters["path"]
	fsal["fs_name"] = parameters["filesystem"]
	if transports, exists := parameters["transports"]; exists {
		export["transports"] = transports
	}
	if protocols, exists := parameters["protocols"]; exists {
		export["protocols"] = protocols
	}
	if label, exists := parameters["security_label"]; exists {
		export["security_label"] = label
	}
	if squash, exists := parameters["squash"]; exists {
		export["squash"] = squash
	}
	if readOnly, ok := parameters["read_only"].(bool); ok {
		export["access_type"] = map[bool]string{true: "RO", false: "RW"}[readOnly]
	}
	return json.Marshal(export)
}

func nfsExportUpdateMatches(request Request, data []byte) bool {
	cluster, id, err := decodePair(last(resourceTail(request.ResourceKey)))
	if err != nil {
		return false
	}
	export, err := nfsExportRecord(data, cluster, id)
	if err != nil {
		return false
	}
	return nfsExportAttributesMatch(export, request.Parameters)
}

func nfsExportCreateMatches(p map[string]any, data []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var exports []map[string]any
	if decoder.Decode(&exports) != nil {
		return false
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return false
	}
	found := false
	for _, export := range exports {
		pseudo, ok := export["pseudo"].(string)
		if !ok || path.Clean(pseudo) != path.Clean(optional(p, "pseudo")) {
			continue
		}
		id, ok := export["export_id"].(json.Number)
		if !ok {
			return false
		}
		if value, err := strconv.ParseUint(id.String(), 10, 64); err != nil || value == 0 {
			return false
		}
		if found || export["cluster_id"] != p["cluster"] || !nfsExportAttributesMatch(export, p) {
			return false
		}
		found = true
	}
	return found
}

func nfsExportAttributesMatch(export, p map[string]any) bool {
	if transports, exists := p["transports"]; exists {
		wanted, err := nfsTransports(transports)
		actual, actualErr := nfsTransports(export["transports"])
		if err != nil || actualErr != nil || wanted != actual {
			return false
		}
	}
	if protocols, exists := p["protocols"]; exists {
		wanted, err := nfsProtocols(protocols)
		actual, actualErr := nfsProtocols(export["protocols"])
		if err != nil || actualErr != nil || wanted != actual {
			return false
		}
	}
	if label, exists := p["security_label"]; exists && export["security_label"] != label {
		return false
	}
	if squash, exists := p["squash"]; exists && export["squash"] != squash {
		return false
	}
	fsal, ok := export["fsal"].(map[string]any)
	if !ok || fsal["name"] != "CEPH" || fsal["fs_name"] != p["filesystem"] || path.Clean(optional(export, "pseudo")) != path.Clean(optional(p, "pseudo")) || path.Clean(optional(export, "path")) != path.Clean(optional(p, "path")) {
		return false
	}
	if readOnly, ok := p["read_only"].(bool); ok {
		return export["access_type"] == map[bool]string{true: "RO", false: "RW"}[readOnly]
	}
	return true
}
