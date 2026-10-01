package mutation

import (
	"context"
	"strings"
	"testing"
)

func TestNFSConfigStringSafety(t *testing.T) {
	for _, field := range []string{"cluster", "pseudo", "path", "filesystem", "rgw_user_id"} {
		for _, value := range []any{"/x\"; Access_Type=RW; #", "/trailing\\", "/line\nnext", "/tab\tx", "/nul\x00x", "/del\x7fx", 3} {
			for _, action := range []string{"nfs_export.create", "nfs_export.update"} {
				p := map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "/data", "filesystem": "fs"}
				p[field] = value
				if _, err := build(Request{Action: action}, p); err == nil {
					t.Fatalf("%s accepted unsafe %s", action, field)
				}
			}
		}
	}
	for _, value := range []string{"/目录/共享", "/data with spaces", "/data;literal", "/it's-safe", "/data{literal}"} {
		p := map[string]any{"cluster": "nfs-a", "pseudo": value, "path": value, "filesystem": "fs"}
		if _, err := build(Request{Action: "nfs_export.create"}, p); err != nil {
			t.Fatalf("safe path rejected: %v", err)
		}
	}
	if _, err := requiredNFSPath(map[string]any{"path": "/" + strings.Repeat("a", 4096)}, "path"); err == nil {
		t.Fatal("oversized path accepted")
	}
}

func TestNFSUnsafeConfigRejectedBeforeCommands(t *testing.T) {
	service, _, id := newCephUserService(t)
	runner := &directoryRenameExecutor{}
	service.executor = runner
	_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "nfs_export.create", ResourceKey: "nfs/export", Parameters: map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "bucket", "fsal_type": "RGW", "rgw_user_id": "owner\";"}})
	if err == nil || len(runner.specs) != 0 {
		t.Fatal("unsafe config reached executor")
	}
}
