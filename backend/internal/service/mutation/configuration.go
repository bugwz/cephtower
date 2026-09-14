package mutation

import (
	"encoding/base64"
	"regexp"
	"strings"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
)

var configurationScope = regexp.MustCompile(`^(global|mon|mgr|osd|mds|client)(\.[A-Za-z0-9_.-]+)?(/[A-Za-z0-9_.:-]+)*$`)
var configurationName = regexp.MustCompile(`^(mgr/[A-Za-z][A-Za-z0-9_]*/)?[A-Za-z][A-Za-z0-9_]{0,255}$`)

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
