package configinit

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestCreateGeneratesProtectedStableConfiguration(t *testing.T) {
	dir := t.TempDir()
	template := filepath.Join(dir, "template.yaml")
	target := filepath.Join(dir, "app", "config.yaml")
	if err := os.WriteFile(template, []byte("server:\n    dir: \"/opt/cephtower\"\ndatabase:\n    encryption_key: \"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Create(template, target, "./app"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`encryption_key: "[A-Za-z0-9_-]{32}"`).Match(data) || !strings.Contains(string(data), `dir: "./app"`) {
		t.Fatalf("generated config = %s", data)
	}
	info, _ := os.Stat(target)
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o", info.Mode().Perm())
	}
	if err := Create(template, target, "./app"); err == nil {
		t.Fatal("existing config was overwritten")
	}
}
