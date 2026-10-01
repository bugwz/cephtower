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
	if _, err := service.Execute(context.Background(), request); err == nil || err.Error() != "SMB credential creation failed" {
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
