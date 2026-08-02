package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const Version = 1

type State struct {
	Version     int       `json:"version"`
	ClusterName string    `json:"cluster_name"`
	RegionID    string    `json:"region_id"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Network     Network   `json:"network,omitempty"`
	Nodes       []Node    `json:"nodes"`
	Ceph        *Ceph     `json:"ceph,omitempty"`
}

type Network struct {
	VPCID                string `json:"vpc_id,omitempty"`
	VSwitchID            string `json:"v_switch_id,omitempty"`
	SecurityGroupID      string `json:"security_group_id,omitempty"`
	CreatedVPC           bool   `json:"created_vpc,omitempty"`
	CreatedVSwitch       bool   `json:"created_v_switch,omitempty"`
	CreatedSecurityGroup bool   `json:"created_security_group,omitempty"`
}

type Node struct {
	Name       string `json:"name"`
	InstanceID string `json:"instance_id"`
	Status     string `json:"status,omitempty"`
	PublicIP   string `json:"public_ip,omitempty"`
	PrivateIP  string `json:"private_ip,omitempty"`
	SSH        SSH    `json:"ssh"`
}

type SSH struct {
	Host              string `json:"host,omitempty"`
	Port              int    `json:"port"`
	User              string `json:"user"`
	Password          string `json:"password"`
	PasswordGenerated bool   `json:"password_generated"`
	LogPath           string `json:"log_path,omitempty"`
}

type Ceph struct {
	ClusterName            string          `json:"cluster_name,omitempty"`
	FSID                   string          `json:"fsid,omitempty"`
	ClientAdmin            CephClientAdmin `json:"client_admin"`
	Dashboard              CephDashboard   `json:"dashboard"`
	Monitors               CephMonitors    `json:"monitors"`
	CephTowerClusterCreate CephTowerCreate `json:"cephtower_cluster_create"`
}

type CephClientAdmin struct {
	Username string `json:"username"`
	Key      string `json:"key"`
	Keyring  string `json:"keyring,omitempty"`
}

type CephDashboard struct {
	URL      string `json:"url,omitempty"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type CephMonitors struct {
	MonitorAddresses string                `json:"monitor_addresses"`
	V1Addresses      string                `json:"v1_addresses,omitempty"`
	V2Addresses      string                `json:"v2_addresses,omitempty"`
	Endpoints        []CephMonitorEndpoint `json:"endpoints,omitempty"`
}

type CephMonitorEndpoint struct {
	Name     string `json:"name,omitempty"`
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Host     string `json:"host,omitempty"`
	Port     uint16 `json:"port,omitempty"`
	Nonce    uint64 `json:"nonce,omitempty"`
}

type CephTowerCreate struct {
	Name             string `json:"name"`
	MonitorAddresses string `json:"monitor_addresses"`
	ClientUsername   string `json:"client_username"`
	ClientKey        string `json:"client_key"`
}

func Load(path string) (*State, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}
	var value State
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("decode state: %w", err)
	}
	if value.Version != Version {
		return nil, fmt.Errorf("unsupported state version %d", value.Version)
	}
	if value.ClusterName == "" || value.RegionID == "" {
		return nil, errors.New("state is missing cluster_name or region_id")
	}
	return &value, nil
}

func Save(path string, value *State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	raw = append(raw, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return fmt.Errorf("write temporary state: %w", err)
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		return fmt.Errorf("secure temporary state: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("replace state: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("secure state file: %w", err)
	}
	return nil
}
