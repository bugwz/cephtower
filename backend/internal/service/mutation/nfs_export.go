package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

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
	if readOnly, ok := parameters["read_only"].(bool); ok {
		export["access_type"] = map[bool]string{true: "RO", false: "RW"}[readOnly]
	}
	return json.Marshal(export)
}
