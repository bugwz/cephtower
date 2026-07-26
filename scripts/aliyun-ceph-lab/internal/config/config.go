package config

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/netip"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	generatedSSHPasswordLength = 10
	defaultMaxRuntime          = 6 * time.Hour
	defaultWaitTimeout         = 15 * time.Minute
)

var safeName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{1,63}$`)

type Config struct {
	ClusterName                 string
	RegionID                    string
	ZoneID                      string
	AccessKeyID                 string
	AccessKeySecret             string
	SecurityToken               string
	VPCID                       string
	VSwitchID                   string
	SecurityGroupID             string
	ImageID                     string
	CPUOptions                  *CPUOptions
	HTTPTokens                  string
	SecurityEnhancementStrategy string
	SSHPassword                 string
	SSHUser                     string
	MaxRuntime                  string
	WaitTimeout                 string
	SystemDiskCategory          string
	SystemDiskPerformanceLevel  string
	SystemDiskSizeGiB           int32
	InternetChargeType          string
	InternetMaxBandwidthOutMbps int32
	Network                     Network
	InitScript                  string
	DeployScript                string
	Endpoint                    string
	VPCEndpoint                 string
	Nodes                       []Node
	configDir                   string
	maxRuntimeDuration          time.Duration
	waitTimeoutDuration         time.Duration
	sshPasswordGenerated        bool
}

type fileConfig struct {
	Credentials credentialConfig `yaml:"credentials"`
	Cluster     clusterConfig    `yaml:"cluster"`
	Network     Network          `yaml:"network"`
	ECS         ecsConfig        `yaml:"ecs"`
	SSH         sshConfig        `yaml:"ssh"`
	Lifecycle   lifecycleConfig  `yaml:"lifecycle"`
	Hooks       hooksConfig      `yaml:"hooks"`
	Nodes       []Node           `yaml:"nodes"`
}

type credentialConfig struct {
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	SecurityToken   string `yaml:"security_token,omitempty"`
}

type clusterConfig struct {
	Name        string `yaml:"name"`
	RegionID    string `yaml:"region_id"`
	ZoneID      string `yaml:"zone_id"`
	ECSEndpoint string `yaml:"ecs_endpoint,omitempty"`
	VPCEndpoint string `yaml:"vpc_endpoint,omitempty"`
}

type ecsConfig struct {
	ImageID                     string           `yaml:"image_id"`
	CPUOptions                  *CPUOptions      `yaml:"cpu_options,omitempty"`
	HTTPTokens                  string           `yaml:"http_tokens,omitempty"`
	SecurityEnhancementStrategy string           `yaml:"security_enhancement_strategy,omitempty"`
	SystemDisk                  systemDiskConfig `yaml:"system_disk"`
	Internet                    internetConfig   `yaml:"internet"`
}

type systemDiskConfig struct {
	Category         string `yaml:"category"`
	PerformanceLevel string `yaml:"performance_level,omitempty"`
	SizeGiB          int32  `yaml:"size_gib"`
}

type internetConfig struct {
	ChargeType          string `yaml:"charge_type"`
	MaxBandwidthOutMbps int32  `yaml:"max_bandwidth_out_mbps"`
}

type sshConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type lifecycleConfig struct {
	MaxRuntime  string `yaml:"max_runtime"`
	WaitTimeout string `yaml:"wait_timeout"`
}

type hooksConfig struct {
	InitScript   string `yaml:"init_script"`
	DeployScript string `yaml:"deploy_script"`
}

type Node struct {
	Name         string     `yaml:"name"`
	InstanceType string     `yaml:"instance_type"`
	DataDisks    []DataDisk `yaml:"data_disks,omitempty"`
}

type CPUOptions struct {
	CoreCount      int32 `yaml:"core_count"`
	ThreadsPerCore int32 `yaml:"threads_per_core"`
}

type DataDisk struct {
	Category         string `yaml:"category"`
	PerformanceLevel string `yaml:"performance_level,omitempty"`
	SizeGiB          int32  `yaml:"size_gib"`
}

type Network struct {
	VPCID                   string `yaml:"vpc_id,omitempty"`
	VSwitchID               string `yaml:"v_switch_id,omitempty"`
	SecurityGroupID         string `yaml:"security_group_id,omitempty"`
	AutoCreate              *bool  `yaml:"auto_create,omitempty"`
	ReuseManagedResources   *bool  `yaml:"reuse_managed_resources,omitempty"`
	VPCCIDR                 string `yaml:"vpc_cidr,omitempty"`
	VSwitchCIDR             string `yaml:"v_switch_cidr,omitempty"`
	SSHSourceCIDR           string `yaml:"ssh_source_cidr,omitempty"`
	CleanupCreatedResources *bool  `yaml:"cleanup_created_resources,omitempty"`
}

func Load(path string) (*Config, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".yaml" && ext != ".yml" {
		return nil, fmt.Errorf("configuration must be a YAML file with a .yaml or .yml extension: %s", path)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var file fileConfig
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	cfg := file.runtimeConfig()
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve config path: %w", err)
	}
	cfg.configDir = filepath.Dir(abs)
	if err := cfg.applyDefaultsAndValidate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (f fileConfig) runtimeConfig() Config {
	return Config{
		ClusterName: f.Cluster.Name,
		RegionID:    f.Cluster.RegionID, ZoneID: f.Cluster.ZoneID,
		AccessKeyID: f.Credentials.AccessKeyID, AccessKeySecret: f.Credentials.AccessKeySecret,
		SecurityToken: f.Credentials.SecurityToken,
		Endpoint:      f.Cluster.ECSEndpoint, VPCEndpoint: f.Cluster.VPCEndpoint,
		VPCID: f.Network.VPCID, VSwitchID: f.Network.VSwitchID,
		SecurityGroupID: f.Network.SecurityGroupID, Network: f.Network,
		ImageID: f.ECS.ImageID, CPUOptions: f.ECS.CPUOptions, HTTPTokens: f.ECS.HTTPTokens,
		SecurityEnhancementStrategy: f.ECS.SecurityEnhancementStrategy,
		SystemDiskCategory:          f.ECS.SystemDisk.Category,
		SystemDiskPerformanceLevel:  f.ECS.SystemDisk.PerformanceLevel,
		SystemDiskSizeGiB:           f.ECS.SystemDisk.SizeGiB,
		InternetChargeType:          f.ECS.Internet.ChargeType,
		InternetMaxBandwidthOutMbps: f.ECS.Internet.MaxBandwidthOutMbps,
		SSHUser:                     f.SSH.User, SSHPassword: f.SSH.Password,
		MaxRuntime: f.Lifecycle.MaxRuntime, WaitTimeout: f.Lifecycle.WaitTimeout,
		InitScript: f.Hooks.InitScript, DeployScript: f.Hooks.DeployScript,
		Nodes: f.Nodes,
	}
}

func (c *Config) applyDefaultsAndValidate() error {
	if c.SSHPassword == "" {
		password, err := generateSSHPassword()
		if err != nil {
			return fmt.Errorf("generate ssh.password: %w", err)
		}
		c.SSHPassword = password
		c.sshPasswordGenerated = true
	}
	if c.SSHUser == "" {
		c.SSHUser = "root"
	}
	if c.MaxRuntime == "" {
		c.MaxRuntime = defaultMaxRuntime.String()
	}
	if c.WaitTimeout == "" {
		c.WaitTimeout = defaultWaitTimeout.String()
	}
	if c.SystemDiskCategory == "" {
		c.SystemDiskCategory = "cloud_essd"
	}
	if c.SystemDiskSizeGiB == 0 {
		c.SystemDiskSizeGiB = 40
	}
	if c.InternetChargeType == "" {
		c.InternetChargeType = "PayByTraffic"
	}
	if c.Network.AutoCreate == nil {
		c.Network.AutoCreate = boolPointer(true)
	}
	if c.Network.ReuseManagedResources == nil {
		c.Network.ReuseManagedResources = boolPointer(true)
	}
	if c.Network.CleanupCreatedResources == nil {
		c.Network.CleanupCreatedResources = boolPointer(false)
	}
	if c.Network.VPCCIDR == "" {
		c.Network.VPCCIDR = "172.31.0.0/16"
	}
	if c.Network.VSwitchCIDR == "" {
		c.Network.VSwitchCIDR = "172.31.0.0/24"
	}
	if c.Network.SSHSourceCIDR == "" {
		c.Network.SSHSourceCIDR = "0.0.0.0/0"
	}
	if c.InitScript == "" {
		c.InitScript = "hooks/init-node.sh"
	}
	if c.DeployScript == "" {
		c.DeployScript = "hooks/deploy-ceph.sh"
	}

	var err error
	c.maxRuntimeDuration, err = time.ParseDuration(c.MaxRuntime)
	if err != nil {
		return fmt.Errorf("invalid lifecycle.max_runtime: %w", err)
	}
	c.waitTimeoutDuration, err = time.ParseDuration(c.WaitTimeout)
	if err != nil {
		return fmt.Errorf("invalid lifecycle.wait_timeout: %w", err)
	}
	if c.maxRuntimeDuration < 30*time.Minute || c.maxRuntimeDuration > 3*365*24*time.Hour {
		return errors.New("lifecycle.max_runtime must be between 30m and 26280h")
	}
	if c.waitTimeoutDuration <= 0 {
		return errors.New("lifecycle.wait_timeout must be positive")
	}
	if c.SystemDiskSizeGiB <= 0 {
		return errors.New("ecs.system_disk.size_gib must be positive")
	}
	if c.InternetMaxBandwidthOutMbps <= 0 {
		return errors.New("ecs.internet.max_bandwidth_out_mbps must be positive so the tool can initialize nodes over SSH")
	}
	if c.CPUOptions != nil && (c.CPUOptions.CoreCount <= 0 || c.CPUOptions.ThreadsPerCore <= 0) {
		return errors.New("ecs.cpu_options requires positive core_count and threads_per_core")
	}
	if err := validateSSHPassword(c.SSHPassword); err != nil {
		return fmt.Errorf("invalid ssh.password: %w", err)
	}
	vpcPrefix, err := netip.ParsePrefix(c.Network.VPCCIDR)
	if err != nil || !vpcPrefix.Addr().Is4() || vpcPrefix.Bits() < 16 || vpcPrefix.Bits() > 28 {
		return errors.New("network.vpc_cidr must be an IPv4 CIDR with a /16 to /28 mask")
	}
	vSwitchPrefix, err := netip.ParsePrefix(c.Network.VSwitchCIDR)
	if err != nil || !vSwitchPrefix.Addr().Is4() || !vpcPrefix.Contains(vSwitchPrefix.Addr()) ||
		vSwitchPrefix.Bits() < vpcPrefix.Bits() || vSwitchPrefix.Bits() > 29 {
		return errors.New("network.v_switch_cidr must be a /16 to /29 IPv4 subnet of network.vpc_cidr")
	}
	if c.Network.SSHSourceCIDR != "" {
		sshPrefix, parseErr := netip.ParsePrefix(c.Network.SSHSourceCIDR)
		if parseErr != nil || !sshPrefix.Addr().Is4() {
			return errors.New("network.ssh_source_cidr must be a valid IPv4 CIDR")
		}
	}
	if (c.VSwitchID == "" || c.SecurityGroupID == "") && !c.NetworkAutoCreate() && !c.NetworkReuseManagedResources() {
		return errors.New("network IDs are missing while automatic creation and managed-resource reuse are disabled")
	}

	required := map[string]string{
		"cluster.name": c.ClusterName, "cluster.region_id": c.RegionID, "cluster.zone_id": c.ZoneID,
		"ecs.image_id": c.ImageID,
	}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", field)
		}
	}
	if !safeName.MatchString(c.ClusterName) {
		return errors.New("cluster.name must start with a letter and contain only letters, digits, _ or -")
	}
	if c.SSHUser != "root" {
		return errors.New(`ssh.user must be "root" because the tool uses ECS root password login`)
	}
	if len(c.Nodes) == 0 {
		return errors.New("nodes must contain at least one node")
	}
	seen := make(map[string]struct{}, len(c.Nodes))
	for i, node := range c.Nodes {
		if !safeName.MatchString(node.Name) {
			return fmt.Errorf("nodes[%d].name is invalid", i)
		}
		if node.InstanceType == "" {
			return fmt.Errorf("nodes[%d].instance_type is required", i)
		}
		if _, exists := seen[node.Name]; exists {
			return fmt.Errorf("duplicate node name %q", node.Name)
		}
		seen[node.Name] = struct{}{}
		for j, disk := range node.DataDisks {
			if disk.Category == "" || disk.SizeGiB <= 0 {
				return fmt.Errorf("nodes[%d].data_disks[%d] requires category and positive size_gib", i, j)
			}
		}
	}
	return nil
}

func (c *Config) MaxRuntimeDuration() time.Duration  { return c.maxRuntimeDuration }
func (c *Config) WaitTimeoutDuration() time.Duration { return c.waitTimeoutDuration }
func (c *Config) NetworkAutoCreate() bool {
	return c.Network.AutoCreate != nil && *c.Network.AutoCreate
}
func (c *Config) NetworkReuseManagedResources() bool {
	return c.Network.ReuseManagedResources != nil && *c.Network.ReuseManagedResources
}
func (c *Config) CleanupCreatedNetworkResources() bool {
	return c.Network.CleanupCreatedResources != nil && *c.Network.CleanupCreatedResources
}

func (c *Config) SSHPasswordWasGenerated() bool { return c.sshPasswordGenerated }

func (c *Config) ValidateCloudCredentials() error {
	if strings.TrimSpace(c.AccessKeyID) == "" || strings.TrimSpace(c.AccessKeySecret) == "" {
		return errors.New("credentials.access_key_id and credentials.access_key_secret must be configured")
	}
	return nil
}

func (c *Config) DefaultStatePath() string {
	return filepath.Join(c.configDir, ".state", c.ClusterName+".json")
}

func boolPointer(value bool) *bool { return &value }

func validateSSHPassword(password string) error {
	if len(password) < 8 || len(password) > 30 {
		return errors.New("must contain 8 to 30 ASCII characters")
	}
	classes := 0
	hasUpper, hasLower, hasDigit, hasSpecial := false, false, false, false
	const allowedSpecial = "()`~!@#$%^&*-_+=|{}[]:;\\\"'<>,.?/"
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune(allowedSpecial, char):
			hasSpecial = true
		default:
			return fmt.Errorf("contains unsupported character %q", char)
		}
	}
	for _, present := range []bool{hasUpper, hasLower, hasDigit, hasSpecial} {
		if present {
			classes++
		}
	}
	if classes < 3 {
		return errors.New("must contain at least three of uppercase letters, lowercase letters, digits, and special characters")
	}
	return nil
}

