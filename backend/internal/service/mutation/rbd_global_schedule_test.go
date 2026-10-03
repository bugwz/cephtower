package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func TestGlobalMirrorScheduleCommands(t *testing.T) {
	for _, tc := range []struct {
		action, interval string
		args             []string
	}{
		{"mirror-schedule-add", "1h", []string{"mirror", "snapshot", "schedule", "add", "1h"}},
		{"mirror-schedule-remove", "1h", []string{"mirror", "snapshot", "schedule", "remove", "1h"}},
		{"mirror-schedule-remove", "", []string{"mirror", "snapshot", "schedule", "remove"}},
	} {
		p := map[string]any{"action": tc.action}
		if tc.interval != "" {
			p["interval"] = tc.interval
		}
		r := Request{Action: "rbd_mirroring.global_schedule", Parameters: p}
		cmd, err := build(r, p)
		if err != nil || cmd.binary != executor.BinaryRBD || !reflect.DeepEqual(cmd.args, tc.args) || !reflect.DeepEqual(cmd.check, []string{"mirror", "snapshot", "schedule", "list", "--recursive", "--format", "json"}) {
			t.Fatal(cmd, err)
		}
		present := []byte(`[{"pool":"-","namespace":"-","image":"-","items":[{"interval":"60m","start_time":""}]}]`)
		otherScope := []byte(`[{"pool":"images","namespace":"-","image":"-","items":[{"interval":"1h","start_time":""}]}]`)
		if rbdMirrorScheduleReadbackMatches(r, present) != (tc.action == "mirror-schedule-add") || rbdMirrorScheduleReadbackMatches(r, otherScope) != (tc.action == "mirror-schedule-remove") {
			t.Fatal("wrong scope readback", tc)
		}
	}
	for _, p := range []map[string]any{{"action": "flatten"}, {"action": "mirror-schedule-add"}, {"action": "mirror-schedule-remove", "start_time": "01:00"}, {"action": "mirror-schedule-add", "interval": "0h"}} {
		if _, err := build(Request{Action: "rbd_mirroring.global_schedule"}, p); err == nil {
			t.Fatal("accepted", p)
		}
	}
	for _, key := range []string{"pool", "namespace", "image", "image_spec"} {
		p := map[string]any{"action": "mirror-schedule-remove", key: ""}
		if _, err := build(Request{Action: "rbd_mirroring.global_schedule"}, p); err == nil {
			t.Fatal("accepted scope", key)
		}
	}
}

func TestGlobalMirrorScheduleExecutionAndUncertainReadback(t *testing.T) {
	s, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "rbd_mirroring.global_schedule", ResourceKey: "global/snapshot-schedule", Parameters: map[string]any{"action": "mirror-schedule-add", "interval": "1h", "start_time": "01:00"}}
	e := &directoryRenameExecutor{outputs: map[string]string{r.Action: "", r.Action + ".post_check": `[{"pool":"-","namespace":"-","image":"-","items":[{"interval":"60m","start_time":"01:00:00"}]}]`}}
	s.executor = e
	if _, err := s.Execute(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	for _, response := range []string{`null`, `[]`, `{}`, `[{"pool":"images","namespace":"-","image":"-","items":[{"interval":"1h","start_time":"01:00:00"}]}]`} {
		e.outputs[r.Action+".post_check"] = response
		_, err := s.Execute(context.Background(), r)
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != "post_check_failed" || failure.Retryable {
			t.Fatal(response, err)
		}
	}
}
