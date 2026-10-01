package mutation

import (
	"context"
	"errors"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestErasureProfileCreationReadback(t *testing.T) {
	s, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "erasure_code_profile.create", ResourceKey: "erasure-code-profile", Parameters: map[string]any{
		"name": "ec-isa", "plugin": "isa", "k": "4", "m": "2", "crush-root": "default", "crush-device-class": "ssd",
	}}
	e := &directoryRenameExecutor{outputs: map[string]string{r.Action: "created"}}
	s.executor = e
	e.outputs[r.Action+".post_check"] = `{"plugin":"isa","k":"4","m":"2","crush-root":"default","crush-device-class":"ssd","technique":"reed_sol_van"}`
	if _, err := s.Execute(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{`null`, `{}`, `[]`, `{"plugin":"isa"}`, `{"plugin":"isa","k":4}`, `{"plugin":"isa","k":"4","m":"2","crush-root":"default","crush-device-class":"hdd"}`, e.outputs[r.Action+".post_check"] + `{}`} {
		e.outputs[r.Action+".post_check"] = bad
		_, err := s.Execute(context.Background(), r)
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != "post_check_failed" || failure.Retryable {
			t.Fatal(bad, err)
		}
	}
}

func TestErasureProfileDeletionReadback(t *testing.T) {
	s, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "erasure_code_profile.delete", ResourceKey: "erasure-code-profile/ec-isa"}
	e := &directoryRenameExecutor{outputs: map[string]string{r.Action: "removed"}}
	s.executor = e
	for _, good := range []string{`[]`, `["default","other"]`} {
		e.outputs[r.Action+".post_check"] = good
		if _, err := s.Execute(context.Background(), r); err != nil {
			t.Fatal(good, err)
		}
	}
	for _, bad := range []string{`null`, `{}`, `[null]`, `[""]`, `["ec-isa"]`, `["default","ec-isa"]`, `[] {}`, `[1]`} {
		e.outputs[r.Action+".post_check"] = bad
		_, err := s.Execute(context.Background(), r)
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != "post_check_failed" || failure.Retryable {
			t.Fatal(bad, err)
		}
	}
}
