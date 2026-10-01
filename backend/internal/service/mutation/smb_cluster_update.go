package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
)

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
	for key, value := range expected {
		if !reflect.DeepEqual(found[key], value) {
			return false
		}
	}
	return true
}
