package mutation

import (
	"context"
	"errors"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func localClassDeleteParams(realm string) map[string]any {
	p := groupClassDeleteParams(realm)
	p["zone_id"] = "z"
	p["zone_name"] = "zone"
	return p
}
func localClassDeleteFixture(realm string) *placementExecutor {
	r := groupClassDeleteFixture(realm)
	before := strings.Replace(r.bodies["before"], `,"tier_targets":[{"key":"COLD","val":{"tier_type":"cloud-s3","s3":{"secret":"private"}}}]`, "", 1)
	r.bodies["before"], r.bodies["recheck"], r.bodies["group_after_zone"] = before, before, before
	zone := strings.Replace(placementFixture(realm).bodies["zone_before"], `"key":"default"`, `"key":"p"`, 1)
	after := strings.Replace(zone, `,"COLD":{"data_pool":"data:old","compression_type":"none"}`, "", 1)
	for _, stage := range []string{"zone_before", "zone_recheck"} {
		r.bodies[stage] = zone
	}
	for _, stage := range []string{"zone_remove", "zone_after", "zone_final"} {
		r.bodies[stage] = after
	}
	return r
}
func TestLocalStorageClassDeletionChain(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, realm := range []string{"", "r"} {
		r := localClassDeleteFixture(realm)
		s.executor = r
		req := Request{ClusterID: cluster, Action: "rgw_zonegroup.storage_class_delete_local", Parameters: localClassDeleteParams(realm)}
		result, err := s.Execute(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		d := result.Details.(map[string]any)
		if d["zone_mappings_removed"] != true || d["objects_removed"] != false || d["period_published"] != (realm != "") {
			t.Fatal(d)
		}
		writes := []string{}
		for _, call := range r.calls {
			if call.Mutating {
				writes = append(writes, strings.Join(call.Args, " "))
			}
			for _, mode := range []string{"exit", "error"} {
				broken := localClassDeleteFixture(realm)
				stage := strings.TrimPrefix(call.ID, "rgw_zonegroup.storage_class_delete_local.")
				if mode == "exit" {
					broken.codeStage = stage
				} else {
					broken.fail = stage
				}
				s.executor = broken
				_, err := s.Execute(context.Background(), req)
				var ae *cephdomain.ActionError
				if !errors.As(err, &ae) || ae.Retryable || strings.Contains(err.Error(), "private") {
					t.Fatal(err)
				}
				if broken.calls[len(broken.calls)-1].ID != call.ID {
					t.Fatal("continued after failure")
				}
			}
		}
		count := 2
		if realm != "" {
			count = 3
		}
		if len(writes) != count || writes[0] != "zone placement rm --zone-id z --placement-id p --storage-class COLD --format json" || writes[1] != "zonegroup placement rm --zonegroup-id g --placement-id p --storage-class COLD --format json" {
			t.Fatal(writes)
		}
	}
}
func TestLocalStorageClassDeletionGuards(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, scenario := range []string{"zone_identity", "zone_realm", "membership", "tier", "zone_class_missing", "zone_concurrent", "zone_write_drift", "zone_read_drift", "group_after_zone", "zone_final"} {
		r := localClassDeleteFixture("r")
		switch scenario {
		case "zone_identity":
			r.bodies["zone_before"] = strings.Replace(r.bodies["zone_before"], `"name":"zone"`, `"name":"wrong"`, 1)
		case "zone_realm":
			r.bodies["zone_before"] = strings.Replace(r.bodies["zone_before"], `"realm_id":"r"`, `"realm_id":"wrong"`, 1)
		case "membership":
			r.bodies["before"] = strings.Replace(r.bodies["before"], `"id":"z"`, `"id":"other"`, 1)
		case "tier":
			r.bodies["before"] = groupClassDeleteFixture("r").bodies["before"]
		case "zone_class_missing":
			r.bodies["zone_before"] = r.bodies["zone_after"]
		case "zone_concurrent":
			r.bodies["zone_recheck"] = r.bodies["zone_after"]
		case "zone_write_drift":
			r.bodies["zone_remove"] = strings.Replace(r.bodies["zone_remove"], "index:old", "wrong", 1)
		case "zone_read_drift":
			r.bodies["zone_after"] = strings.Replace(r.bodies["zone_after"], "index:old", "wrong", 1)
		case "group_after_zone":
			r.bodies["group_after_zone"] = strings.Replace(r.bodies["group_after_zone"], "keep", "wrong", 1)
		case "zone_final":
			r.bodies["zone_final"] = strings.Replace(r.bodies["zone_final"], "index:old", "wrong", 1)
		}
		s.executor = r
		_, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.storage_class_delete_local", Parameters: localClassDeleteParams("r")})
		if err == nil {
			t.Fatal(scenario)
		}
		for _, call := range r.calls {
			if strings.Contains(call.ID, "period") {
				t.Fatal("published after validation failure")
			}
			if scenario != "zone_final" && strings.HasSuffix(call.ID, ".remove") {
				t.Fatal("removed group after validation failure")
			}
		}
	}
	for _, key := range []string{"zone_id", "zone_name", "storage_class"} {
		p := localClassDeleteParams("")
		p[key] = ""
		if key == "storage_class" {
			p[key] = "STANDARD"
		}
		if _, err := buildLocalStorageClassDelete(p); err == nil {
			t.Fatal("invalid delete accepted")
		}
	}
}
