package mutation

import (
	"bytes"
	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"encoding/json"
	"io"
	"reflect"
	"time"
)

func periodDocument(body []byte) map[string]any {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var result map[string]any
	var extra any
	if decoder.Decode(&result) != nil || decoder.Decode(&extra) != io.EOF {
		return nil
	}
	return result
}

func (s *Service) executePeriodCommit(ctx context.Context, access executor.ClusterAccess, request Request, spec command) (cephdomain.ActionResult, error) {
	fail := func(code, message string) (cephdomain.ActionResult, error) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: code, Message: message, Retryable: false}
	}
	realm := syncGroupString(request.Parameters, "realm_id")
	run := func(stage string, args []string, mutating bool) (executor.CommandResult, error) {
		return s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: executor.BinaryRGWAdmin, Args: args, Mutating: mutating, Timeout: 2 * time.Minute, MaxOutput: executor.DefaultMaxOutput})
	}
	realmArgs := []string{"realm", "get", "--realm-id", realm, "--format", "json"}
	before, err := run("pre_check", realmArgs, false)
	snapshot := periodDocument(before.Stdout)
	if err != nil || snapshot == nil || snapshot["id"] != realm || snapshot["current_period"] != syncGroupString(request.Parameters, "expected_current_period") {
		return fail("pre_check_failed", "realm missing or current period changed; refresh before committing")
	}
	written, err := run("commit", spec.args, true)
	if err != nil {
		return fail("command_failed", "period commit outcome uncertain; inspect realm before any manual retry")
	}
	committed := periodDocument(written.Stdout)
	id, ok := committed["id"].(string)
	epoch, epochOK := committed["epoch"].(json.Number)
	epochValue, epochErr := epoch.Int64()
	if !ok || !syncFlowToken(id) || committed["realm_id"] != realm || !epochOK || epochErr != nil || epochValue <= 0 {
		return fail("post_check_failed", "commit returned an unverifiable period; inspect realm before retrying")
	}
	after, err := run("realm_post_check", realmArgs, false)
	actualRealm := periodDocument(after.Stdout)
	if err != nil || actualRealm == nil || actualRealm["id"] != realm || actualRealm["current_period"] != id {
		return fail("post_check_failed", "committed period is not verified as realm current period")
	}
	current, err := run("period_post_check", spec.check, false)
	if err != nil || !reflect.DeepEqual(periodDocument(current.Stdout), committed) {
		return fail("post_check_failed", "current period differs from commit result; inspect concurrent changes before retrying")
	}
	return cephdomain.ActionResult{}, nil
}
