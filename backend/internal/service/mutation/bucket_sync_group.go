package mutation

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func syncGroupString(p map[string]any, key string) string {
	value, _ := p[key].(string)
	return value
}

func bucketSyncGroupCommand(action string, p map[string]any, rgw func([]string, []string) command) (command, error) {
	encoded := syncGroupString(p, "bucket_id")
	raw, err := base64.RawURLEncoding.Strict().DecodeString(encoded)
	pair := strings.Split(string(raw), "\x00")
	if err != nil || !utf8.Valid(raw) || base64.RawURLEncoding.EncodeToString(raw) != encoded || len(pair) != 2 || pair[1] == "" {
		return command{}, invalid("invalid bucket_id")
	}
	for _, part := range pair {
		if strings.HasPrefix(part, "-") || strings.ContainsAny(part, "/:\\") || strings.IndexFunc(part, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0 {
			return command{}, invalid("invalid bucket identity")
		}
	}
	group := syncGroupString(p, "group_id")
	if group == "" || len(group) > 512 || !utf8.ValidString(group) || strings.HasPrefix(group, "-") || strings.IndexFunc(group, unicode.IsControl) >= 0 {
		return command{}, invalid("invalid group_id")
	}
	status := syncGroupString(p, "status")
	if action == "rgw_bucket.sync_pipe_delete" || action == "rgw_bucket.sync_pipe_create" || action == "rgw_bucket.sync_pipe_update" || action == "rgw_bucket.sync_pipe_zones" {
		var args []string
		var err error
		if action == "rgw_bucket.sync_pipe_zones" {
			args, err = bucketSyncPipeZonesArgs(p)
		} else if action == "rgw_bucket.sync_pipe_update" {
			args, err = bucketSyncPipeUpdateArgs(p)
		} else if action == "rgw_bucket.sync_pipe_create" {
			args, err = bucketSyncPipeCreateArgs(p)
		} else {
			args, err = bucketSyncPipeDeleteArgs(p)
		}
		if err != nil {
			return command{}, err
		}
		target := []string{"--bucket", pair[1], "--tenant", pair[0]}
		return rgw(append(args, target...), append([]string{"sync", "policy", "get"}, target...)), nil
	}
	if action == "rgw_bucket.sync_flow_create" || action == "rgw_bucket.sync_flow_delete" || action == "rgw_bucket.sync_flow_update" {
		var args []string
		var err error
		if action == "rgw_bucket.sync_flow_delete" {
			args, err = bucketSyncFlowDeleteArgs(p)
		} else if action == "rgw_bucket.sync_flow_update" {
			args, err = bucketSyncFlowArgs(symmetricalFlowParameters(p))
		} else {
			args, err = bucketSyncFlowArgs(p)
		}
		if err != nil {
			return command{}, err
		}
		target := []string{"--bucket", pair[1], "--tenant", pair[0]}
		return rgw(append(args, target...), append([]string{"sync", "policy", "get"}, target...)), nil
	}
	if action != "rgw_bucket.sync_group_delete" && status != "enabled" && status != "allowed" && status != "forbidden" {
		return command{}, invalid("invalid sync group status")
	}
	if action == "rgw_bucket.sync_group" && syncGroupString(p, "expected_status") == "" {
		return command{}, invalid("expected_status is required")
	}
	target := []string{"--bucket", pair[1], "--tenant", pair[0]}
	verb := "modify"
	if action == "rgw_bucket.sync_group_create" {
		verb = "create"
	}
	args := append([]string{"sync", "group", verb, "--group-id", group, "--status", status}, target...)
	if action == "rgw_bucket.sync_group_delete" {
		if syncGroupString(p, "expected_group") == "" {
			return command{}, invalid("expected_group is required")
		}
		args = append([]string{"sync", "group", "remove", "--group-id", group}, target...)
	}
	return rgw(args, append([]string{"sync", "policy", "get"}, target...)), nil
}

func bucketSyncPolicyDocument(body []byte) (map[string]any, map[string]map[string]any, bool) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var policy map[string]any
	if decoder.Decode(&policy) != nil {
		return nil, nil, false
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return nil, nil, false
	}
	groups, ok := policy["groups"].([]any)
	if !ok {
		return nil, nil, false
	}
	indexed := map[string]map[string]any{}
	for _, value := range groups {
		group, ok := value.(map[string]any)
		if !ok {
			return nil, nil, false
		}
		id, idOK := group["id"].(string)
		status, statusOK := group["status"].(string)
		_, flowOK := group["data_flow"].(map[string]any)
		_, pipesOK := group["pipes"].([]any)
		if !idOK || !statusOK || status == "" || !flowOK || !pipesOK || indexed[id] != nil {
			return nil, nil, false
		}
		indexed[id] = group
	}
	// Group order is not semantic: native maps can insert new IDs before old IDs.
	policy["groups"] = indexed
	return policy, indexed, true
}

