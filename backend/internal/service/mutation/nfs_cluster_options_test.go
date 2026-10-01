package mutation

import (
	"reflect"
	"testing"
)

func TestNFSClusterIngressOptions(t *testing.T) {
	for _, tc := range []struct {
		p    map[string]any
		want []string
	}{
		{map[string]any{"nfs_port": 65535}, []string{"--port=65535"}},
		{map[string]any{"ingress": true, "virtual_ip": "192.0.2.10/24", "nfs_port": 2049, "ingress_mode": "haproxy-protocol"}, []string{"--ingress", "--virtual-ip=192.0.2.10/24", "--ingress-mode=haproxy-protocol", "--port=2049"}},
		{map[string]any{"ingress": true, "virtual_ip": "2001:db8::1/64", "nfs_port": 58535, "ingress_mode": "keepalive-only"}, []string{"--ingress", "--virtual-ip=2001:db8::1/64", "--ingress-mode=keepalive-only", "--port=58535"}},
	} {
		tc.p["name"] = "nfs-a"
		spec, err := build(Request{Action: "nfs_cluster.create"}, tc.p)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(spec.args, append([]string{"nfs", "cluster", "create", "nfs-a"}, tc.want...)) {
			t.Fatal(spec.args)
		}
	}
	for _, p := range []map[string]any{
		{"ingress": "true"}, {"ingress": true}, {"virtual_ip": "192.0.2.1"}, {"ingress_mode": "default"},
		{"ingress": true, "virtual_ip": "bad"}, {"ingress": true, "virtual_ip": "192.0.2.1/99"},
		{"ingress": true, "virtual_ip": "192.0.2.1", "ingress_mode": "unknown"},
		{"ingress": true, "virtual_ip": "192.0.2.1", "nfs_port": 55536},
		{"ingress": true, "virtual_ip": "192.0.2.1", "ingress_mode": "keepalive-only", "nfs_port": 58536},
		{"nfs_port": 0}, {"nfs_port": 65536}, {"nfs_port": 1.5}, {"nfs_port": "2049"}, {"nfs_port": nil},
	} {
		if _, err := nfsClusterIngressArgs(p); err == nil {
			t.Fatalf("invalid options accepted: %v", p)
		}
	}
}
