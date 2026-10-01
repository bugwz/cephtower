package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestNFSExportUpdatePreservesNativeAttributes(t *testing.T) {
	service, _, id := newCephUserService(t)
	runner := &directoryRenameExecutor{outputs: map[string]string{"nfs_export.update.pre_check": `[{"export_id":2,"cluster_id":"nfs-a","pseudo":"/old","path":"/data","access_type":"RO","squash":"root_squash","protocols":[4],"transports":["TCP"],"clients":[{"addresses":["10.0.0.0/8"],"access_type":"RO"}],"fsal":{"name":"CEPH","fs_name":"cephfs","user_id":"nfs.user","cmount_path":"/"}}]`, "nfs_export.update.post_check": `[]`}}
	service.executor = runner
	runner.outputs["nfs_export.update.post_check"] = `[{"export_id":2,"cluster_id":"nfs-a","pseudo":"/renamed","path":"/new","access_type":"RW","fsal":{"name":"CEPH","fs_name":"cephfs"}}]`
	request := Request{ClusterID: id, Action: "nfs_export.update", ResourceKey: "nfs/export/" + base64.RawURLEncoding.EncodeToString([]byte("nfs-a\x002")), Parameters: map[string]any{"cluster": "nfs-a", "pseudo": "/renamed", "path": "/new", "filesystem": "cephfs", "read_only": false}}
	if _, err := service.Execute(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if len(runner.specs) != 3 {
		t.Fatalf("commands=%+v", runner.specs)
	}
	var payload map[string]any
	if err := json.Unmarshal(runner.specs[1].Stdin, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["export_id"] != float64(2) || payload["pseudo"] != "/renamed" || payload["path"] != "/new" || payload["access_type"] != "RW" || payload["squash"] != "root_squash" {
		t.Fatalf("payload=%v", payload)
	}
	if !reflect.DeepEqual(payload["clients"], []any{map[string]any{"addresses": []any{"10.0.0.0/8"}, "access_type": "RO"}}) || payload["fsal"].(map[string]any)["user_id"] != "nfs.user" {
		t.Fatalf("attributes lost: %v", payload)
	}
	for _, post := range []string{`[]`, runner.outputs["nfs_export.update.pre_check"], `[{"export_id":3,"pseudo":"/renamed","path":"/new","access_type":"RW","fsal":{"name":"CEPH","fs_name":"cephfs"}}]`} {
		runner.outputs["nfs_export.update.post_check"] = post
		_, err := service.Execute(context.Background(), request)
		var actionErr *cephdomain.ActionError
		if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" {
			t.Fatalf("unverified update accepted: %v", err)
		}
	}
	runner.specs = nil
	request.Parameters["cluster"] = "nfs-b"
	if _, err := service.Execute(context.Background(), request); err == nil || len(runner.specs) != 0 {
		t.Fatalf("cross-cluster update accepted: %v", err)
	}
}

func TestNFSExportSquashOptions(t *testing.T) {
	for _, squash := range []string{"root_squash", "root_id_squash", "all_squash", "no_root_squash"} {
		params := map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "/", "filesystem": "cephfs", "squash": squash}
		cmd, err := build(Request{Action: "nfs_export.create", ResourceKey: "nfs/export"}, params)
		if err != nil {
			t.Fatal(err)
		}
		var payload map[string]any
		if err := json.Unmarshal(cmd.stdin, &payload); err != nil {
			t.Fatal(err)
		}
		if payload["squash"] != squash {
			t.Fatalf("squash lost: %v", payload)
		}
		existing := map[string]any{"fsal": map[string]any{"name": "CEPH"}, "access_type": "RW", "squash": "root_squash"}
		updated, err := nfsExportUpdateJSON(existing, params)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(updated, &payload); err != nil {
			t.Fatal(err)
		}
		if payload["squash"] != squash {
			t.Fatalf("update lost squash: %v", payload)
		}
	}
	if _, err := build(Request{Action: "nfs_export.create", ResourceKey: "nfs/export"}, map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "/", "filesystem": "cephfs", "squash": "invalid"}); err == nil {
		t.Fatal("invalid squash accepted")
	}
}

