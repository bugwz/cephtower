package mutation

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestSMBUsersGroupsCreate(t *testing.T) {
	s, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "smb_usersgroups.create", ResourceKey: "smb/usersgroup/target", Parameters: map[string]any{"name": "target", "users": []any{map[string]any{"name": "alice", "password": " secret "}}, "groups": []string{"staff"}}}
	good := `{"resources":[{"resource_type":"ceph.smb.usersgroups","users_groups_id":"target","values":{"users":[{"name":"alice","password":"***"}],"groups":[{"name":"staff"}]}}]}`
	e := &directoryRenameExecutor{outputs: map[string]string{r.Action + ".pre_check": `{"resources":[]}`, r.Action: `{"success":true}`, r.Action + ".post_check": good}}
	s.executor = e
	if _, err := s.Execute(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if len(e.specs) != 3 || !reflect.DeepEqual(e.specs[1].Args, []string{"smb", "apply", "-i", "-", "--password-filter-out=hidden", "--format", "json"}) {
		t.Fatal(e.specs)
	}
	var payload map[string]any
	if json.Unmarshal(e.specs[1].Stdin, &payload) != nil || payload["values"].(map[string]any)["users"].([]any)[0].(map[string]any)["password"] != " secret " {
		t.Fatal("password changed")
	}
	e.outputs[r.Action+".pre_check"] = good
	e.specs = nil
	if _, err := s.Execute(context.Background(), r); err == nil || len(e.specs) != 1 {
		t.Fatal("existing resource overwritten")
	}
	for _, bad := range []string{`null`, `{}`, `{"resources":[]}`, good + ` {}`, `{"resources":[{"resource_type":"ceph.smb.usersgroups","users_groups_id":"target","values":{"users":[],"groups":[]}}]}`} {
		if smbUsersGroupsCreated(r, []byte(bad)) {
			t.Fatal("invalid readback accepted")
		}
	}
	for _, users := range []any{nil, []any{map[string]any{"name": "a"}}, []any{map[string]any{"name": "a", "password": "p"}, map[string]any{"name": "a", "password": "q"}}} {
		r.Parameters["users"] = users
		if _, err := smbUsersGroupsJSON(r.Parameters); err == nil {
			t.Fatal("invalid users accepted")
		}
	}
}

func TestSMBUsersGroupsUpdate(t *testing.T) {
	s, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "smb_usersgroups.update", ResourceKey: "smb/usersgroup/target", Parameters: map[string]any{"name": "target", "users": []any{map[string]any{"name": "new", "password": "replacement"}}, "groups": []string{}}}
	old := `{"resources":[{"resource_type":"ceph.smb.usersgroups","users_groups_id":"target","values":{"users":[{"name":"old","password":"***"}],"groups":[{"name":"staff"}]},"linked_to_cluster":"a"}]}`
	updated := `{"resources":[{"resource_type":"ceph.smb.usersgroups","users_groups_id":"target","values":{"users":[{"name":"new","password":"***"}],"groups":[]}}]}`
	e := &directoryRenameExecutor{outputs: map[string]string{r.Action + ".pre_check": old, r.Action: `{"success":true}`, r.Action + ".post_check": updated}}
	s.executor = e
	if _, err := s.Execute(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if json.Unmarshal(e.specs[1].Stdin, &payload) != nil || payload["linked_to_cluster"] != nil {
		t.Fatal("binding not cleared")
	}
	for _, before := range []string{`{"resources":[]}`, `null`, `{}`, `{"resources":[null]}`} {
		e.specs = nil
		e.outputs[r.Action+".pre_check"] = before
		if _, err := s.Execute(context.Background(), r); err == nil || len(e.specs) != 1 {
			t.Fatal("invalid precheck accepted")
		}
	}
	e.outputs[r.Action+".pre_check"] = old
	e.outputs[r.Action+".post_check"] = old
	if _, err := s.Execute(context.Background(), r); err == nil {
		t.Fatal("unchanged users accepted")
	}
}

func TestSMBUsersGroupsEmptyUsers(t *testing.T) {
	s, _, id := newCephUserService(t)
	empty := `{"resources":[{"resource_type":"ceph.smb.usersgroups","users_groups_id":"target","values":{"users":[],"groups":[]}}]}`
	old := `{"resources":[{"resource_type":"ceph.smb.usersgroups","users_groups_id":"target","values":{"users":[{"name":"old","password":"***"}],"groups":[]}}]}`
	for _, action := range []string{"smb_usersgroups.create", "smb_usersgroups.update"} {
		r := Request{ClusterID: id, Action: action, ResourceKey: "smb/usersgroup/target", Parameters: map[string]any{"name": "target", "users": []any{}, "groups": []string{}}}
		before := old
		if action == "smb_usersgroups.create" {
			before = `{"resources":[]}`
		}
		e := &directoryRenameExecutor{outputs: map[string]string{action + ".pre_check": before, action: `{"success":true}`, action + ".post_check": empty}}
		s.executor = e
		if _, err := s.Execute(context.Background(), r); err != nil {
			t.Fatal(err)
		}
		var payload map[string]any
		if json.Unmarshal(e.specs[1].Stdin, &payload) != nil {
			t.Fatal("invalid payload")
		}
		users, ok := payload["values"].(map[string]any)["users"].([]any)
		if !ok || len(users) != 0 {
			t.Fatal("empty users not serialized as array")
		}
		e.outputs[action+".post_check"] = old
		if _, err := s.Execute(context.Background(), r); err == nil {
			t.Fatal("stale users accepted after clearing")
		}
	}
}
