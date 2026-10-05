package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func cloudCreateParams(tier, def string) map[string]any {
	p := connectionParams()
	delete(p, "confirm_connection")
	p["confirm_create"] = true
	p["expected_default_placement"] = def
	p["tier_type"] = tier
	p["target"] = targetSettings(true)
	p["acls"] = []any{map[string]any{"source_id": "true", "dest_id": "", "type": "email"}}
	p["retain_head_object"], p["allow_read_through"] = false, true
	p["read_through_restore_days"], p["restore_storage_class"] = 0, "STANDARD"
	if tier == "cloud-s3-glacier" {
		p["glacier_restore_days"], p["glacier_restore_tier_type"] = 0, "Expedited"
	}
	return p
}

func cloudCreateFixture(tier, def string) *placementExecutor {
	r := placementFixture("r")
	f := groupClassFixture(def)
	for _, stage := range []string{"before", "recheck", "after"} {
		group := periodDocument([]byte(f.bodies[stage]))
		target := group["placement_targets"].([]any)[0].(map[string]any)
		target["tier_targets"] = []any{}
		if stage == "after" {
			s3 := map[string]any{"endpoint": "https://new.example:9443/s3", "access_key": "new-access", "secret": `literal,{secret}"=true`, "region": "", "host_style": "virtual", "target_path": "new,{path}", "target_storage_class": "true", "multipart_sync_threshold": 0, "multipart_min_part_size": 16, "acl_mappings": []any{map[string]any{"key": "true", "val": map[string]any{"source_id": "true", "dest_id": "", "type": "email"}}}}
			val := map[string]any{"tier_type": tier, "storage_class": "COLD", "retain_head_object": false, "allow_read_through": true, "read_through_restore_days": 0, "restore_storage_class": "STANDARD", "s3": s3}
			if tier == "cloud-s3-glacier" {
				val["s3-glacier"] = map[string]any{"glacier_restore_days": 0, "glacier_restore_tier_type": "Expedited"}
			}
			target["tier_targets"] = []any{map[string]any{"key": "COLD", "val": val}}
		}
		raw, _ := json.Marshal(group)
		r.bodies[stage] = string(raw)
	}
	return r
}

func TestCloudClassCreationChain(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, tier := range []string{"cloud-s3", "cloud-s3-glacier"} {
		for _, def := range []string{"other", ""} {
			r := cloudCreateFixture(tier, def)
			s.executor = r
			result, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.cloud_create", Parameters: cloudCreateParams(tier, def)})
			if err != nil {
				t.Fatal(tier, def, err)
			}
			d := result.Details.(map[string]any)
			if d["cloud_class_created"] != true || d["period_published"] != true || d["remote_connection_tested"] != false || d["data_migrated"] != false || len(r.calls) != 9 {
				t.Fatal("wrong creation result")
			}
			raw, _ := json.Marshal(result)
			if strings.Contains(string(raw), "new-access") || strings.Contains(string(raw), "secret") {
				t.Fatal("credential result leak")
			}
			cmd := r.calls[3]
			want := `zonegroup placement add --zonegroup-id g --placement-id p --storage-class COLD --format json --tier-config endpoint="https://new.example:9443/s3",access_key="new-access",secret="literal\u002c\u007bsecret\u007d\"=true",retain_head_object=false,allow_read_through=true,read_through_restore_days=0,restore_storage_class="STANDARD"`
			if tier == "cloud-s3-glacier" {
				want += `,glacier_restore_days=0,glacier_restore_tier_type=Expedited`
			}
			want += `,region="",host_style="virtual",target_path="new\u002c\u007bpath\u007d",target_storage_class="true",multipart_sync_threshold=0,multipart_min_part_size=16,acls[0].source_id="true",acls[0].dest_id="",acls[0].type="email" --tier-type ` + tier
			if strings.Join(cmd.Args, " ") != want {
				t.Fatal("incorrect create command")
			}
			if _, ok := cmd.SensitiveArgs[len(cmd.Args)-3]; !ok || len(cmd.SensitiveArgs) != 1 {
				t.Fatal("credential argument index changed")
			}
			for i, call := range r.calls {
				if i != 3 && len(call.SensitiveArgs) > 0 {
					t.Fatal("sensitive index leaked to another stage")
				}
			}
		}
	}
	req := Request{ClusterID: cluster, Action: "rgw_zonegroup.cloud_create", Parameters: cloudCreateParams("cloud-s3", "other")}
	for _, stage := range []string{"before", "realm_before", "recheck", "add", "after", "period.pre_check", "period.commit", "period.realm_post_check", "period.period_post_check"} {
		for _, mode := range []string{"error", "exit"} {
			r := cloudCreateFixture("cloud-s3", "other")
			s.executor = r
			if mode == "error" {
				r.fail = stage
			} else {
				r.codeStage = stage
			}
			_, err := s.Execute(context.Background(), req)
			var ae *cephdomain.ActionError
			if !errors.As(err, &ae) || ae.Retryable || strings.Contains(err.Error(), "private") || r.calls[len(r.calls)-1].ID != req.Action+"."+stage {
				t.Fatal(stage, err)
			}
		}
	}
}

