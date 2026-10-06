package clusterinspect

import (
	"context"
	"reflect"
	"testing"
)

func TestCrushMap(t *testing.T) {
	s, runner, id := testInspection(t)
	runner.output = `{"nodes":[{"id":0,"name":"osd.0","type":"osd","status":"up","crush_weight":1},{"id":-1,"name":"default","type":"root","children":[0]}]}`
	result, err := s.CrushMap(context.Background(), id)
	if err != nil || !reflect.DeepEqual(result.Roots, []int64{-1}) || result.Nodes[0]["status"] != "up" {
		t.Fatal(result, err)
	}
	if runner.specs[0].Mutating || !reflect.DeepEqual(runner.specs[0].Args, []string{"osd", "tree", "--format", "json"}) {
		t.Fatal(runner.specs)
	}
	runner.output = `{"nodes":[]}`
	result, err = s.CrushMap(context.Background(), id)
	if err != nil || result.Roots == nil || len(result.Roots) != 0 {
		t.Fatal(result, err)
	}
	for _, invalid := range []string{
		`null`, `{}`, `{"nodes":null}`, `{"nodes":[{}]}`,
		`{"nodes":[{"id":-1,"name":"a","type":"root","children":[-1]}]}`,
		`{"nodes":[{"id":-1,"name":"a","type":"root","children":[3]}]}`,
		`{"nodes":[{"id":-1,"name":"a","type":"root","children":null}]}`,
		`{"nodes":[{"id":0,"name":"a","type":"osd"},{"id":0,"name":"b","type":"osd"}]}`,
		`{"nodes":[{"id":-1,"name":"a","type":"root","children":[0,0]},{"id":0,"name":"osd.0","type":"osd"}]}`,
	} {
		runner.output = invalid
		if _, err := s.CrushMap(context.Background(), id); err == nil {
			t.Fatal("accepted", invalid)
		}
	}
}

func TestCrushMapAllowsSharedNodesAcrossDifferentParents(t *testing.T) {
	s, runner, id := testInspection(t)
	runner.output = `{"nodes":[{"id":-1,"name":"a","type":"root","children":[0]},{"id":-2,"name":"b","type":"root","children":[0]},{"id":0,"name":"osd.0","type":"osd"}]}`
	result, err := s.CrushMap(context.Background(), id)
	if err != nil || !reflect.DeepEqual(result.Roots, []int64{-1, -2}) {
		t.Fatalf("valid shared placement rejected: %+v %v", result, err)
	}
}
