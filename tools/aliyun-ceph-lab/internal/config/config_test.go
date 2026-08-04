package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAllowsSingleNode(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	raw := `cluster:
  name: test-lab
  region_id: cn-test
  zone_id: cn-test-a
network:
  v_switch_id: vsw-test
  security_group_id: sg-test
ecs:
  image_id: img-test
  internet:
    max_bandwidth_out_mbps: 5
nodes:
  - name: node-1
    instance_type: ecs.test
`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() rejected a single-node configuration: %v", err)
	}
	if len(cfg.Nodes) != 1 || cfg.Nodes[0].Name != "node-1" {
		t.Fatalf("unexpected single-node configuration: %#v", cfg.Nodes)
	}
}

func TestLoadRejectsEmptyNodeList(t *testing.T) {
	t.Parallel()
	cfg := Config{
		ClusterName: "test-lab", RegionID: "cn-test", ZoneID: "cn-test-a",
		VSwitchID: "vsw-test", SecurityGroupID: "sg-test", ImageID: "img-test",
		SSHPassword: "CephTower#123", InternetMaxBandwidthOutMbps: 5,
	}
	if err := cfg.applyDefaultsAndValidate(); err == nil {
		t.Fatal("applyDefaultsAndValidate() accepted an empty node list")
	}
}

func TestDefaultMatchesSingaporeLaunchAdvisorConfiguration(t *testing.T) {
	t.Parallel()
	cfg, err := Load(filepath.Join("..", "..", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RegionID != "ap-southeast-1" || cfg.ZoneID != "ap-southeast-1b" {
		t.Fatalf("unexpected location: %s/%s", cfg.RegionID, cfg.ZoneID)
	}
	if cfg.CPUOptions != nil {
		t.Fatalf("unexpected CPU options: %#v", cfg.CPUOptions)
	}
	if len(cfg.Nodes) != 3 || cfg.Nodes[0].InstanceType != "ecs.e-c1m2.xlarge" || len(cfg.Nodes[0].DataDisks) != 2 {
		t.Fatalf("unexpected node defaults: %#v", cfg.Nodes)
	}
	if cfg.Nodes[0].DataDisks[0].PerformanceLevel != "PL0" || cfg.InternetMaxBandwidthOutMbps != 100 {
		t.Fatalf("unexpected disk/network defaults: %#v / %d", cfg.Nodes[0].DataDisks, cfg.InternetMaxBandwidthOutMbps)
	}
	if !cfg.SSHPasswordWasGenerated() || validateSSHPassword(cfg.SSHPassword) != nil {
		t.Fatalf("unexpected generated SSH password: %q", cfg.SSHPassword)
	}
	if cfg.Network.AccessSourceCIDR != "0.0.0.0/0" {
		t.Fatalf("unexpected default security group source CIDR: %q", cfg.Network.AccessSourceCIDR)
	}
	if cfg.AccessKeyID != "" || cfg.AccessKeySecret != "" || cfg.SecurityToken != "" {
		t.Fatal("the default configuration must not contain real cloud credentials")
	}
}

func TestValidateSSHPassword(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{name: "configured", password: "CephTower#123", valid: true},
		{name: "three classes", password: "CephTower123", valid: true},
		{name: "too short", password: "Aa1#", valid: false},
		{name: "only two classes", password: "cephtower123", valid: false},
		{name: "unsupported character", password: "CephTower密码123", valid: false},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := validateSSHPassword(test.password)
			if test.valid && err != nil {
				t.Fatalf("validateSSHPassword() rejected valid password: %v", err)
			}
			if !test.valid && err == nil {
				t.Fatal("validateSSHPassword() accepted invalid password")
			}
		})
	}
}

func TestValidateCloudCredentials(t *testing.T) {
	t.Parallel()
	cfg := Config{}
	if err := cfg.ValidateCloudCredentials(); err == nil {
		t.Fatal("ValidateCloudCredentials() accepted empty credentials")
	}
	cfg.AccessKeyID = "test-access-key-id"
	cfg.AccessKeySecret = "test-access-key-secret"
	if err := cfg.ValidateCloudCredentials(); err != nil {
		t.Fatalf("ValidateCloudCredentials() rejected configured credentials: %v", err)
	}
}

func TestRuntimeConfigMapsCloudCredentials(t *testing.T) {
	t.Parallel()
	cfg := (fileConfig{Credentials: credentialConfig{
		AccessKeyID: "test-access-key-id", AccessKeySecret: "test-access-key-secret",
		SecurityToken: "test-security-token",
	}}).runtimeConfig()
	if cfg.AccessKeyID != "test-access-key-id" ||
		cfg.AccessKeySecret != "test-access-key-secret" ||
		cfg.SecurityToken != "test-security-token" {
		t.Fatalf("unexpected mapped cloud credentials: %#v", cfg)
	}
}

