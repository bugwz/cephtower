package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestSMBAuthDeletion(t *testing.T) {
	for _, action := range []string{"smb_join_auth.delete", "smb_usersgroups.delete"} {
		t.Run(action, func(t *testing.T) {
			service, _, clusterID := newCephUserService(t)
			key := "smb/join/auth/target"
			if action == "smb_usersgroups.delete" {
				key = "smb/usersgroup/target"
			}
			request := Request{ClusterID: clusterID, Action: action, ResourceKey: key, Parameters: map[string]any{"name": "different"}}
			runner := &directoryRenameExecutor{outputs: map[string]string{action: `{"success":true}`, action + ".post_check": `{"resources":[]}`}}
			service.executor = runner
			if _, err := service.Execute(context.Background(), request); err != nil {
				t.Fatal(err)
			}
			resourceType, idField := smbAuthResourceType(action)
			var payload map[string]any
			if err := json.Unmarshal(runner.specs[0].Stdin, &payload); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(payload, map[string]any{"resource_type": resourceType, idField: "target", "intent": "removed"}) {
				t.Fatal(payload)
			}
			if len(runner.specs) != 2 || !runner.specs[0].Mutating || runner.specs[1].Mutating || !reflect.DeepEqual(runner.specs[1].Args, []string{"smb", "show", resourceType, "--results=full", "--password-filter=hidden", "--format", "json"}) {
				t.Fatal(runner.specs)
			}
			for _, bad := range []string{`null`, `{}`, `{"resources":null}`, `{"resources":[null]}`, `{"resources":[]} {}`, `{"resources":[{"resource_type":"` + resourceType + `","` + idField + `":"target"}]}`, `{"resources":[{"resource_type":"wrong","` + idField + `":"other"}]}`} {
				runner.outputs[action+".post_check"] = bad
				_, err := service.Execute(context.Background(), request)
				var failure *cephdomain.ActionError
				if !errors.As(err, &failure) || failure.Code != "post_check_failed" {
					t.Fatalf("accepted %s: %v", bad, err)
				}
			}
			for _, bad := range []string{`{"success":false}`, `{}`, `null`, `invalid`} {
				runner.outputs[action] = bad
				_, err := service.Execute(context.Background(), request)
				var failure *cephdomain.ActionError
				if !errors.As(err, &failure) || failure.Code != "ceph_command_failed" {
					t.Fatalf("accepted apply %s: %v", bad, err)
				}
			}
			request.ResourceKey = "smb/join/auth/bad.id"
			if _, err := build(request, request.Parameters); err == nil {
				t.Fatal("invalid identifier accepted")
			}
		})
	}
}

