package mutation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"math"
	"math/big"
	"regexp"
	"sort"
	"strconv"
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
	return configurationMatchesType(resource, data, expected, "")
}

func configurationMatchesType(resource string, data []byte, expected *string, kind string) bool {
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
			matched = expected != nil && configurationEquivalent(*expected, *row.Value, kind)
		}
	}
	return matched
}

func configurationEquivalent(expected, actual, kind string) bool {
	if expected == actual {
		return true
	}
	// Option's double printer uses std::fixed with the default precision (6).
	if kind == "float" {
		value, err := strconv.ParseFloat(strings.TrimLeft(expected, "\t\n\r "), 64)
		return err == nil && !math.IsNaN(value) && !math.IsInf(value, 0) && strconv.FormatFloat(value, 'f', 6, 64) == actual
	}
	normalize := func(value string) (string, bool) {
		if kind == "bool" {
			if strings.EqualFold(value, "true") {
				return "true", true
			}
			if strings.EqualFold(value, "false") {
				return "false", true
			}
			if !regexp.MustCompile(`^[\t\n\r ]*[+-]?[0-9]+$`).MatchString(value) {
				return "", false
			}
			n, ok := new(big.Int).SetString(strings.TrimLeft(value, "\t\n\r "), 10)
			if !ok || !n.IsInt64() || n.Int64() < -2147483648 || n.Int64() > 2147483647 {
				return "", false
			}
			if n.Sign() == 0 {
				return "false", true
			}
			return "true", true
		}
		pattern, base := `^[\t\n\r ]*([+-]?[0-9]+)([KMGTPE]?)$`, int64(1000)
		if kind == "size" {
			pattern, base = `^([+]?[0-9]+)([KMGTPE](?:i?B|i)?|B)?$`, 1024
		}
		if kind != "int" && kind != "uint" && kind != "size" {
			return "", false
		}
		parts := regexp.MustCompile(pattern).FindStringSubmatch(value)
		if parts == nil {
			return "", false
		}
		n, ok := new(big.Int).SetString(parts[1], 10)
		if !ok {
			return "", false
		}
		if parts[2] != "" && parts[2] != "B" {
			power := strings.IndexByte("KMGTPE", parts[2][0]) + 1
			n.Mul(n, new(big.Int).Exp(big.NewInt(base), big.NewInt(int64(power)), nil))
		}
		if kind == "int" && !n.IsInt64() || kind != "int" && !n.IsUint64() {
			return "", false
		}
		return n.String(), true
	}
	a, okA := normalize(expected)
	b, okB := normalize(actual)
	return okA && okB && a == b
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