func generateSSHPassword() (string, error) {
	const (
		uppercase = "ABCDEFGHJKLMNPQRSTUVWXYZ"
		lowercase = "abcdefghijkmnopqrstuvwxyz"
		digits    = "23456789"
		specials  = "!@#$%^&*-_+="
	)
	all := uppercase + lowercase + digits + specials
	password := make([]byte, 0, generatedSSHPasswordLength)
	for _, characters := range []string{uppercase, lowercase, digits, specials} {
		character, err := randomCharacter(characters)
		if err != nil {
			return "", err
		}
		password = append(password, character)
	}
	for len(password) < generatedSSHPasswordLength {
		character, err := randomCharacter(all)
		if err != nil {
			return "", err
		}
		password = append(password, character)
	}
	for index := len(password) - 1; index > 0; index-- {
		selection, err := rand.Int(rand.Reader, big.NewInt(int64(index+1)))
		if err != nil {
			return "", fmt.Errorf("shuffle password: %w", err)
		}
		other := int(selection.Int64())
		password[index], password[other] = password[other], password[index]
	}
	return string(password), nil
}

func randomCharacter(characters string) (byte, error) {
	selection, err := rand.Int(rand.Reader, big.NewInt(int64(len(characters))))
	if err != nil {
		return 0, fmt.Errorf("read cryptographic randomness: %w", err)
	}
	return characters[selection.Int64()], nil
}

func (c *Config) ResolvePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(c.configDir, path)
}
