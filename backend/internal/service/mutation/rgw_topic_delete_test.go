package mutation

import (
	"context"
	"encoding/base64"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type topicDeleteExecutor struct {
	calls   []executor.CommandSpec
	bodies  map[string]string
	failure string
}

func (e *topicDeleteExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.calls = append(e.calls, spec)
	stage := strings.TrimPrefix(spec.ID, "rgw_topic.delete.")
	if stage == e.failure {
		return executor.CommandResult{}, errors.New("native error")
	}
	return executor.CommandResult{Stdout: []byte(e.bodies[stage])}, nil
}

func TestTopicDeletionNativeStages(t *testing.T) {
	for _, scope := range []string{"", "team", "RGW12345678901234567"} {
		for _, scenario := range []string{"success", "stale", "missing version", "wrong metadata key", "wrong name", "legacy", "changed configuration", "malformed bindings", "pre_check", "topic_check", "delete", "post_check", "still present", "null list", "duplicate list"} {
			t.Run(scope+"/"+scenario, func(t *testing.T) {
				service, _, cluster := newCephUserService(t)
				key := scope + ":events"
				id := base64.RawURLEncoding.EncodeToString([]byte(key))
				version := `{"tag":"revision","ver":9007199254740993}`
				data := `{"name":"events","owner":"u","arn":"arn:topic","dest":{"push_endpoint":"https://user:secret@host","persistent":true}}`
				before := `{"key":"topic:` + key + `","ver":` + version + `,"data":` + data + `}`
				topic := strings.TrimSuffix(data, "}") + `,"subscribed_buckets":["bucket"]}`
				runner := &topicDeleteExecutor{bodies: map[string]string{"pre_check": before, "topic_check": topic, "post_check": `["other:events"]`}}
				params := map[string]any{"topic_id": id, "expected_version": version}
				count := 4
				switch scenario {
				case "success":
				case "stale":
					params["expected_version"] = `{"tag":"revision","ver":9007199254740992}`
					count = 1
				case "missing version":
					runner.bodies["pre_check"] = strings.Replace(before, version, `{}`, 1)
					count = 1
				case "wrong metadata key":
					runner.bodies["pre_check"] = strings.Replace(before, "topic:"+key, "topic:wrong", 1)
					count = 1
				case "wrong name":
					runner.bodies["pre_check"] = strings.Replace(before, `"name":"events"`, `"name":"other"`, 1)
					count = 1
				case "legacy":
					runner.bodies["topic_check"] = data
					count = 2
				case "changed configuration":
					runner.bodies["topic_check"] = strings.Replace(topic, "secret", "changed", 1)
					count = 2
				case "malformed bindings":
					runner.bodies["topic_check"] = strings.Replace(topic, `["bucket"]`, `[42]`, 1)
					count = 2
				case "still present":
					runner.bodies["post_check"] = `["` + key + `"]`
				case "null list":
					runner.bodies["post_check"] = `null`
				case "duplicate list":
					runner.bodies["post_check"] = `["other","other"]`
				default:
					runner.failure = scenario
					count = map[string]int{"pre_check": 1, "topic_check": 2, "delete": 3, "post_check": 4}[scenario]
				}
				service.executor = runner
				_, err := service.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_topic.delete", ResourceKey: "rgw/topic/" + id, Parameters: params})
				if len(runner.calls) != count {
					t.Fatalf("calls=%d expected=%d error=%v", len(runner.calls), count, err)
				}
				if scenario == "success" {
					if err != nil {
						t.Fatal(err)
					}
				} else {
					var ae *cephdomain.ActionError
					if !errors.As(err, &ae) || ae.Retryable {
						t.Fatalf("wrong error %v", err)
					}
				}
				want := [][]string{{"metadata", "get", "topic:" + key, "--format", "json"}, {"topic", "get", "--tenant", scope, "--topic", "events", "--format", "json"}, {"topic", "rm", "--tenant", scope, "--topic", "events", "--format", "json"}, {"metadata", "list", "topic", "--format", "json"}}
				for i, call := range runner.calls {
					if call.Binary != executor.BinaryRGWAdmin || call.Mutating != (i == 2) || !reflect.DeepEqual(call.Args, want[i]) {
						t.Fatalf("wrong stage %+v", call)
					}
				}
			})
		}
	}
}

func TestTopicDeleteRejectsInvalidIdentity(t *testing.T) {
	encode := func(key string) string { return base64.RawURLEncoding.EncodeToString([]byte(key)) }
	for _, id := range []string{"", "bad=", encode("no-scope"), encode("scope:"), encode("scope:-option"), encode("-scope:topic"), encode("scope:bad\nname")} {
		if _, err := build(Request{Action: "rgw_topic.delete"}, map[string]any{"topic_id": id, "expected_version": `{"tag":"t","ver":1}`}); err == nil {
			t.Fatalf("invalid identity accepted %q", id)
		}
	}
}
