package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

var groupNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.][A-Za-z0-9_.-]{0,254}$`)

func subvolumeGroupHasAttributes(p map[string]any) bool {
	for _, key := range []string{"pool", "uid", "gid", "mode"} {
		if _, ok := p[key]; ok {
			return true
		}
	}
	return false
}

func subvolumeGroupUpdateCommand(request Request, p map[string]any) (command, error) {
	parts := strings.Split(resourceTail(request.ResourceKey), "/")
	if len(parts) != 4 || parts[0] != "filesystem" || parts[2] != "subvolume-group" || !groupNamePattern.MatchString(parts[1]) || !groupNamePattern.MatchString(parts[3]) {
		return command{}, invalid("invalid subvolume group scope")
	}
	fs, group := parts[1], parts[3]
	check := []string{"fs", "subvolumegroup", "info", fs, group, "--format", "json"}
	wrap := func(args []string) command {
		return command{binary: executor.BinaryCeph, args: args, check: check, timeout: 2 * time.Minute}
	}
	steps := []command{}
	_, sized := p["size"]
	unlimited := boolParameter(p, "unlimited")
	if sized && unlimited {
		return command{}, invalid("size and unlimited cannot be combined")
	}
	if sized || unlimited {
		size := "inf"
		if !unlimited {
			var err error
			size, err = optionalPositiveInteger(p, "size")
			if err != nil || size == "" {
				return command{}, invalid("size must be a positive integer")
			}
		}
		args := []string{"fs", "subvolumegroup", "resize", fs, group, size}
		if boolParameter(p, "no_shrink") {
			args = append(args, "--no_shrink")
		}
		steps = append(steps, wrap(args))
	} else if boolParameter(p, "no_shrink") {
		return command{}, invalid("no_shrink requires a quota update")
	}
	if subvolumeGroupHasAttributes(p) {
		pool, err := required(p, "pool")
		if err != nil {
			return command{}, err
		}
		uid, err := requiredNonNegativeInteger(p, "uid")
		if err != nil {
			return command{}, err
		}
		gid, err := requiredNonNegativeInteger(p, "gid")
		if err != nil {
			return command{}, err
		}
		for _, value := range []string{uid, gid} {
			if _, err := strconv.ParseUint(value, 10, 32); err != nil {
				return command{}, invalid("uid and gid must be 32-bit unsigned integers")
			}
		}
		mode := rawText(p, "mode")
		if !regexp.MustCompile(`^0?[0-7]{3,4}$`).MatchString(mode) {
			return command{}, invalid("mode must be an octal permission")
		}
		// Existing groups are updated by the volumes module's idempotent create.
		// Supply all attributes to prevent omitted fields from resetting ownership or layout.
		steps = append(steps, wrap([]string{"fs", "subvolumegroup", "create", fs, group, "--pool_layout", pool, "--uid", uid, "--gid", gid, "--mode", mode}))
	}
	if len(steps) == 0 {
		return command{}, invalid("at least one quota or attribute update is required")
	}
	result := steps[0]
	result.followups = steps[1:]
	return result, nil
}

func subvolumeGroupInfo(data []byte) (map[string]any, bool) {
	var info map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if decoder.Decode(&info) != nil || info == nil {
		return nil, false
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return nil, false
	}
	return info, true
}

func subvolumeGroupInfoValid(data []byte) bool {
	info, ok := subvolumeGroupInfo(data)
	if !ok {
		return false
	}
	for _, key := range []string{"uid", "gid", "mode"} {
		if _, err := strconv.ParseUint(optional(info, key), 10, 32); err != nil {
			return false
		}
	}
	return optional(info, "data_pool") != "" && optional(info, "bytes_quota") != ""
}

func subvolumeGroupUpdateMatches(p map[string]any, data []byte) bool {
	info, ok := subvolumeGroupInfo(data)
	if !ok {
		return false
	}
	if boolParameter(p, "unlimited") && optional(info, "bytes_quota") != "infinite" {
		return false
	}
	if _, sized := p["size"]; sized && optional(info, "bytes_quota") != optional(p, "size") {
		return false
	}
	if subvolumeGroupHasAttributes(p) {
		if optional(info, "data_pool") != optional(p, "pool") {
			return false
		}
		for _, key := range []string{"uid", "gid"} {
			if optional(info, key) != optional(p, key) {
				return false
			}
		}
		wanted, err := strconv.ParseUint(rawText(p, "mode"), 8, 32)
		actual, parseErr := strconv.ParseUint(optional(info, "mode"), 10, 32)
		if err != nil || parseErr != nil || actual&07777 != wanted {
			return false
		}
	}
	return true
}
