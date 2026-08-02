package ceph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/connection"
	"cephtower/backend/internal/integration/ceph/executor"
)

type ClusterAccess = executor.ClusterAccess
type Capability struct {
	Name      string         `json:"name"`
	Supported bool           `json:"supported"`
	Reason    string         `json:"reason,omitempty"`
	Version   string         `json:"version,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}
type ProbeResult struct {
	FSID         string         `json:"fsid"`
	Version      string         `json:"version"`
	Capabilities []Capability   `json:"capabilities"`
	Status       map[string]any `json:"status"`
}
type ClusterProvider interface {
	Probe(context.Context, ClusterAccess) (ProbeResult, error)
}

type NativeProvider struct{ Executor executor.Executor }

func (p *NativeProvider) Probe(ctx context.Context, access ClusterAccess) (ProbeResult, error) {
	if _, err := connection.ParseMonitorAddresses(access.MonitorAddresses); err != nil {
		return ProbeResult{}, err
	}
	fsid, err := p.run(ctx, access, "cluster.fsid", 20*time.Second, "fsid")
	if err != nil {
		return ProbeResult{}, err
	}
	versions, err := p.run(ctx, access, "cluster.versions", 30*time.Second, "versions", "--format", "json")
	if err != nil {
		return ProbeResult{}, err
	}
	status, err := p.run(ctx, access, "cluster.status", 30*time.Second, "status", "--format", "json")
	if err != nil {
		return ProbeResult{}, err
	}
	result := ProbeResult{FSID: strings.TrimSpace(string(fsid)), Capabilities: []Capability{{Name: "ceph_cli", Supported: true}, {Name: "dashboard_independent", Supported: true}}}
	if result.FSID == "" {
		return ProbeResult{}, fmt.Errorf("ceph fsid returned an empty value")
	}
	if version := cephVersionFromVersions(versions); version != "" {
		result.Version = version
	}
	if err := decodeJSON(status, &result.Status); err != nil {
		return ProbeResult{}, fmt.Errorf("parse ceph status: %w", err)
	}
	optional := []struct {
		name   string
		binary executor.Binary
		args   []string
	}{{"orchestrator", executor.BinaryCeph, []string{"orch", "status", "--format", "json"}}, {"mgr_module", executor.BinaryCeph, []string{"mgr", "module", "ls", "--format", "json"}}, {"cephfs_volume", executor.BinaryCeph, []string{"fs", "volume", "ls", "--format", "json"}}, {"nfs", executor.BinaryCeph, []string{"nfs", "cluster", "ls", "--format", "json"}}, {"smb", executor.BinaryCeph, []string{"smb", "cluster", "ls", "--format", "json"}}, {"rbd", executor.BinaryRBD, []string{"--version"}}, {"rgw_admin", executor.BinaryRGWAdmin, []string{"--version"}}, {"cephfs_data_access", executor.BinaryCephFSShell, []string{"--version"}}}
	for _, probe := range optional {
		_, err := p.runBinary(ctx, access, probe.binary, "capability."+probe.name, 30*time.Second, probe.args...)
		capability := Capability{Name: probe.name, Supported: err == nil}
		if err != nil {
			capability.Reason = "probe_failed"
		}
		result.Capabilities = append(result.Capabilities, capability)
	}
	return result, nil
}

func (p *NativeProvider) run(ctx context.Context, access ClusterAccess, id string, timeout time.Duration, args ...string) ([]byte, error) {
	return p.runBinary(ctx, access, executor.BinaryCeph, id, timeout, args...)
}
func (p *NativeProvider) runBinary(ctx context.Context, access ClusterAccess, binary executor.Binary, id string, timeout time.Duration, args ...string) ([]byte, error) {
	if p.Executor == nil {
		return nil, fmt.Errorf("ceph executor is unavailable")
	}
	result, err := p.Executor.Run(ctx, access, executor.CommandSpec{ID: id, Binary: binary, Args: args, Timeout: timeout, MaxOutput: executor.DefaultMaxOutput})
	if err != nil {
		return nil, err
	}
	return result.Stdout, nil
}
func decodeJSON(data []byte, out any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder.Decode(out)
}

func cephVersionFromVersions(data []byte) string {
	var wire map[string]map[string]json.Number
	if err := decodeJSON(data, &wire); err != nil {
		return ""
	}
	for _, component := range []string{"mon", "mgr", "osd", "mds"} {
		if version := firstVersion(wire[component]); version != "" {
			return version
		}
	}
	components := make([]string, 0, len(wire))
	for component := range wire {
		components = append(components, component)
	}
	sort.Strings(components)
	for _, component := range components {
		if version := firstVersion(wire[component]); version != "" {
			return version
		}
	}
	return ""
}

func firstVersion(versions map[string]json.Number) string {
	values := make([]string, 0, len(versions))
	for version := range versions {
		if normalized := cephdomain.NormalizeVersion(version); normalized != "" {
			values = append(values, normalized)
		}
	}
	sort.Strings(values)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
