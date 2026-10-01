package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestPoolDeletionReadback(t *testing.T) {
	s, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "pool.delete", ResourceKey: "pool/remove-me"}
	spec, err := build(r, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(spec.check, []string{"osd", "pool", "ls", "--format", "json"}) {
		t.Fatalf("post-check = %v", spec.check)
	}
	e := &directoryRenameExecutor{outputs: map[string]string{r.Action: "removed"}}
	s.executor = e
	for _, good := range []string{`[]`, `["keep-me"]`} {
		e.outputs[r.Action+".post_check"] = good
		if _, err := s.Execute(context.Background(), r); err != nil {
			t.Fatal(good, err)
		}
	}
	for _, bad := range []string{`null`, `{}`, `[null]`, `[""]`, `["remove-me"]`, `[] {}`} {
		e.outputs[r.Action+".post_check"] = bad
		_, err := s.Execute(context.Background(), r)
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != "post_check_failed" || failure.Retryable {
			t.Fatal(bad, err)
		}
	}
}
