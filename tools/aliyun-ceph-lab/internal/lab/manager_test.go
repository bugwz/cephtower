package lab

import (
	"encoding/base64"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"

	"cephtower/tools/aliyun-ceph-lab/internal/config"
	"cephtower/tools/aliyun-ceph-lab/internal/state"
)

func TestGenerateClusterSSHKeyProducesMatchingUniqueKeys(t *testing.T) {
	t.Parallel()
	firstPrivate, firstPublic, err := generateClusterSSHKey("test-cluster")
	if err != nil {
		t.Fatal(err)
	}
	secondPrivate, _, err := generateClusterSSHKey("test-cluster")
	if err != nil {
		t.Fatal(err)
	}
	if firstPrivate == secondPrivate {
		t.Fatal("two cluster creations generated the same private key")
	}

	privatePEM, err := base64.StdEncoding.DecodeString(firstPrivate)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(privatePEM)
	if block == nil {
		t.Fatal("generated private key is not PEM encoded")
	}
	privateKey, err := ssh.ParseRawPrivateKey(privatePEM)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	publicAuthorizedKey, err := base64.StdEncoding.DecodeString(firstPublic)
	if err != nil {
		t.Fatal(err)
	}
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey(publicAuthorizedKey)
	if err != nil {
		t.Fatal(err)
	}
	if string(signer.PublicKey().Marshal()) != string(publicKey.Marshal()) {
		t.Fatal("generated public key does not match private key")
	}
}

func TestJoinNodeFieldKeepsAddressOrder(t *testing.T) {
	t.Parallel()
	nodes := []state.Node{
		{Name: "node-1", PrivateIP: "172.31.0.10"},
		{Name: "node-2", PrivateIP: "172.31.0.11"},
	}
	got := joinNodeField(nodes, func(node state.Node) string { return node.PrivateIP })
	if got != "172.31.0.10,172.31.0.11" || strings.Count(got, ",") != 1 {
		t.Fatalf("joinNodeField() = %q", got)
	}
}

func TestJoinDataDiskCounts(t *testing.T) {
	t.Parallel()
	nodes := []config.Node{
		{Name: "node-1", DataDisks: []config.DataDisk{{}, {}}},
		{Name: "node-2"},
		{Name: "node-3", DataDisks: []config.DataDisk{{}}},
	}
	if got, want := joinDataDiskCounts(nodes), "2,0,1"; got != want {
		t.Fatalf("joinDataDiskCounts() = %q, want %q", got, want)
	}
}

func TestCleanupLocalStateRemovesEntireDotStateDirectory(t *testing.T) {
	t.Parallel()
	stateDir := filepath.Join(t.TempDir(), ".state")
	statePath := filepath.Join(stateDir, "ceph-dev.json")
	knownHostsPath := filepath.Join(stateDir, "ceph-dev.known_hosts")
	for path, content := range map[string]string{
		statePath:      "state",
		knownHostsPath: "host key",
		filepath.Join(stateDir, "log", "node.log"): "log",
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := cleanupLocalState(statePath, knownHostsPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stateDir); !os.IsNotExist(err) {
		t.Fatalf(".state directory still exists or stat failed unexpectedly: %v", err)
	}
}

func TestCleanupLocalStateDoesNotRemoveCustomParentDirectory(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	statePath := filepath.Join(parent, "ceph-dev.json")
	knownHostsPath := filepath.Join(parent, "ceph-dev.known_hosts")
	unrelatedPath := filepath.Join(parent, "unrelated.txt")
	for _, path := range []string{statePath, knownHostsPath, unrelatedPath} {
		if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := cleanupLocalState(statePath, knownHostsPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(unrelatedPath); err != nil {
		t.Fatalf("custom parent content was removed: %v", err)
	}
	for _, path := range []string{statePath, knownHostsPath} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("owned file %s still exists or stat failed unexpectedly: %v", path, err)
		}
	}
}
