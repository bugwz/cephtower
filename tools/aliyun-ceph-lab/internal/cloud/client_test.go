package cloud

import (
	"testing"
	"time"

	"cephtower/tools/aliyun-ceph-lab/internal/config"
)

func TestNewRequiresCredentialsFromConfiguration(t *testing.T) {
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "environment-access-key-id")
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "environment-access-key-secret")
	if _, err := New(&config.Config{RegionID: "cn-test"}); err == nil {
		t.Fatal("New() accepted empty configuration credentials through environment fallback")
	}
}

func TestClientTokenIsStableAndBounded(t *testing.T) {
	t.Parallel()
	expiresAt := time.Date(2026, 7, 26, 8, 0, 0, 0, time.UTC)
	first := clientToken("test-lab", "node-1", expiresAt)
	second := clientToken("test-lab", "node-1", expiresAt)
	if first != second {
		t.Fatalf("clientToken() is not stable: %q != %q", first, second)
	}
	if len(first) > 64 {
		t.Fatalf("clientToken() length = %d, want <= 64", len(first))
	}
	if first == clientToken("test-lab", "node-2", expiresAt) {
		t.Fatal("clientToken() did not change for a different node")
	}
}

func TestValidateVSwitch(t *testing.T) {
	t.Parallel()
	valid := vSwitchInfo{ID: "vsw-test", VPCID: "vpc-test", ZoneID: "cn-test-a", Status: "Available", AvailableIP: 3}
	if err := validateVSwitch(valid, "cn-test-a", 3); err != nil {
		t.Fatalf("validateVSwitch() rejected valid switch: %v", err)
	}
	valid.AvailableIP = 2
	if err := validateVSwitch(valid, "cn-test-a", 3); err == nil {
		t.Fatal("validateVSwitch() allowed insufficient addresses")
	}
	if err := validateVSwitch(valid, "cn-test-a", 2); err != nil {
		t.Fatalf("validateVSwitch() did not adapt to a two-node configuration: %v", err)
	}
}

func TestCIDRContains(t *testing.T) {
	t.Parallel()
	if !cidrContains("172.31.0.0/16", "172.31.10.0/24") {
		t.Fatal("cidrContains() rejected a valid subnet")
	}
	if cidrContains("10.0.0.0/16", "172.31.10.0/24") {
		t.Fatal("cidrContains() accepted an unrelated subnet")
	}
}

func TestManagedNetworkResourceNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		prefix string
		want   string
	}{
		{prefix: vpcNamePrefix, want: "ceph-vpc-test-lab"},
		{prefix: vSwitchNamePrefix, want: "ceph-switch-test-lab"},
		{prefix: securityGroupNamePrefix, want: "ceph-security-group-test-lab"},
	}
	for _, test := range tests {
		if got := resourceName(test.prefix, "test-lab"); got != test.want {
			t.Fatalf("resourceName(%q) = %q, want %q", test.prefix, got, test.want)
		}
	}
}

func TestLabSecurityGroupPermissions(t *testing.T) {
	t.Parallel()
	permissions := labSecurityGroupPermissions("0.0.0.0/0")
	if len(permissions) != 6 {
		t.Fatalf("permission count = %d, want 6", len(permissions))
	}

	wantPorts := map[string]bool{
		"22/22": false, "8443/8443": false, "3300/3300": false,
		"6789/6789": false, "6800/7568": false, "36900/36900": false,
	}
	for _, permission := range permissions {
		port := stringValue(permission.PortRange)
		if _, ok := wantPorts[port]; !ok {
			t.Fatalf("unexpected public port range %q", port)
		}
		wantPorts[port] = true
		if stringValue(permission.IpProtocol) != "TCP" ||
			stringValue(permission.NicType) != "intranet" ||
			stringValue(permission.SourceCidrIp) != "0.0.0.0/0" {
			t.Fatalf("unexpected public permission for %s: %#v", port, permission)
		}
	}
	for port, found := range wantPorts {
		if !found {
			t.Fatalf("missing public port range %s", port)
		}
	}
}

func TestLabSecurityGroupPermissionsRestrictCephTowerToConfiguredSource(t *testing.T) {
	t.Parallel()
	permissions := labSecurityGroupPermissions("203.0.113.10/32")
	for _, permission := range permissions {
		if stringValue(permission.PortRange) != "36900/36900" {
			continue
		}
		if stringValue(permission.SourceCidrIp) != "203.0.113.10/32" {
			t.Fatalf("CephTower source CIDR = %q, want 203.0.113.10/32", stringValue(permission.SourceCidrIp))
		}
		return
	}
	t.Fatal("missing CephTower HTTP permission")
}
