package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
)

func smbShareRecord(data []byte, request Request) (map[string]any, error) {
	cluster, share, err := decodePair(last(resourceTail(request.ResourceKey)))
	if err != nil {
		return nil, err
	}
	var record map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	if decoder.Decode(&record) != nil || record == nil {
		return nil, invalid("invalid SMB share response")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF || record["resource_type"] != "ceph.smb.share" || record["cluster_id"] != cluster || record["share_id"] != share {
		return nil, invalid("SMB share response identity mismatch")
	}
	if _, ok := record["cephfs"].(map[string]any); !ok {
		return nil, invalid("SMB share storage is missing")
	}
	return record, nil
}

func smbShareUpdateJSON(data []byte, request Request) ([]byte, error) {
	record, err := smbShareRecord(data, request)
	if err != nil {
		return nil, err
	}
	fs := record["cephfs"].(map[string]any)
	for _, field := range []string{"readonly", "browseable"} {
		if value, exists := request.Parameters[field]; exists {
			flag, ok := value.(bool)
			if !ok {
				return nil, invalid(field + " must be a boolean")
			}
			record[field] = flag
		}
	}
	fs["volume"] = optional(request.Parameters, "filesystem")
	if value, exists := request.Parameters["path"]; exists {
		path, ok := value.(string)
		if !ok || path == "" {
			return nil, invalid("path must be a non-empty string")
		}
		fs["path"] = path
	}
	return json.Marshal(record)
}

func smbShareUpdateMatches(wanted, actual []byte, request Request) bool {
	expected, err := smbShareRecord(wanted, request)
	if err != nil {
		return false
	}
	found, err := smbShareRecord(actual, request)
	if err != nil {
		return false
	}
	for key, value := range expected {
		if !reflect.DeepEqual(found[key], value) {
			return false
		}
	}
	return true
}
