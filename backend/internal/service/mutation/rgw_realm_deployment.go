package mutation

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func realmDeploymentReady(serviceRaw, daemonRaw []byte, expected map[string]any) (ready, valid bool) {
	if serviceSpecAbsent(serviceRaw) {
		return false, true
	}
	var services []map[string]any
	if json.Unmarshal(serviceRaw, &services) != nil || len(services) != 1 {
		return false, false
	}
	service := services[0]
	for key, value := range expected {
		if key == "spec" {
			// cephadm clears update_endpoints after publishing gateway URLs;
			// its bootstrap token is execution metadata, not a placement setting.
			stable := func(value any) map[string]any {
				spec, _ := value.(map[string]any)
				copy := map[string]any{}
				for k, v := range spec {
					if k != "update_endpoints" && k != "rgw_realm_token" {
						copy[k] = v
					}
				}
				return copy
			}
			if !reflect.DeepEqual(stable(service[key]), stable(value)) {
				return false, false
			}
			continue
		}
		if !reflect.DeepEqual(service[key], value) {
			return false, false
		}
	}
	if service["unmanaged"] == true {
		return false, false
	}
	var status struct {
		Size    *int      `json:"size"`
		Running *int      `json:"running"`
		Refresh time.Time `json:"last_refresh"`
	}
	raw, _ := json.Marshal(service["status"])
	if json.Unmarshal(raw, &status) != nil {
		return false, false
	}
	if status.Size == nil || status.Running == nil || status.Refresh.IsZero() || *status.Size <= 0 || *status.Running != *status.Size {
		return false, true
	}
	placement, _ := expected["placement"].(map[string]any)
	if count, exists := placement["count"]; exists && !reflect.DeepEqual(count, float64(*status.Size)) {
		return false, true
	}
	daemons, ok := setupDaemonSnapshot(daemonRaw, rawText(expected, "service_name"))
	if !ok {
		return false, false
	}
	if len(daemons) != *status.Size {
		return false, true
	}
	for _, daemon := range daemons {
		if daemon.Status == nil || *daemon.Status != 1 || daemon.Started.IsZero() || daemon.Refresh.IsZero() {
			return false, true
		}
	}
	return true, true
}

func waitRealmDeployment(ctx context.Context, interval time.Duration, check func() (bool, bool)) bool {
	for {
		if ctx.Err() != nil {
			return false
		}
		ready, valid := check()
		if !valid || ctx.Err() != nil {
			return false
		}
		if ready {
			return true
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return false
		case <-timer.C:
		}
	}
}

func (s *Service) verifyRealmDeployment(ctx context.Context, access executor.ClusterAccess, request Request, expected map[string]any) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	name := rawText(expected, "service_name")
	read := func(stage, command string) ([]byte, bool) {
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: executor.BinaryCeph, Args: []string{"orch", command, "--service-name", name, "--refresh", "--format", "json"}, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
		clear(result.Stderr)
		if err != nil || result.ExitCode != 0 {
			clear(result.Stdout)
			return nil, false
		}
		return result.Stdout, true
	}
	return waitRealmDeployment(ctx, time.Second, func() (bool, bool) {
		service, ok := read("deployment_service", "ls")
		defer clear(service)
		if !ok {
			return false, false
		}
		daemons, ok := read("deployment_daemons", "ps")
		defer clear(daemons)
		if !ok {
			return false, false
		}
		return realmDeploymentReady(service, daemons, expected)
	})
}
