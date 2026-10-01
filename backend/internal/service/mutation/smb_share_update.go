package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"path"
	"reflect"
	"regexp"
	"strings"
)

var smbShareNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9. _-]{0,63}$`)

func smbSharePath(value string) (string, error) {
	if value == "" || strings.ContainsAny(value, "\x00\r\n") {
		return "", invalid("path must be a non-empty single-line string")
	}
	cleaned := path.Clean(value)
	// Python posixpath.normpath preserves exactly two leading slashes.
	if strings.HasPrefix(value, "//") && !strings.HasPrefix(value, "///") {
		cleaned = "/" + cleaned
	}
	if cleaned != "/" {
		for _, part := range strings.Split(strings.TrimLeft(cleaned, "/"), "/") {
			if part == "" || part == "." || part == ".." {
				return "", invalid("path must resolve within the selected SMB storage scope")
			}
		}
	}
	return cleaned, nil
}

func smbShareName(parameters map[string]any) (string, bool, error) {
	value, exists := parameters["share_name"]
	if !exists {
		return "", false, nil
	}
	name, ok := value.(string)
	if !ok || !smbShareNamePattern.MatchString(name) {
		return "", true, invalid("share_name must contain 1 to 64 ASCII letters, digits, spaces, dots, underscores or hyphens and start with a letter, digit or underscore")
	}
	return name, true, nil
}

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
	if err := applySMBLoginControl(record, request.Parameters); err != nil {
		return nil, err
	}
	if value, exists := request.Parameters["comment"]; exists {
		comment, ok := value.(string)
		if !ok || strings.ContainsAny(comment, "\x00\r\n") {
			return nil, invalid("comment must be a single-line string")
		}
		record["comment"] = comment
	}
	if name, exists, err := smbShareName(request.Parameters); err != nil {
		return nil, err
	} else if exists {
		record["name"] = name
	}
	for _, field := range []string{"readonly", "browseable"} {
		if value, exists := request.Parameters[field]; exists {
			flag, ok := value.(bool)
			if !ok {
				return nil, invalid(field + " must be a boolean")
			}
			record[field] = flag
		}
	}
	volume := optional(request.Parameters, "filesystem")
	if value, changed := request.Parameters["subvolume"]; changed {
		subvolume, ok := value.(string)
		if !ok || strings.ContainsAny(subvolume, "\x00\r\n") {
			return nil, invalid("subvolume must be a name, group/name, or empty to use filesystem scope")
		}
		if optional(request.Parameters, "path") == "" {
			return nil, invalid("path is required when replacing the SMB storage scope")
		}
		delete(fs, "subvolume")
		delete(fs, "subvolumegroup")
		if subvolume != "" {
			parts := strings.Split(subvolume, "/")
			if len(parts) > 2 {
				return nil, invalid("subvolume must be a name or group/name")
			}
			for _, part := range parts {
				if strings.TrimSpace(part) == "" || part == "." || part == ".." {
					return nil, invalid("subvolume must be a name or group/name")
				}
			}
			fs["subvolume"] = parts[len(parts)-1]
			if len(parts) == 2 {
				fs["subvolumegroup"] = parts[0]
			}
		}
	} else if fs["volume"] != volume {
		for _, field := range []string{"subvolume", "subvolumegroup"} {
			if value := fs[field]; value != nil && value != "" {
				return nil, invalid("cannot change the filesystem while preserving an existing SMB subvolume scope")
			}
		}
	}
	fs["volume"] = volume
	if value, exists := request.Parameters["path"]; exists {
		path, ok := value.(string)
		if !ok || path == "" {
			return nil, invalid("path must be a non-empty string")
		}
		cleaned, err := smbSharePath(path)
		if err != nil {
			return nil, err
		}
		fs["path"] = cleaned
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
	if _, changed := request.Parameters["subvolume"]; changed {
		for _, field := range []string{"subvolume", "subvolumegroup"} {
			wantedScope := expected["cephfs"].(map[string]any)[field]
			actualScope := found["cephfs"].(map[string]any)[field]
			if wantedScope == nil {
				wantedScope = ""
			}
			if actualScope == nil {
				actualScope = ""
			}
			if !reflect.DeepEqual(wantedScope, actualScope) {
				return false
			}
			// Empty optional scope fields can be omitted by native serialization.
			if wantedScope == "" {
				delete(found["cephfs"].(map[string]any), field)
			}
		}
	}
	for key, value := range expected {
		// Native serialization omits the quiet false restrict_access field.
		if key == "restrict_access" && value == false {
			if _, exists := found[key]; !exists {
				continue
			}
		}
		if !reflect.DeepEqual(found[key], value) {
			return false
		}
	}
	return true
}
