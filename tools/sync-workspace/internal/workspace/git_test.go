package workspace

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestShouldSyncPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want bool
	}{
		{"Makefile", true},
		{"backend/cmd/main.go", true},
		{"frontend/src/main.tsx", true},
		{"config/config.yaml", true},
		{"tools/sync-workspace/cmd/main.go", false},
		{"app/data/cephtower.db", false},
		{"frontend/node_modules/pkg/index.js", false},
		{"frontend/dist/index.html", false},
		{"backend/internal/webui/frontend/dist/index.html", false},
		{"backend/internal/integration/ceph/testdata/data.json", false},
		{"../outside", false},
		{"/absolute", false},
		{"README.md", false},
		{"AGENTS.md", false},
	}
	for _, test := range tests {
		test := test
		t.Run(test.path, func(t *testing.T) {
			t.Parallel()
			if got := shouldSyncPath(test.path); got != test.want {
				t.Fatalf("shouldSyncPath(%q) = %t, want %t", test.path, got, test.want)
			}
		})
	}
}

func TestScanGitWorkspaceUsesGitScopeAndForcedExclusions(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init")
	writeTestFile(t, repo, ".gitignore", "ignored/\n")
	writeTestFile(t, repo, "backend/main.go", "package main\n")
	writeTestFile(t, repo, "frontend/src/main.ts", "export {}\n")
	writeTestFile(t, repo, "new.txt", "untracked\n")
	writeTestFile(t, repo, "ignored/value.txt", "ignored\n")
	writeTestFile(t, repo, "tools/helper.go", "ignored by tool\n")
	writeTestFile(t, repo, "app/config/config.yaml", "ignored by tool\n")
	runGit(t, repo, "add", ".gitignore", "backend/main.go", "tools/helper.go")

	manifest, err := scanGitWorkspace(context.Background(), repo)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"backend/main.go", "frontend/src/main.ts", "new.txt"} {
		if _, ok := manifest.Files[path]; !ok {
			t.Errorf("expected %s in manifest", path)
		}
	}
	for _, path := range []string{".gitignore", "ignored/value.txt", "tools/helper.go", "app/config/config.yaml"} {
		if _, ok := manifest.Files[path]; ok {
			t.Errorf("did not expect %s in manifest", path)
		}
	}
}

func TestManifestChanges(t *testing.T) {
	t.Parallel()
	previous := sourceManifest{Version: 1, Files: map[string]sourceEntry{
		"same":    {Hash: "1", Kind: "file"},
		"changed": {Hash: "1", Kind: "file"},
		"deleted": {Hash: "1", Kind: "file"},
	}}
	current := sourceManifest{Version: 1, Files: map[string]sourceEntry{
		"same":    {Hash: "1", Kind: "file"},
		"changed": {Hash: "2", Kind: "file"},
		"new":     {Hash: "1", Kind: "file"},
	}}
	changed, deleted := manifestChanges(previous, current)
	if got, want := joinStrings(changed), "changed,new"; got != want {
		t.Fatalf("changed = %q, want %q", got, want)
	}
	if got, want := joinStrings(deleted), "deleted"; got != want {
		t.Fatalf("deleted = %q, want %q", got, want)
	}
}

func runGit(t *testing.T, repo string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", repo}, args...)...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}

func writeTestFile(t *testing.T, root, path, content string) {
	t.Helper()
	fullPath := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func joinStrings(values []string) string {
	result := ""
	for index, value := range values {
		if index > 0 {
			result += ","
		}
		result += value
	}
	return result
}
