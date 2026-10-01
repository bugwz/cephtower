package clusterinspect

import (
	"context"
	"reflect"
	"testing"
)

func TestUpgradeVersions(t *testing.T) {
	s, runner, id := testInspection(t)
	for _, versions := range []string{`[]`, `["20.2.2","19.2.3"]`} {
		runner.output = `{"image":"quay.io/ceph/ceph","registry":"quay.io","versions":` + versions + `}`
		result, err := s.UpgradeVersions(context.Background(), id)
		if err != nil || result.Versions == nil || result.Registry != "quay.io" {
			t.Fatal(result, err)
		}
	}
	if runner.specs[0].Mutating || !reflect.DeepEqual(runner.specs[0].Args, []string{"orch", "upgrade", "ls", "--format", "json"}) {
		t.Fatal(runner.specs)
	}
	for _, invalid := range []string{`null`, `{}`, `{"image":"ceph","registry":"quay.io","versions":null}`, `{"image":"ceph","registry":"quay.io","versions":["latest"]}`} {
		runner.output = invalid
		if _, err := s.UpgradeVersions(context.Background(), id); err == nil {
			t.Fatal("accepted", invalid)
		}
	}
}
