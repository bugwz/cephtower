package mutation

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"time"
	"unicode"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func symmetricalFlowParameters(p map[string]any) map[string]any {
	result := map[string]any{}
	for k, v := range p {
		result[k] = v
	}
	result["flow_type"] = "symmetrical"
	return result
}

func (s *Service) executeBucketSyncFlowUpdate(ctx context.Context, access executor.ClusterAccess, request Request, spec command) (cephdomain.ActionResult, error) {
	fail := func(code, message string) (cephdomain.ActionResult, error) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: code, Message: message, Retryable: false}
	}
	read := func(stage string) (executor.CommandResult, error) {
		return s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: spec.binary, Args: spec.check, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	}
	before, err := read("pre_check")
	if err != nil {
		return fail("pre_check_failed", "policy could not be read; no change submitted")
	}
	wanted, groups, valid := bucketSyncPolicyDocument(before.Stdout)
	_, expected, expectedValid := bucketSyncPolicyDocument([]byte(`{"groups":[` + syncGroupString(request.Parameters, "expected_group") + `]}`))
	group := groups[syncGroupString(request.Parameters, "group_id")]
	if !valid || !expectedValid || len(expected) != 1 || group == nil || !reflect.DeepEqual(group, expected[syncGroupString(request.Parameters, "group_id")]) {
		return fail("pre_check_failed", "group missing or changed; refresh before editing")
	}
	flows, ok := group["data_flow"].(map[string]any)["symmetrical"].([]any)
	if !ok {
		return fail("pre_check_failed", "symmetrical flows unavailable")
	}
	var flow map[string]any
	for _, raw := range flows {
		entry, ok := raw.(map[string]any)
		if !ok {
			return fail("pre_check_failed", "flow entry invalid")
		}
		if entry["id"] == syncGroupString(request.Parameters, "flow_id") {
			if flow != nil {
				return fail("pre_check_failed", "flow ID ambiguous")
			}
			flow = entry
		}
	}
	if flow == nil {
		return fail("pre_check_failed", "flow missing")
	}
	zonesResult, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + ".zones", Binary: spec.binary, Args: []string{"zonegroup", "get", "--format", "json"}, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	if err != nil {
		return fail("pre_check_failed", "zonegroup could not be read; no change submitted")
	}
	p := symmetricalFlowParameters(request.Parameters)
	resolved, err := resolveBucketSyncFlow(p, zonesResult.Stdout)
	if err != nil {
		return fail("pre_check_failed", err.Error())
	}
	var zoneDocument struct {
		Zones []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"zones"`
	}
	if json.Unmarshal(zonesResult.Stdout, &zoneDocument) != nil {
		return fail("pre_check_failed", "zonegroup mapping invalid")
	}
	byName := map[string]string{}
	byID := map[string]string{}
	for _, zone := range zoneDocument.Zones {
		byName[zone.Name] = zone.ID
		byID[zone.ID] = zone.Name
	}
	oldNames, ok := flow["zones"].([]any)
	if !ok {
		return fail("pre_check_failed", "existing zone list invalid")
	}
	oldIDs := map[string]bool{}
	for _, raw := range oldNames {
		name, ok := raw.(string)
		id := byName[name]
		if !ok || !syncFlowToken(id) || strings.ContainsAny(id, ",;=*") || strings.IndexFunc(id, unicode.IsSpace) >= 0 || oldIDs[id] {
			return fail("pre_check_failed", "existing zone mapping unavailable or ambiguous")
		}
		oldIDs[id] = true
	}
	desired := map[string]bool{}
	added := []string{}
	removed := []string{}
	for _, raw := range p["zones"].([]any) {
		id := raw.(string)
		desired[id] = true
		if !oldIDs[id] {
			added = append(added, id)
		}
	}
	for id := range oldIDs {
		if !desired[id] {
			removed = append(removed, id)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	if len(added) == 0 && len(removed) == 0 {
		return fail("pre_check_failed", "zone membership unchanged")
	}
	setNames := func(ids map[string]bool) []any {
		keys := []string{}
		for id := range ids {
			keys = append(keys, id)
		}
		sort.Strings(keys)
		names := []any{}
		for _, id := range keys {
			names = append(names, byID[id])
		}
		return names
	}
	mutate := func(stage, verb string, ids []string) error {
		args := append([]string(nil), spec.args...)
		args[3] = verb
		for i := range args {
			if args[i] == "--zone-ids" {
				args[i+1] = strings.Join(ids, ",")
				break
			}
		}
		_, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: spec.binary, Args: args, Mutating: true, Timeout: spec.timeout, MaxOutput: executor.DefaultMaxOutput})
		if err != nil {
			return &cephdomain.ActionError{Code: "command_failed", Message: stage + " outcome uncertain; changes may be partial, refresh before retrying", Retryable: false}
		}
		after, err := read(stage + "_post_check")
		actual, _, valid := bucketSyncPolicyDocument(after.Stdout)
		if err != nil || !valid || !reflect.DeepEqual(wanted, actual) {
			return &cephdomain.ActionError{Code: "post_check_failed", Message: stage + " submitted but policy did not match; stop, refresh and inspect partial changes", Retryable: false}
		}
		return nil
	}
	if len(added) > 0 {
		for _, id := range added {
			oldIDs[id] = true
		}
		flow["zones"] = setNames(oldIDs)
		if err := mutate("add", "create", added); err != nil {
			return cephdomain.ActionResult{}, err
		}
	}
	if len(removed) > 0 {
		flow["zones"] = resolved["zones"]
		if err := mutate("remove", "remove", removed); err != nil {
			return cephdomain.ActionResult{}, err
		}
	}
	return cephdomain.ActionResult{}, nil
}
