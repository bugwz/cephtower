package clusterinspect

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sort"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

// RGWDaemon describes service-map registration, not a live health probe.
type RGWDaemon struct {
	ServiceMapID  string `json:"service_map_id"`
	ID            string `json:"id"`
	Hostname      string `json:"hostname"`
	Version       string `json:"version"`
	RealmName     string `json:"realm_name"`
	ZonegroupName string `json:"zonegroup_name"`
	ZonegroupID   string `json:"zonegroup_id"`
	ZoneName      string `json:"zone_name"`
}

func (s *Service) RGWDaemons(ctx context.Context, clusterID uint64) ([]RGWDaemon, error) {
	if clusterID == 0 {
		return nil, invalid("cluster_id is required")
	}
	access, err := s.clusters.Access(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: "rgw.daemons.read", Binary: executor.BinaryCeph, Args: []string{"service", "dump", "--format", "json"}, Timeout: 20 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	defer clear(result.Stdout)
	defer clear(result.Stderr)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil || result.ExitCode != 0 {
		return nil, &cephdomain.ActionError{Code: "ceph_command_failed", Message: "RGW service map could not be read"}
	}
	return decodeRGWDaemons(result.Stdout)
}

func decodeRGWDaemons(body []byte) ([]RGWDaemon, error) {
	fail := func() ([]RGWDaemon, error) {
		return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "RGW service map is incomplete or invalid"}
	}
	var root struct {
		Services map[string]json.RawMessage `json:"services"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	if decoder.Decode(&root) != nil || decoder.Decode(new(any)) != io.EOF || root.Services == nil {
		return fail()
	}
	rows := []RGWDaemon{}
	raw, exists := root.Services["rgw"]
	if !exists {
		return rows, nil
	}
	var service struct {
		Daemons map[string]json.RawMessage `json:"daemons"`
	}
	if json.Unmarshal(raw, &service) != nil || service.Daemons == nil {
		return fail()
	}
	for key, raw := range service.Daemons {
		if key == "summary" {
			var summary string
			if json.Unmarshal(raw, &summary) != nil {
				return fail()
			}
			continue
		}
		var daemon struct {
			Metadata map[string]string `json:"metadata"`
		}
		if key == "" || json.Unmarshal(raw, &daemon) != nil || daemon.Metadata == nil || daemon.Metadata["id"] == "" {
			return fail()
		}
		m := daemon.Metadata
		// Do not forward arbitrary metadata: frontend configs can contain credentials.
		rows = append(rows, RGWDaemon{ServiceMapID: key, ID: m["id"], Hostname: m["hostname"], Version: m["ceph_version"], RealmName: m["realm_name"], ZonegroupName: m["zonegroup_name"], ZonegroupID: m["zonegroup_id"], ZoneName: m["zone_name"]})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ID == rows[j].ID {
			return rows[i].ServiceMapID < rows[j].ServiceMapID
		}
		return rows[i].ID < rows[j].ID
	})
	return rows, nil
}
