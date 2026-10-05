package mutation

import (
	"context"
	"reflect"
	"sort"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func buildRealmDelete(p map[string]any) (command, error) {
	for _, field := range []string{"realm_id", "name", "expected_current_period"} {
		if !syncFlowToken(syncGroupString(p, field)) {
			return command{}, invalid("realm_id, name and expected_current_period are required")
		}
	}
	if p["confirm_delete"] != true {
		return command{}, invalid("confirm_delete must be true")
	}
	return command{binary: executor.BinaryRGWAdmin, args: []string{"realm", "rm", "--realm-id", syncGroupString(p, "realm_id")}, timeout: time.Minute}, nil
}

func (s *Service) executeRealmDelete(ctx context.Context, access executor.ClusterAccess, req Request, spec command) (cephdomain.ActionResult, error) {
	id, name := syncGroupString(req.Parameters, "realm_id"), syncGroupString(req.Parameters, "name")
	fail := func(code, message string) (cephdomain.ActionResult, error) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: code, Message: message, Retryable: false}
	}
	run := func(stage string, args []string, write bool) (map[string]any, int, bool) {
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: req.Action + "." + stage, Binary: spec.binary, Args: args, Timeout: spec.timeout, MaxOutput: 1 << 20, Mutating: write})
		defer func() { clear(result.Stdout); clear(result.Stderr) }()
		return periodDocument(result.Stdout), result.ExitCode, err == nil && result.ExitCode == 0 && len(result.Stderr) == 0
	}
	list := func(stage string) ([]string, string, bool) {
		doc, _, ok := run(stage, []string{"realm", "list", "--format", "json"}, false)
		names, valid := realmSetupStrings(doc["realms"])
		def, defaultValid := doc["default_info"].(string)
		sort.Strings(names)
		return names, def, ok && valid && defaultValid
	}
	before, defaultID, ok := list("list_before")
	if !ok || defaultID == id {
		return fail("pre_check_failed", "Realm list or default identity unavailable; switch the default Realm before deleting it")
	}
	expected := make([]string, 0, len(before))
	found := false
	for _, item := range before {
		if item == name {
			found = true
		} else {
			expected = append(expected, item)
		}
	}
	if !found {
		return fail("pre_check_failed", "Realm name is absent; no deletion submitted")
	}
	doc, _, ok := run("identity", []string{"realm", "get", "--realm-id", id, "--format", "json"}, false)
	if !ok || doc["id"] != id || doc["name"] != name || doc["current_period"] != req.Parameters["expected_current_period"] {
		return fail("pre_check_failed", "Realm identity or current Period changed; no deletion submitted")
	}
	if _, _, ok := run("delete", spec.args, true); !ok {
		return fail("command_failed", "Realm deletion failed or partially applied; inspect configuration before any further action")
	}
	_, code, _ := run("absence", []string{"realm", "get", "--realm-id", id, "--format", "json"}, false)
	if code != 2 {
		return fail("post_check_failed", "Deletion submitted but native Realm absence was not verified")
	}
	after, afterDefault, ok := list("list_after")
	if !ok || afterDefault != defaultID || !reflect.DeepEqual(after, expected) {
		return fail("post_check_failed", "Deletion submitted but Realm name index or default reference changed unexpectedly")
	}
	return cephdomain.ActionResult{Details: map[string]any{"realm_id": id, "realm_absence_verified": true, "related_resources_deleted": false}}, nil
}
