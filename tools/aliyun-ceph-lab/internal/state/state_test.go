package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoad(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "nested", "state.json")
	want := &State{
		Version: Version, ClusterName: "test-lab", RegionID: "cn-test",
		CreatedAt: time.Date(2026, 7, 26, 1, 2, 3, 0, time.UTC),
		ExpiresAt: time.Date(2026, 7, 26, 7, 2, 3, 0, time.UTC),
		Network:   Network{VPCID: "vpc-test", VSwitchID: "vsw-test", CreatedVPC: true},
		Ceph: &Ceph{
			ClusterName: "ceph",
			FSID:        "00000000-0000-0000-0000-000000000001",
			ClientAdmin: CephClientAdmin{Username: "client.admin", Key: "admin-key", Keyring: "[client.admin]\n\tkey = admin-key\n"},
			Dashboard:   CephDashboard{URL: "https://203.0.113.10:8443/", Username: "admin", Password: "dashboard-password"},
			Monitors: CephMonitors{
				MonitorAddresses: "[v2:172.31.0.10:3300/0,v1:172.31.0.10:6789/0]",
				V1Addresses:      "v1:172.31.0.10:6789/0",
				V2Addresses:      "v2:172.31.0.10:3300/0",
				Endpoints: []CephMonitorEndpoint{{
					Name: "node-1", Protocol: "v2", Address: "172.31.0.10:3300", Host: "172.31.0.10", Port: 3300,
				}},
			},
			CephTowerClusterCreate: CephTowerCreate{
				Name: "test-lab", MonitorAddresses: "[v2:172.31.0.10:3300/0,v1:172.31.0.10:6789/0]",
				ClientUsername: "client.admin", ClientKey: "admin-key",
			},
		},
		Nodes: []Node{{
			Name: "node-1", InstanceID: "i-test", Status: "Running",
			SSH: SSH{
				Host: "203.0.113.10", Port: 22, User: "root", Password: "test-password",
				PasswordGenerated: true, LogPath: "/tmp/log/203.0.113.10.log",
			},
		}},
	}
	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.ClusterName != want.ClusterName || len(got.Nodes) != 1 || got.Nodes[0].InstanceID != "i-test" ||
		got.Network.VPCID != "vpc-test" || !got.Network.CreatedVPC ||
		got.Nodes[0].SSH.Host != "203.0.113.10" || got.Nodes[0].SSH.Port != 22 ||
		got.Nodes[0].SSH.User != "root" || got.Nodes[0].SSH.Password != "test-password" ||
		!got.Nodes[0].SSH.PasswordGenerated || got.Nodes[0].SSH.LogPath == "" ||
		got.Ceph == nil || got.Ceph.ClientAdmin.Key != "admin-key" ||
		got.Ceph.Monitors.MonitorAddresses == "" ||
		got.Ceph.CephTowerClusterCreate.ClientUsername != "client.admin" {
		t.Fatalf("Load() = %#v", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("state file permissions = %o, want 600", got)
	}
}
