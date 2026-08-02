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

func TestBuildCephMonitorsProducesMonitorAddresses(t *testing.T) {
	t.Parallel()
	dump := monDumpWire{Mons: []monWire{
		{Name: "node-1", PublicAddrs: struct {
			Addrvec []monAddrvecWire `json:"addrvec"`
		}{Addrvec: []monAddrvecWire{
			{Type: "v2", Addr: "172.31.0.10:3300", Nonce: 0},
			{Type: "v1", Addr: "172.31.0.10:6789", Nonce: 0},
		}}},
		{Name: "node-2", PublicAddrs: struct {
			Addrvec []monAddrvecWire `json:"addrvec"`
		}{Addrvec: []monAddrvecWire{
			{Type: "v2", Addr: "172.31.0.11:3300", Nonce: 0},
			{Type: "v1", Addr: "172.31.0.11:6789", Nonce: 0},
		}}},
	}}
	monitors, err := buildCephMonitors(dump)
	if err != nil {
		t.Fatal(err)
	}
	want := "[v2:172.31.0.10:3300/0,v1:172.31.0.10:6789/0],[v2:172.31.0.11:3300/0,v1:172.31.0.11:6789/0]"
	if monitors.MonitorAddresses != want {
		t.Fatalf("MonitorAddresses = %q, want %q", monitors.MonitorAddresses, want)
	}
	if monitors.V1Addresses != "v1:172.31.0.10:6789/0,v1:172.31.0.11:6789/0" {
		t.Fatalf("V1Addresses = %q", monitors.V1Addresses)
	}
	if monitors.V2Addresses != "v2:172.31.0.10:3300/0,v2:172.31.0.11:3300/0" {
		t.Fatalf("V2Addresses = %q", monitors.V2Addresses)
	}
	if len(monitors.Endpoints) != 4 || monitors.Endpoints[0].Host != "172.31.0.10" ||
		monitors.Endpoints[0].Port != 3300 {
		t.Fatalf("Endpoints = %#v", monitors.Endpoints)
	}
}

func TestGenerateDashboardPasswordContainsRequiredClasses(t *testing.T) {
	t.Parallel()
	password, err := generateDashboardPassword()
	if err != nil {
		t.Fatal(err)
	}
	if len(password) != 20 {
		t.Fatalf("password length = %d, want 20", len(password))
	}
	if !strings.ContainsAny(password, "abcdefghijkmnopqrstuvwxyz") ||
		!strings.ContainsAny(password, "ABCDEFGHJKLMNPQRSTUVWXYZ") ||
		!strings.ContainsAny(password, "23456789") ||
		!strings.ContainsAny(password, "#%+.-_") {
		t.Fatalf("password does not contain every class: %q", password)
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

func TestStateHasCloudResources(t *testing.T) {
	t.Parallel()
	if stateHasCloudResources(&state.State{Network: state.Network{VPCID: "vpc-reused"}}) {
		t.Fatal("reused network IDs should not force state retention")
	}
	if !stateHasCloudResources(&state.State{Network: state.Network{CreatedVPC: true}}) {
		t.Fatal("created network resources should force state retention")
	}
	if !stateHasCloudResources(&state.State{Nodes: []state.Node{{InstanceID: "i-test"}}}) {
		t.Fatal("created ECS nodes should force state retention")
	}
}
