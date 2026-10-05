package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func targetSettings(updated bool) map[string]any {
	v := map[string]any{"region": "old-region", "host_style": "path", "target_path": "old/path", "target_storage_class": "OLD", "multipart_sync_threshold": 32, "multipart_min_part_size": 8}
	if updated {
		v = map[string]any{"region": "", "host_style": "virtual", "target_path": "new,{path}", "target_storage_class": "true", "multipart_sync_threshold": 0, "multipart_min_part_size": 16}
	}
	return v
}
func targetParams() map[string]any {
	p := groupClassParams("other")
	delete(p, "confirm_create")
	p["confirm_target"] = true
	p["tier_type"] = "cloud-s3"
	p["expected_target"] = targetSettings(false)
	p["target"] = targetSettings(true)
	return p
}
func targetFixture() *placementExecutor {
	r := cloudRestoreFixture()
	for _, stage := range []string{"before", "recheck", "after"} {
		group := periodDocument([]byte(r.bodies["before"]))
		s3 := group["placement_targets"].([]any)[0].(map[string]any)["tier_targets"].([]any)[0].(map[string]any)["val"].(map[string]any)["s3"].(map[string]any)
		for k, v := range targetSettings(stage == "after") {
			s3[k] = v
		}
		raw, _ := json.Marshal(group)
		r.bodies[stage] = string(raw)
	}
	return r
}
func TestCloudTargetChain(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	req := Request{ClusterID: cluster, Action: "rgw_zonegroup.cloud_target", Parameters: targetParams()}
	r := targetFixture()
	s.executor = r
	result, err := s.Execute(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	d := result.Details.(map[string]any)
	if d["target_configuration_verified"] != true || d["period_published"] != true || d["data_migrated"] != false {
		t.Fatal(d)
	}
	if len(r.calls) != 9 || strings.Join(r.calls[3].Args, " ") != `zonegroup placement modify --zonegroup-id g --placement-id p --storage-class COLD --format json --tier-config region="",host_style="virtual",target_path="new\u002c\u007bpath\u007d",target_storage_class="true",multipart_sync_threshold=0,multipart_min_part_size=16` {
		t.Fatal(r.calls)
	}
	for _, stage := range []string{"before", "realm_before", "recheck", "add", "after", "period.pre_check", "period.commit", "period.realm_post_check", "period.period_post_check"} {
		for _, mode := range []string{"error", "exit"} {
			r = targetFixture()
			s.executor = r
			if mode == "error" {
				r.fail = stage
			} else {
				r.codeStage = stage
			}
			_, err = s.Execute(context.Background(), req)
			var ae *cephdomain.ActionError
			if !errors.As(err, &ae) || ae.Retryable || strings.Contains(err.Error(), "private") {
				t.Fatal(stage, err)
			}
			if r.calls[len(r.calls)-1].ID != "rgw_zonegroup.cloud_target."+stage {
				t.Fatal("continued after failure")
			}
		}
	}
	for _, scenario := range []string{"old_changed", "wrong_tier", "missing_field", "drift", "secret_changed"} {
		r = targetFixture()
		s.executor = r
		switch scenario {
		case "old_changed":
			r.bodies["before"] = strings.ReplaceAll(r.bodies["before"], "old-region", "concurrent")
		case "wrong_tier":
			r.bodies["before"] = strings.ReplaceAll(r.bodies["before"], "cloud-s3", "unknown")
		case "missing_field":
			r.bodies["before"] = strings.ReplaceAll(r.bodies["before"], `"multipart_sync_threshold":32,`, "")
		case "drift":
			r.bodies["recheck"] = strings.ReplaceAll(r.bodies["recheck"], "restricted", "changed")
		case "secret_changed":
			r.bodies["after"] = strings.ReplaceAll(r.bodies["after"], "private-secret", "changed")
		}
		if _, err = s.Execute(context.Background(), req); err == nil {
			t.Fatal(scenario)
		}
		if scenario != "secret_changed" {
			for _, c := range r.calls {
				if c.Mutating {
					t.Fatal("unsafe write", scenario)
				}
			}
		}
	}
}
func TestCloudTargetGlacierAndValidation(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	p := targetParams()
	p["tier_type"] = "cloud-s3-glacier"
	r := targetFixture()
	for _, stage := range []string{"before", "recheck", "after"} {
		r.bodies[stage] = strings.ReplaceAll(r.bodies[stage], `"tier_type":"cloud-s3"`, `"s3-glacier":{"glacier_restore_days":7,"glacier_restore_tier_type":"Standard"},"tier_type":"cloud-s3-glacier"`)
	}
	s.executor = r
	if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.cloud_target", Parameters: p}); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"target", "expected_target"} {
		for _, change := range []map[string]any{{"host_style": "other"}, {"region": nil}, {"multipart_sync_threshold": -1}, {"multipart_min_part_size": 1.5}, {"multipart_min_part_size": 9007199254740992.0}, {"target_path": "\n"}, {"target_path": strings.Repeat("x", 4097)}, {"endpoint": "other"}} {
			p = targetParams()
			v := p[key].(map[string]any)
			for k, x := range change {
				v[k] = x
			}
			if _, err := buildCloudTarget(p); err == nil {
				t.Fatal(key, change)
			}
		}
	}
	p = targetParams()
	p["target"] = p["expected_target"]
	if _, err := buildCloudTarget(p); err == nil {
		t.Fatal("no-op accepted")
	}
	for key, value := range map[string]any{"realm_id": "", "tier_type": "local", "confirm_target": false, "storage_class": "STANDARD"} {
		p = targetParams()
		p[key] = value
		if _, err := buildCloudTarget(p); err == nil {
			t.Fatal(key)
		}
	}
}