func (s *Service) executeBucketSyncGroup(ctx context.Context, access executor.ClusterAccess, request Request, spec command) (cephdomain.ActionResult, error) {
	fail := func(code, message string) (cephdomain.ActionResult, error) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: code, Message: message, Retryable: false}
	}
	read := func(stage string) (executor.CommandResult, error) {
		return s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: spec.binary, Args: spec.check, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	}
	before, err := read("pre_check")
	if err != nil {
		return fail("pre_check_failed", "bucket sync policy could not be read; no change submitted")
	}
	wanted, groups, ok := bucketSyncPolicyDocument(before.Stdout)
	if !ok {
		return fail("pre_check_failed", "bucket sync policy is invalid; no change submitted")
	}
	id := syncGroupString(request.Parameters, "group_id")
	group := groups[id]
	if request.Action == "rgw_bucket.sync_group_create" {
		if group != nil {
			return fail("pre_check_failed", "sync group already exists; creation must not overwrite it")
		}
		group = map[string]any{"id": id, "data_flow": map[string]any{}, "pipes": []any{}}
		groups[id] = group
	} else if request.Action == "rgw_bucket.sync_group_delete" || request.Action == "rgw_bucket.sync_flow_create" || request.Action == "rgw_bucket.sync_flow_delete" || request.Action == "rgw_bucket.sync_pipe_delete" || request.Action == "rgw_bucket.sync_pipe_create" || request.Action == "rgw_bucket.sync_pipe_update" {
		_, expected, valid := bucketSyncPolicyDocument([]byte(`{"groups":[` + syncGroupString(request.Parameters, "expected_group") + `]}`))
		if !valid || len(expected) != 1 || group == nil || !reflect.DeepEqual(group, expected[id]) {
			return fail("pre_check_failed", "sync group missing or changed; refresh before changing it")
		}
		if request.Action == "rgw_bucket.sync_pipe_update" {
			if err := updateBucketSyncPipe(group, request.Parameters); err != nil {
				return fail("pre_check_failed", err.Error())
			}
		} else if request.Action == "rgw_bucket.sync_group_delete" {
			delete(groups, id)
		} else if request.Action == "rgw_bucket.sync_pipe_delete" {
			if err := removeBucketSyncPipe(group, request.Parameters); err != nil {
				return fail("pre_check_failed", err.Error())
			}
		} else if request.Action == "rgw_bucket.sync_flow_delete" && syncGroupString(request.Parameters, "flow_type") == "symmetrical" {
			if err := removeBucketSyncFlow(group, request.Parameters); err != nil {
				return fail("pre_check_failed", err.Error())
			}
		} else {
			zones, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + ".zones", Binary: spec.binary, Args: []string{"zonegroup", "get", "--format", "json"}, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
			if err != nil {
				return fail("pre_check_failed", "zonegroup could not be read; no change submitted")
			}
			var changeErr error
			if request.Action == "rgw_bucket.sync_pipe_create" {
				changeErr = addBucketSyncPipe(group, request.Parameters, zones.Stdout)
			} else {
				resolved, err := resolveBucketSyncFlow(request.Parameters, zones.Stdout)
				if err != nil {
					return fail("pre_check_failed", err.Error())
				}
				if request.Action == "rgw_bucket.sync_flow_delete" {
					changeErr = removeBucketSyncFlow(group, resolved)
				} else {
					changeErr = addBucketSyncFlow(group, resolved)
				}
			}
			if changeErr != nil {
				return fail("pre_check_failed", changeErr.Error())
			}
		}
	} else if group == nil || group["status"] != syncGroupString(request.Parameters, "expected_status") {
		return fail("pre_check_failed", "sync group missing or status changed; refresh before retrying")
	}
	if request.Action == "rgw_bucket.sync_group" || request.Action == "rgw_bucket.sync_group_create" {
		group["status"] = syncGroupString(request.Parameters, "status")
	}
	_, err = s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action, Binary: spec.binary, Args: spec.args, Mutating: true, Timeout: spec.timeout, MaxOutput: executor.DefaultMaxOutput})
	if err != nil {
		return fail("command_failed", "sync group write outcome is uncertain; refresh before retrying")
	}
	after, err := read("post_check")
	actual, _, valid := bucketSyncPolicyDocument(after.Stdout)
	if err != nil || !valid || !reflect.DeepEqual(wanted, actual) {
		return fail("post_check_failed", "sync group change submitted but complete policy did not match; refresh before another change")
	}
	return cephdomain.ActionResult{}, nil
}