func TestSMBJoinAuthCreation(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	request := Request{ClusterID: clusterID, Action: "smb_join_auth.create", ResourceKey: "smb/join/auth/target", Parameters: map[string]any{"name": "target", "username": "admin", "password": " secret \n", "linked_to_cluster": "cluster-a"}}
	good := `{"resources":[{"resource_type":"ceph.smb.join.auth","auth_id":"target","intent":"present","auth":{"username":"admin","password":"***"},"linked_to_cluster":"cluster-a"}]}`
	runner := &directoryRenameExecutor{outputs: map[string]string{request.Action + ".pre_check": `{"resources":[]}`, request.Action: `{"success":true}`, request.Action + ".post_check": good}}
	service.executor = runner
	if _, err := service.Execute(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if len(runner.specs) != 3 || !reflect.DeepEqual(runner.specs[1].Args, []string{"smb", "apply", "-i", "-", "--password-filter-out=hidden", "--format", "json"}) {
		t.Fatal(runner.specs)
	}
	var payload map[string]any
	if json.Unmarshal(runner.specs[1].Stdin, &payload) != nil || payload["auth"].(map[string]any)["password"] != " secret \n" {
		t.Fatal("password changed")
	}
	runner.outputs[request.Action+".pre_check"] = good
	runner.specs = nil
	if _, err := service.Execute(context.Background(), request); err == nil || len(runner.specs) != 1 {
		t.Fatal("existing credential overwritten")
	}
	runner.outputs[request.Action+".pre_check"] = `{"resources":[]}`
	runner.failID = request.Action
	if _, err := service.Execute(context.Background(), request); err == nil || err.Error() != "SMB credential write failed" {
		t.Fatalf("native error was not replaced: %v", err)
	}
	for _, bad := range []string{`null`, `{"resources":[]}`, `{"resources":[{"resource_type":"ceph.smb.join.auth","auth_id":"target","auth":{"username":"other"}}]}`, good + ` {}`} {
		if smbJoinAuthCreated(request, []byte(bad)) {
			t.Fatal("invalid creation readback accepted")
		}
	}
	for _, params := range []map[string]any{{"name": "bad.id", "username": "a", "password": "p"}, {"name": "a", "username": "a"}, {"name": "a", "username": "a", "password": "p", "linked_to_cluster": "bad.id"}} {
		if _, err := smbJoinAuthCreateJSON(params); err == nil {
			t.Fatal("invalid creation accepted")
		}
	}
}

func TestSMBJoinAuthUpdate(t *testing.T) {
	service, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "smb_join_auth.update", ResourceKey: "smb/join/auth/target", Parameters: map[string]any{"name": "target", "username": "new", "password": "replacement"}}
	old := `{"resources":[{"resource_type":"ceph.smb.join.auth","auth_id":"target","auth":{"username":"old","password":"***"},"linked_to_cluster":"a"}]}`
	updated := `{"resources":[{"resource_type":"ceph.smb.join.auth","auth_id":"target","auth":{"username":"new","password":"***"}}]}`
	runner := &directoryRenameExecutor{outputs: map[string]string{r.Action + ".pre_check": old, r.Action: `{"success":true}`, r.Action + ".post_check": updated}}
	service.executor = runner
	if _, err := service.Execute(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if json.Unmarshal(runner.specs[1].Stdin, &payload) != nil || payload["linked_to_cluster"] != nil || payload["auth"].(map[string]any)["password"] != "replacement" {
		t.Fatal("replacement payload invalid")
	}
	for _, before := range []string{`{"resources":[]}`, `null`, `{}`, `{"resources":[null]}`} {
		runner.specs = nil
		runner.outputs[r.Action+".pre_check"] = before
		if _, err := service.Execute(context.Background(), r); err == nil || len(runner.specs) != 1 {
			t.Fatal("invalid precheck permitted update")
		}
	}
	runner.outputs[r.Action+".pre_check"] = old
	runner.outputs[r.Action+".post_check"] = old
	if _, err := service.Execute(context.Background(), r); err == nil {
		t.Fatal("unchanged credentials accepted")
	}
	runner.outputs[r.Action+".post_check"] = `{"resources":[{"resource_type":"ceph.smb.join.auth","auth_id":"target","auth":{"username":"new","password":"***"},"linked_to_cluster":false}]}`
	if _, err := service.Execute(context.Background(), r); err == nil {
		t.Fatal("malformed cluster binding reported as successful unlink")
	}
	r.Parameters["name"] = "different"
	if _, err := build(r, r.Parameters); err == nil {
		t.Fatal("identity mismatch accepted")
	}
}

func TestSMBJoinAuthLinkedClusterReadback(t *testing.T) {
	request := Request{Parameters: map[string]any{"name": "target", "username": "admin"}}
	for _, wanted := range []string{"", "cluster-a"} {
		delete(request.Parameters, "linked_to_cluster")
		if wanted != "" {
			request.Parameters["linked_to_cluster"] = wanted
		}
		for _, field := range []string{"", `,"linked_to_cluster":null`, `,"linked_to_cluster":"cluster-a"`, `,"linked_to_cluster":"cluster-b"`, `,"linked_to_cluster":""`, `,"linked_to_cluster":false`, `,"linked_to_cluster":1`, `,"linked_to_cluster":[]`, `,"linked_to_cluster":{}`} {
			raw := `{"resources":[{"resource_type":"ceph.smb.join.auth","auth_id":"target","auth":{"username":"admin"}` + field + `}]}`
			want := wanted == "" && (field == "" || field == `,"linked_to_cluster":null`) || wanted == "cluster-a" && field == `,"linked_to_cluster":"cluster-a"`
			if got := smbJoinAuthCreated(request, []byte(raw)); got != want {
				t.Fatalf("wanted cluster %q, readback %s: got %v, want %v", wanted, field, got, want)
			}
		}
	}
}
