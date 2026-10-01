package mutation

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"encoding/json"
	"io"
	"regexp"
	"strings"
	"time"
)

func snapshotRetentionMatches(p map[string]any, output []byte) bool {
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	decoder.UseNumber()
	var rows []map[string]any
	if decoder.Decode(&rows) != nil || len(rows) == 0 {
		return false
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return false
	}
	rules := regexp.MustCompile(`([1-9][0-9]*)([nmhdwMy])`).FindAllStringSubmatch(optional(p, "retention"), -1)
	if len(rules) == 0 {
		return false
	}
	for _, row := range rows {
		retention, ok := row["retention"].(map[string]any)
		if !ok {
			return false
		}
		for _, rule := range rules {
			value, exists := retention[rule[2]]
			if optional(p, "action") == "remove" {
				if exists {
					return false
				}
			} else {
				count, ok := value.(json.Number)
				if !ok || count.String() != rule[1] {
					return false
				}
			}
		}
	}
	return true
}

func snapshotRetention(request Request, p map[string]any) (command, error) {
	fs := pathValue(resourceTail(request.ResourceKey), "filesystem")
	path, err := required(p, "path")
	if err != nil {
		return command{}, err
	}
	if fs == "" || !strings.HasPrefix(path, "/") {
		return command{}, invalid("filesystem and absolute path are required")
	}
	verb, err := enum(p, "action", "add", "remove")
	if err != nil {
		return command{}, err
	}
	retention, err := required(p, "retention")
	if err != nil {
		return command{}, err
	}
	if !regexp.MustCompile(`^([1-9][0-9]*[nmhdwMy])+$`).MatchString(retention) {
		return command{}, invalid("invalid retention specification")
	}
	seen := map[rune]bool{}
	for _, unit := range retention {
		if unit >= '0' && unit <= '9' {
			continue
		}
		if seen[unit] {
			return command{}, invalid("retention units must not repeat")
		}
		seen[unit] = true
	}
	args := []string{"fs", "snap-schedule", "retention", verb, path, retention, "--fs=" + fs}
	check := []string{"fs", "snap-schedule", "status", path, "--fs=" + fs, "--format=json"}
	for _, option := range []string{"subvol", "group"} {
		value := optional(p, option)
		if strings.ContainsAny(value, "\x00\r\n") {
			return command{}, invalid("invalid subvolume scope")
		}
		if value != "" {
			args = append(args, "--"+option+"="+value)
			check = append(check, "--"+option+"="+value)
		}
	}
	if optional(p, "group") != "" && optional(p, "subvol") == "" {
		return command{}, invalid("subvolume is required with group")
	}
	return command{binary: executor.BinaryCeph, args: args, check: check, timeout: time.Minute}, nil
}