func TestGenerateSSHPassword(t *testing.T) {
	t.Parallel()
	seen := make(map[string]struct{})
	for range 100 {
		password, err := generateSSHPassword()
		if err != nil {
			t.Fatal(err)
		}
		if len(password) != generatedSSHPasswordLength {
			t.Fatalf("generated password length = %d, want %d", len(password), generatedSSHPasswordLength)
		}
		if err := validateSSHPassword(password); err != nil {
			t.Fatalf("generated invalid password %q: %v", password, err)
		}
		if strings.Contains(password, "&") {
			t.Fatalf("generated password contains unsupported JSON-sensitive character: %q", password)
		}
		if _, exists := seen[password]; exists {
			t.Fatalf("generated duplicate password %q", password)
		}
		seen[password] = struct{}{}
	}
}

func TestConfiguredSSHPasswordIsPreserved(t *testing.T) {
	t.Parallel()
	cfg := Config{
		ClusterName: "test-lab", RegionID: "cn-test", ZoneID: "cn-test-a",
		VSwitchID: "vsw-test", SecurityGroupID: "sg-test", ImageID: "img-test",
		SSHPassword: "CephTower#123", InternetMaxBandwidthOutMbps: 5,
		Nodes: []Node{
			{Name: "node-1", InstanceType: "ecs.test"},
			{Name: "node-2", InstanceType: "ecs.test"},
			{Name: "node-3", InstanceType: "ecs.test"},
		},
	}
	if err := cfg.applyDefaultsAndValidate(); err != nil {
		t.Fatal(err)
	}
	if cfg.SSHPassword != "CephTower#123" || cfg.SSHPasswordWasGenerated() {
		t.Fatalf("configured SSH password was changed: %q", cfg.SSHPassword)
	}
	if cfg.SSHUser != "root" {
		t.Fatalf("default SSH user = %q, want root", cfg.SSHUser)
	}
}

func TestSSHUserMustBeRoot(t *testing.T) {
	t.Parallel()
	cfg := Config{
		ClusterName: "test-lab", RegionID: "cn-test", ZoneID: "cn-test-a",
		VSwitchID: "vsw-test", SecurityGroupID: "sg-test", ImageID: "img-test",
		SSHUser: "ecs-user", SSHPassword: "CephTower#123", InternetMaxBandwidthOutMbps: 5,
		Nodes: []Node{
			{Name: "node-1", InstanceType: "ecs.test"},
			{Name: "node-2", InstanceType: "ecs.test"},
			{Name: "node-3", InstanceType: "ecs.test"},
		},
	}
	if err := cfg.applyDefaultsAndValidate(); err == nil {
		t.Fatal("applyDefaultsAndValidate() accepted a non-root SSH user")
	}
}

func TestLoadRejectsJSONExtension(t *testing.T) {
	t.Parallel()
	if _, err := Load(filepath.Join(t.TempDir(), "config.json")); err == nil {
		t.Fatal("Load() accepted a JSON configuration path")
	}
}

func TestLoadRejectsUnknownYAMLField(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("unknown_field: value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load() accepted an unknown YAML field")
	}
}

func TestLoadRejectsRemovedImageOptions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	raw := `cluster:
  name: test-lab
  region_id: cn-test
  zone_id: cn-test-a
network:
  v_switch_id: vsw-test
  security_group_id: sg-test
ecs:
  image_id: img-test
  image_options:
    login_as_non_root: false
  internet:
    max_bandwidth_out_mbps: 5
nodes:
  - name: node-1
    instance_type: ecs.test
  - name: node-2
    instance_type: ecs.test
  - name: node-3
    instance_type: ecs.test
`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load() accepted the removed ecs.image_options field")
	}
}

func TestLoadAllowsAutomaticNetworkCreation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	raw := `cluster:
  name: test-lab
  region_id: cn-test
  zone_id: cn-test-a
network:
  auto_create: true
  access_source_cidr: 203.0.113.10/32
ecs:
  image_id: img-test
  internet:
    max_bandwidth_out_mbps: 5
nodes:
  - name: node-1
    instance_type: ecs.test
  - name: node-2
    instance_type: ecs.test
  - name: node-3
    instance_type: ecs.test
`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.NetworkAutoCreate() || cfg.VSwitchID != "" || cfg.SecurityGroupID != "" {
		t.Fatalf("unexpected automatic network config: %#v", cfg.Network)
	}
	if cfg.Network.AccessSourceCIDR != "203.0.113.10/32" {
		t.Fatalf("configured security group source CIDR was changed: %q", cfg.Network.AccessSourceCIDR)
	}
}

func TestLoadDefaultsManagedSecurityGroupSourceCIDR(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	raw := `cluster:
  name: test-lab
  region_id: cn-test
  zone_id: cn-test-a
ecs:
  image_id: img-test
  internet:
    max_bandwidth_out_mbps: 5
nodes:
  - name: node-1
    instance_type: ecs.test
  - name: node-2
    instance_type: ecs.test
  - name: node-3
    instance_type: ecs.test
`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.NetworkReuseManagedResources() || cfg.SecurityGroupID != "" || cfg.Network.AccessSourceCIDR != "0.0.0.0/0" {
		t.Fatalf("unexpected managed security group config: %#v", cfg.Network)
	}
}
