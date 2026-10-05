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
	run := func(stage string, args []string, mutating bool) (map[string]any, bool) {
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: executor.BinaryRGWAdmin, Args: args, Mutating: mutating, Timeout: 2 * time.Minute, MaxOutput: executor.DefaultMaxOutput})
		defer func() { clear(result.Stdout); clear(result.Stderr) }()
		if err != nil || result.ExitCode != 0 {
			return nil, false
		}
		// Successful native commands may emit informational diagnostics on stderr.
		return periodDocument(result.Stdout), true
	}
	realmArgs := []string{"realm", "get", "--realm-id", realm, "--format", "json"}
	snapshot, success := run("pre_check", realmArgs, false)
	if !success || snapshot == nil || snapshot["id"] != realm || snapshot["current_period"] != syncGroupString(request.Parameters, "expected_current_period") {
		return fail("pre_check_failed", "realm missing or current period changed; refresh before committing")
	}
	committed, success := run("commit", spec.args, true)
	if !success {
		return fail("command_failed", "period commit outcome uncertain; inspect realm before any manual retry")
	}
	id, ok := committed["id"].(string)
	epoch, epochOK := committed["epoch"].(json.Number)
	epochValue, epochErr := epoch.Int64()
	if !ok || !syncFlowToken(id) || committed["realm_id"] != realm || !epochOK || epochErr != nil || epochValue <= 0 {
		return fail("post_check_failed", "commit returned an unverifiable period; inspect realm before retrying")
	}
	actualRealm, success := run("realm_post_check", realmArgs, false)
	if !success || actualRealm == nil || actualRealm["id"] != realm || actualRealm["current_period"] != id {
		return fail("post_check_failed", "committed period is not verified as realm current period")
	}
	current, success := run("period_post_check", spec.check, false)
	if !success || !reflect.DeepEqual(current, committed) {
		return fail("post_check_failed", "current period differs from commit result; inspect concurrent changes before retrying")
	}
	return cephdomain.ActionResult{}, nil
}
