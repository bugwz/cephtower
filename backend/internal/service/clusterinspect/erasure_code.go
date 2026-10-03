package clusterinspect

import (
	"context"
	"regexp"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

type ErasureCodeInfo struct {
	Manager    string    `json:"manager"`
	Plugins    []string  `json:"plugins"`
	Directory  string    `json:"directory"`
	ObservedAt time.Time `json:"observed_at"`
}

// Read the running manager's effective configuration, including defaults, rather
// than config get, which does not include daemon-local overrides.
func (s *Service) ErasureCodeInfo(ctx context.Context, clusterID uint64) (ErasureCodeInfo, error) {
	var mgr struct {
		ActiveName string `json:"active_name"`
	}
	if err := s.read(ctx, clusterID, "erasure_code.manager", []string{"mgr", "dump", "--format", "json"}, &mgr); err != nil {
		return ErasureCodeInfo{}, err
	}
	bad := func() (ErasureCodeInfo, error) {
		return ErasureCodeInfo{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned incomplete erasure code configuration"}
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]*$`).MatchString(mgr.ActiveName) {
		return bad()
	}
	manager := "mgr." + mgr.ActiveName
	var config []struct {
		Name  string  `json:"name"`
		Value *string `json:"value"`
	}
	if err := s.read(ctx, clusterID, "erasure_code.configuration", []string{"config", "show-with-defaults", manager, "--format", "json"}, &config); err != nil {
		return ErasureCodeInfo{}, err
	}
	values := map[string]string{}
	for _, row := range config {
		if row.Name != "osd_erasure_code_plugins" && row.Name != "erasure_code_dir" {
			continue
		}
		if _, duplicate := values[row.Name]; duplicate || row.Value == nil || strings.TrimSpace(*row.Value) == "" {
			return bad()
		}
		values[row.Name] = *row.Value
	}
	if len(values) != 2 {
		return bad()
	}
	plugins := []string{}
	seen := map[string]bool{}
	for _, plugin := range strings.Fields(values["osd_erasure_code_plugins"]) {
		if !seen[plugin] {
			plugins = append(plugins, plugin)
			seen[plugin] = true
		}
	}
	return ErasureCodeInfo{Manager: manager, Plugins: plugins, Directory: values["erasure_code_dir"], ObservedAt: time.Now().UTC()}, nil
}