func TestNFSExportPaths(t *testing.T) {
	for _, action := range []string{"nfs_export.create", "nfs_export.update"} {
		for _, pair := range [][2]string{{"/", "/data"}, {"relative", "/data"}, {"/share", "relative"}} {
			_, err := build(Request{Action: action, ResourceKey: "nfs/export"}, map[string]any{"cluster": "nfs-a", "pseudo": pair[0], "path": pair[1], "filesystem": "cephfs"})
			if err == nil {
				t.Fatalf("%s accepted invalid paths: %v", action, pair)
			}
		}
		if _, err := build(Request{Action: action, ResourceKey: "nfs/export"}, map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "/", "filesystem": "cephfs"}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNFSExportSecurityLabel(t *testing.T) {
	for _, label := range []any{true, false, "false"} {
		params := map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "/", "filesystem": "cephfs", "security_label": label}
		cmd, err := build(Request{Action: "nfs_export.create", ResourceKey: "nfs/export"}, params)
		if label == "false" {
			if err == nil {
				t.Fatal("string label accepted")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		var payload map[string]any
		if err := json.Unmarshal(cmd.stdin, &payload); err != nil {
			t.Fatal(err)
		}
		if payload["security_label"] != label {
			t.Fatalf("label lost: %v", payload)
		}
	}
}

func TestNFSExportProtocols(t *testing.T) {
	for _, value := range []any{[]int{3}, []int{4}, []int{4, 3}, []any{float64(3), float64(4)}, []any{json.Number("4")}} {
		p := map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "/", "filesystem": "cephfs", "protocols": value}
		for _, action := range []string{"nfs_export.create", "nfs_export.update"} {
			cmd, err := build(Request{Action: action, ResourceKey: "nfs/export"}, p)
			if err != nil {
				t.Fatal(err)
			}
			var record map[string]any
			if err := json.Unmarshal(cmd.stdin, &record); err != nil {
				t.Fatal(err)
			}
			if !nfsExportAttributesMatch(record, p) {
				t.Fatal("protocol payload mismatch")
			}
			record["access_type"] = "RW"
			record["protocols"] = []int{3, 4}
			data, err := nfsExportUpdateJSON(record, p)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(data, &record); err != nil {
				t.Fatal(err)
			}
			if !nfsExportAttributesMatch(record, p) {
				t.Fatal("update protocols lost")
			}
			record["protocols"] = nil
			if nfsExportAttributesMatch(record, p) {
				t.Fatal("missing readback accepted")
			}
		}
	}
	for _, value := range []any{nil, []int{}, []int{3, 3}, []int{2}, []int{5}, "4", []string{"4"}, []float64{3.5}, []any{true}} {
		p := map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "/", "filesystem": "cephfs", "protocols": value}
		if _, err := build(Request{Action: "nfs_export.create"}, p); err == nil {
			t.Fatalf("invalid protocols accepted: %v", value)
		}
	}
	left, err := nfsProtocols([]int{4, 3})
	right, otherErr := nfsProtocols([]int{3, 4})
	if err != nil || otherErr != nil || left != right {
		t.Fatal("protocol order must not matter")
	}
}

func TestNFSExportCreateDoesNotOverwriteExistingPseudo(t *testing.T) {
	for _, output := range []string{`[{"pseudo":"/share"}]`, `[{"pseudo":"/share/"}]`, `[{}]`, `null`, `[{"pseudo":"/other","cluster_id":"nfs-b"}]`} {
		service, _, id := newCephUserService(t)
		runner := &directoryRenameExecutor{outputs: map[string]string{"nfs_export.create.pre_check": output}}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "nfs_export.create", ResourceKey: "nfs/export", Parameters: map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "/", "filesystem": "cephfs"}})
		if err == nil || len(runner.specs) != 1 || runner.specs[0].Mutating {
			t.Fatalf("unexpected create: %v %+v", err, runner.specs)
		}
	}
	for _, output := range []string{`[]`, `[{"pseudo":"/other","cluster_id":"nfs-a"}]`} {
		if !nfsExportCreateAvailable([]byte(output), "nfs-a", "/share") {
			t.Fatal("available path rejected")
		}
	}
}