func TestCloudClassCreationConflictGuards(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, scenario := range []string{"existing_class", "orphan_tier", "missing_target", "missing_tiers", "duplicate_tier", "restore_cloud", "restore_missing", "drift", "collateral_change"} {
		r := cloudCreateFixture("cloud-s3", "other")
		s.executor = r
		p := cloudCreateParams("cloud-s3", "other")
		switch scenario {
		case "existing_class":
			r.bodies["before"] = strings.Replace(r.bodies["before"], `"storage_classes":["STANDARD"]`, `"storage_classes":["COLD","STANDARD"]`, 1)
		case "orphan_tier":
			r.bodies["before"] = strings.Replace(r.bodies["before"], `"tier_targets":[]`, `"tier_targets":[{"key":"COLD","val":{"storage_class":"COLD"}}]`, 1)
		case "missing_target":
			p["placement_id"] = "missing"
		case "missing_tiers":
			r.bodies["before"] = strings.Replace(r.bodies["before"], `"tier_targets":[]`, `"tier_targets":null`, 1)
		case "duplicate_tier":
			r.bodies["before"] = strings.Replace(r.bodies["before"], `"tier_targets":[]`, `"tier_targets":[{"key":"X","val":{"storage_class":"X"}},{"key":"X","val":{"storage_class":"X"}}]`, 1)
		case "restore_cloud":
			r.bodies["before"] = strings.Replace(r.bodies["before"], `"tier_targets":[]`, `"tier_targets":[{"key":"STANDARD","val":{"storage_class":"STANDARD"}}]`, 1)
		case "restore_missing":
			p["restore_storage_class"] = "MISSING"
		case "drift":
			r.bodies["recheck"] = strings.ReplaceAll(r.bodies["recheck"], "restricted", "changed")
		case "collateral_change":
			r.bodies["after"] = strings.ReplaceAll(r.bodies["after"], "restricted", "changed")
		}
		if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.cloud_create", Parameters: p}); err == nil {
			t.Fatal(scenario)
		}
		if scenario != "collateral_change" {
			for _, call := range r.calls {
				if call.Mutating {
					t.Fatal("unsafe write", scenario)
				}
			}
		}
	}
}

func TestCloudClassCreationExplicitFields(t *testing.T) {
	for _, key := range []string{"confirm_create", "credentials_saved", "endpoint", "access_key", "secret", "target", "acls", "retain_head_object", "allow_read_through", "read_through_restore_days", "restore_storage_class", "tier_type"} {
		p := cloudCreateParams("cloud-s3", "other")
		delete(p, key)
		if _, err := buildCloudCreate(p); err == nil {
			t.Fatal("missing field accepted", key)
		}
	}
	for key, value := range map[string]any{"storage_class": "STANDARD", "realm_id": "", "endpoint": "https://u:p@host", "secret": "[REDACTED]", "read_through_restore_days": -1, "glacier_restore_days": 1, "confirm_create": false, "credentials_saved": false, "acls": nil} {
		p := cloudCreateParams("cloud-s3", "other")
		p[key] = value
		if _, err := buildCloudCreate(p); err == nil {
			t.Fatal("invalid field accepted", key)
		}
	}
	p := cloudCreateParams("cloud-s3", "other")
	p["acls"] = []any{}
	if _, err := buildCloudCreate(p); err != nil {
		t.Fatal("empty ACL rejected", err)
	}
	entries := []any{}
	for i := 0; i < 20; i++ {
		entries = append(entries, map[string]any{"source_id": strings.Repeat("x", i+1), "dest_id": strings.Repeat("\\", 4096), "type": "id"})
	}
	p["acls"] = entries
	if _, err := buildCloudCreate(p); err == nil {
		t.Fatal("oversized command accepted")
	}
}

func TestCloudClassCreationPreservesOtherTiersAndEmptyACL(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	p := cloudCreateParams("cloud-s3", "other")
	p["acls"] = []any{}
	r := cloudCreateFixture("cloud-s3", "other")
	for _, stage := range []string{"before", "recheck", "after"} {
		group := periodDocument([]byte(r.bodies[stage]))
		target := group["placement_targets"].([]any)[0].(map[string]any)
		target["storage_classes"] = append(target["storage_classes"].([]any), "ZOLD")
		tiers := target["tier_targets"].([]any)
		if stage == "after" {
			tiers[0].(map[string]any)["val"].(map[string]any)["s3"].(map[string]any)["acl_mappings"] = []any{}
		}
		target["tier_targets"] = append(tiers, map[string]any{"key": "ZOLD", "val": map[string]any{"storage_class": "ZOLD", "tier_type": "cloud-s3-glacier", "s3": map[string]any{"secret": "preserved-secret"}, "future": true}})
		raw, _ := json.Marshal(group)
		r.bodies[stage] = string(raw)
	}
	s.executor = r
	if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.cloud_create", Parameters: p}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(r.calls[3].Args, " "), "acls[") {
		t.Fatal("empty ACL emitted spurious assignment")
	}
}
