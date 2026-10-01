package mutation

import (
	"context"
	"encoding/base64"
	"reflect"
	"testing"
)

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
