package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestPoolRenameReadback(t *testing.T) {
	s, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "pool.update", ResourceKey: "pool/old", Parameters: map[string]any{"operation": "rename", "name": "new"}}
	spec, err := build(r, r.Parameters)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(spec.check, []string{"osd", "pool", "ls", "--format", "json"}) {
		t.Fatalf("post-check = %v", spec.check)
	}
	e := &directoryRenameExecutor{outputs: map[string]string{r.Action: "renamed"}}
	s.executor = e
	for _, good := range []string{`["new"]`, `["keep", "new"]`} {
		e.outputs[r.Action+".post_check"] = good
		if _, err := s.Execute(context.Background(), r); err != nil {
			t.Fatal(good, err)
		}
	}
	for _, bad := range []string{`[]`, `null`, `{}`, `[null]`, `[""]`, `["old"]`, `["old","new"]`, `["new","new"]`, `["new"] {}`} {
		e.outputs[r.Action+".post_check"] = bad
		_, err := s.Execute(context.Background(), r)
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != "post_check_failed" || failure.Retryable {
			t.Fatal(bad, err)
		}
	}
}
