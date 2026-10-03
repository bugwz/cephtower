package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	cephdomain "cephtower/backend/internal/domain/ceph"
	cephprovider "cephtower/backend/internal/integration/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func topicDeleteIdentity(p map[string]any) (key, scope, name string, err error) {
	id := syncGroupString(p, "topic_id")
	decoded, decodeErr := base64.RawURLEncoding.Strict().DecodeString(id)
	key = string(decoded)
	scope, name, found := strings.Cut(key, ":")
	if decodeErr != nil || !utf8.Valid(decoded) || base64.RawURLEncoding.EncodeToString(decoded) != id || !found || name == "" || strings.TrimSpace(key) != key || strings.IndexFunc(key, unicode.IsControl) >= 0 || len(key) > 1024 || strings.HasPrefix(name, "-") || strings.HasPrefix(scope, "-") {
		return "", "", "", invalid("valid scoped topic_id required")
	}
	if _, valid := cephprovider.RGWTopicMetadataVersion(periodDocument([]byte(syncGroupString(p, "expected_version")))); !valid {
		return "", "", "", invalid("valid native topic metadata version required")
	}
	return key, scope, name, nil
}

func topicDeleteCommand(p map[string]any, rgw func([]string, []string) command) (command, error) {
	key, scope, name, err := topicDeleteIdentity(p)
	if err != nil {
		return command{}, err
	}
	return rgw([]string{"topic", "rm", "--tenant", scope, "--topic", name}, []string{"metadata", "get", "topic:" + key}), nil
}

func (s *Service) executeTopicDelete(ctx context.Context, access executor.ClusterAccess, request Request, spec command) (cephdomain.ActionResult, error) {
	fail := func(code, message string) (cephdomain.ActionResult, error) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: code, Message: message, Retryable: false}
	}
	key, scope, name, _ := topicDeleteIdentity(request.Parameters)
	run := func(stage string, args []string, write bool) (executor.CommandResult, error) {
		return s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: spec.binary, Args: args, Mutating: write, Timeout: spec.timeout, MaxOutput: executor.DefaultMaxOutput})
	}
	before, err := run("pre_check", spec.check, false)
	document := periodDocument(before.Stdout)
	version, valid := cephprovider.RGWTopicMetadataVersion(document["ver"])
	expected, _ := cephprovider.RGWTopicMetadataVersion(periodDocument([]byte(syncGroupString(request.Parameters, "expected_version"))))
	data, ok := document["data"].(map[string]any)
	if err != nil || !valid || version != expected || document["key"] != "topic:"+key || !ok || data["name"] != name {
		return fail("pre_check_failed", "topic identity or metadata version changed; no deletion submitted")
	}
	// Require the v2 topic read to resolve the exact same configuration. The
	// subscribed_buckets field distinguishes this path from legacy topic storage.
	current, err := run("topic_check", []string{"topic", "get", "--tenant", scope, "--topic", name, "--format", "json"}, false)
	topic := periodDocument(current.Stdout)
	bindings, ok := topic["subscribed_buckets"].([]any)
	if err != nil || !ok {
		return fail("pre_check_failed", "native v2 topic scope cannot be verified; no deletion submitted")
	}
	for _, binding := range bindings {
		if _, ok := binding.(string); !ok {
			return fail("pre_check_failed", "invalid topic binding data; no deletion submitted")
		}
	}
	delete(topic, "subscribed_buckets")
	if !reflect.DeepEqual(data, topic) {
		return fail("pre_check_failed", "topic configuration changed or scope differs; no deletion submitted")
	}
	if _, err := run("delete", spec.args, true); err != nil {
		return fail("command_failed", "topic deletion failed or partially applied; inspect the topic and persistent queue before retrying")
	}
	after, err := run("post_check", []string{"metadata", "list", "topic", "--format", "json"}, false)
	var keys []string
	if err != nil || json.Unmarshal(after.Stdout, &keys) != nil || keys == nil {
		return fail("post_check_failed", "deletion submitted but topic absence could not be verified")
	}
	seen := map[string]bool{}
	for _, item := range keys {
		_, itemName, scoped := strings.Cut(item, ":")
		if !scoped || itemName == "" || strings.TrimSpace(item) != item || strings.IndexFunc(item, unicode.IsControl) >= 0 || seen[item] || item == key {
			return fail("post_check_failed", "deletion submitted but topic absence could not be verified")
		}
		seen[item] = true
	}
	return cephdomain.ActionResult{}, nil
}