func TestNFSExportCreationReadback(t *testing.T) {
	for _, tt := range []struct {
		post  string
		valid bool
	}{
		{`[{"export_id":1,"cluster_id":"nfs-a","pseudo":"/share","path":"/","fsal":{"name":"CEPH","fs_name":"cephfs"},"access_type":"RO"}]`, true},
		{`[]`, false},
		{`[{"export_id":1,"cluster_id":"nfs-a","pseudo":"/share","path":"/other","fsal":{"name":"CEPH","fs_name":"cephfs"},"access_type":"RO"}]`, false},
		{`[{"export_id":1,"cluster_id":"nfs-b","pseudo":"/share","path":"/","fsal":{"name":"CEPH","fs_name":"cephfs"},"access_type":"RO"}]`, false},
	} {
		service, _, id := newCephUserService(t)
		runner := &directoryRenameExecutor{outputs: map[string]string{"nfs_export.create.pre_check": "[]", "nfs_export.create.post_check": tt.post}}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "nfs_export.create", ResourceKey: "nfs/export", Parameters: map[string]any{"cluster": "nfs-a", "pseudo": "/share/", "path": "/", "filesystem": "cephfs", "read_only": true}})
		if tt.valid {
			if err != nil {
				t.Fatal(err)
			}
		} else {
			var actionErr *cephdomain.ActionError
			if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" {
				t.Fatalf("unverified creation accepted: %v", err)
			}
		}
	}
}

func TestNFSExportDeleteResolvesNativePseudo(t *testing.T) {
	for _, tt := range []struct {
		name, output string
		valid        bool
	}{
		{"matching id", `[{"export_id":1,"pseudo":"/wrong","cluster_id":"nfs-a"},{"export_id":2,"pseudo":"/right","cluster_id":"nfs-a"}]`, true},
		{"missing", `[{"export_id":1,"pseudo":"/wrong"}]`, false},
		{"wrong cluster", `[{"export_id":2,"pseudo":"/right","cluster_id":"nfs-b"}]`, false},
		{"duplicate", `[{"export_id":2,"pseudo":"/right"},{"export_id":2,"pseudo":"/other"}]`, false},
		{"trailing", `[{"export_id":2,"pseudo":"/right"}] {}`, false},
		{"bad path", `[{"export_id":2,"pseudo":"relative"}]`, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			service, _, id := newCephUserService(t)
			runner := &directoryRenameExecutor{outputs: map[string]string{"nfs_export.delete.pre_check": tt.output, "nfs_export.delete.post_check": "[]"}}
			service.executor = runner
			_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "nfs_export.delete", ResourceKey: "nfs/export/" + base64.RawURLEncoding.EncodeToString([]byte("nfs-a\x002"))})
			if tt.valid {
				if err != nil {
					t.Fatal(err)
				}
				if len(runner.specs) != 3 || !reflect.DeepEqual(runner.specs[1].Args, []string{"nfs", "export", "rm", "nfs-a", "/right"}) {
					t.Fatalf("commands=%+v", runner.specs)
				}
			} else if err == nil || len(runner.specs) != 1 {
				t.Fatalf("unexpected mutation: err=%v commands=%+v", err, runner.specs)
			}
		})
	}
}

func TestNFSExportDeletionReadback(t *testing.T) {
	request := Request{ResourceKey: "nfs/export/" + base64.RawURLEncoding.EncodeToString([]byte("nfs-a\x002"))}
	for _, data := range []string{`[]`, `[{"export_id":1,"cluster_id":"nfs-a","pseudo":"/other"}]`} {
		if !nfsExportDeleted(request, []byte(data)) {
			t.Fatalf("valid deletion rejected: %s", data)
		}
	}
	for _, data := range []string{`null`, `{}`, `[{}]`, `[{"export_id":2}]`, `[{"export_id":1,"cluster_id":"nfs-b"}]`, `[] {}`, `["/export"]`} {
		if nfsExportDeleted(request, []byte(data)) {
			t.Fatalf("unverified deletion accepted: %s", data)
		}
	}
}
