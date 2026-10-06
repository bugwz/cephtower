package mutation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
)

func configurationDeleted(resource string, data []byte) bool {
	return configurationMatches(resource, data, nil)
}

func configurationSet(resource string, data []byte, parameters map[string]any) bool {
	value, ok := parameters["value"].(string)
	return ok && configurationMatches(resource, data, &value)
}

func configurationMatches(resource string, data []byte, expected *string) bool {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(strings.TrimPrefix(resource, "configuration/value/"))
	parts := strings.Split(string(decoded), "\x00")
	if err != nil || len(parts) != 2 || !configurationScope.MatchString(parts[0]) || !configurationName.MatchString(parts[1]) {
		return false
	}
	var rows []struct {
		Section       string  `json:"section"`
		Name          string  `json:"name"`
		Value         *string `json:"value"`
		Mask          string  `json:"mask"`
		LocationType  string  `json:"location_type"`
		LocationValue string  `json:"location_value"`
		DeviceClass   string  `json:"device_class"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if decoder.Decode(&rows) != nil || decoder.Decode(new(any)) != io.EOF || rows == nil {
		return false
	}
	seen := map[string]bool{}
	matched := expected == nil
	for _, row := range rows {
		if row.Section == "" || row.Name == "" || row.Value == nil {
			return false
		}
		mask := row.Mask
		if row.LocationType != "" || row.LocationValue != "" {
			if row.LocationType == "" || row.LocationValue == "" {
				return false
			}
			location := row.LocationType + ":" + row.LocationValue
			if row.DeviceClass != "" {
				location += "/class:" + row.DeviceClass
			}
			if mask != "" && configurationScopeKey("osd/"+mask) != configurationScopeKey("osd/"+location) {
				return false
			}
			mask = location
		} else if row.DeviceClass != "" {
			if mask != "" && mask != "class:"+row.DeviceClass {
				return false
			}
			mask = "class:" + row.DeviceClass
		}
		who := row.Section
		if mask != "" {
			who += "/" + mask
		}
		who = configurationScopeKey(who)
		if who == "" {
			return false
		}
		key := who + "\x00" + row.Name
		if seen[key] {
			return false
		}
		seen[key] = true
		if who == configurationScopeKey(parts[0]) && row.Name == parts[1] {
			matched = expected != nil && *row.Value == *expected
		}
	}
	return matched
}

func configurationScopeKey(who string) string {
	if !configurationScope.MatchString(who) {
		return ""
	}
	parts := strings.Split(who, "/")
	sort.Strings(parts[1:])
	return strings.Join(parts, "/")
}

var configurationScope = regexp.MustCompile(`^(global|mon|mgr|osd|mds|client)(\.[A-Za-z0-9_.-]+)?(/[A-Za-z0-9_.:-]+)*$`)
var configurationName = regexp.MustCompile(`^(mgr/[A-Za-z][A-Za-z0-9_]*/([A-Za-z0-9_][A-Za-z0-9_.-]{0,255}/)?)?[A-Za-z][A-Za-z0-9_]{0,255}$`)

func configurationCommand(request Request, p map[string]any) (command, error) {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(strings.TrimPrefix(request.ResourceKey, "configuration/value/"))
	parts := strings.Split(string(decoded), "\x00")
	if err != nil || len(parts) != 2 {
		return command{}, invalid("configuration scope and name are required")
	}
	who, name := parts[0], parts[1]
	if !configurationScope.MatchString(who) || !configurationName.MatchString(name) {
		return command{}, invalid("invalid configuration scope or name")
	}
	spec := command{binary: executor.BinaryCeph, timeout: 30 * time.Second, check: []string{"config", "dump", "--format", "json"}}
	if request.Action == "config_value.delete" {
		spec.args = []string{"config", "rm", who, name}
		return spec, nil
	}
	value, ok := p["value"].(string)
	if !ok || len(value) > 32<<10 || strings.ContainsRune(value, 0) {
		return command{}, invalid("configuration value must be text of at most 32 KiB")
	}
	// Named syntax preserves empty values, whitespace and values beginning with '-'.
	spec.args = []string{"config", "set", who, name}
	if value == "" || strings.ContainsAny(value, "\r\n") {
		// The equals parser does not match empty values or span newlines. Stop
		// global option parsing before passing the complete value separately.
		spec.args = []string{"--", "config", "set", who, name, "--value", value}
	} else {
		spec.args = append(spec.args, "--value="+value)
	}
	if security.IsSensitiveName(name) {
		spec.sensitive = map[int]struct{}{len(spec.args) - 1: {}}
	}
	return spec, nil
}
