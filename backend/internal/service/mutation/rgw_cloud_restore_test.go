package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"errors"
	"strings"
	"testing"
)

func cloudRestoreParams() map[string]any {
	p := groupClassParams("other")
	delete(p, "confirm_create")
	p["confirm_restore"] = true
	p["tier_type"] = "cloud-s3"
	p["retain_head_object"] = false
	p["allow_read_through"] = true
	p["read_through_restore_days"] = 0
	p["restore_storage_class"] = "STANDARD"
	return p
}
func cloudRestoreFixture() *placementExecutor {
	r := placementFixture("r")
	before := `{"id":"g","name":"group","realm_id":"r","default_placement":"other","placement_targets":[{"name":"p","tags":["restricted"],"storage_classes":["COLD","STANDARD"],"tier_targets":[{"key":"COLD","val":{"tier_type":"cloud-s3","storage_class":"COLD","retain_head_object":true,"allow_read_through":false,"read_through_restore_days":1,"restore_storage_class":"STANDARD","s3":{"secret":"private-secret","endpoint":"https://cloud.example","acl_mappings":[]}}}]}]}`
	r.bodies["before"], r.bodies["recheck"] = before, before
	r.bodies["after"] = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(before, `"retain_head_object":true`, `"retain_head_object":false`), `"allow_read_through":false`, `"allow_read_through":true`), `"read_through_restore_days":1`, `"read_through_restore_days":0`)
	return r
}
func TestCloudRestoreChain(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	req := Request{ClusterID: cluster, Action: "rgw_zonegroup.cloud_restore", Parameters: cloudRestoreParams()}
	r := cloudRestoreFixture()
	s.executor = r
	result, err := s.Execute(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	d := result.Details.(map[string]any)
	if d["restore_configuration_verified"] != true || d["period_published"] != true || d["objects_restored"] != false {
		t.Fatal(d)
	}
	if len(r.calls) != 9 || strings.Join(r.calls[3].Args, " ") != "zonegroup placement modify --zonegroup-id g --placement-id p --storage-class COLD --format json --tier-config retain_head_object=false,allow_read_through=true,read_through_restore_days=0,restore_storage_class=STANDARD" {
		t.Fatal(r.calls)
	}
	for _, stage := range []string{"before", "realm_before", "recheck", "add", "after", "period.pre_check", "period.commit", "period.realm_post_check", "period.period_post_check"} {
		for _, mode := range []string{"error", "exit"} {
			r = cloudRestoreFixture()
			s.executor = r
			if mode == "error" {
				r.fail = stage
			} else {
				r.codeStage = stage
			}
			_, err = s.Execute(context.Background(), req)
			var failure *cephdomain.ActionError
			if !errors.As(err, &failure) || failure.Retryable || strings.Contains(err.Error(), "private") {
				t.Fatal(stage, err)
			}
			if r.calls[len(r.calls)-1].ID != "rgw_zonegroup.cloud_restore."+stage {
				t.Fatal("continued after failure", stage)
			}
		}
	}
	for _, scenario := range []string{"missing_tier", "wrong_type", "restore_is_cloud", "drift", "credentials_changed"} {
		r = cloudRestoreFixture()
		s.executor = r
		switch scenario {
		case "missing_tier":
			r.bodies["before"] = strings.ReplaceAll(r.bodies["before"], `"key":"COLD"`, `"key":"OTHER"`)
		case "wrong_type":
			r.bodies["before"] = strings.ReplaceAll(r.bodies["before"], "cloud-s3", "unknown")
		case "restore_is_cloud":
			r.bodies["before"] = strings.ReplaceAll(r.bodies["before"], `"key":"COLD"`, `"key":"STANDARD"`)
		case "drift":
			r.bodies["recheck"] = strings.ReplaceAll(r.bodies["recheck"], "restricted", "changed")
		case "credentials_changed":
			r.bodies["after"] = strings.ReplaceAll(r.bodies["after"], "private-secret", "changed")
		}
		if _, err = s.Execute(context.Background(), req); err == nil {
			t.Fatal(scenario)
		}
		if scenario != "credentials_changed" {
			for _, c := range r.calls {
				if c.Mutating {
					t.Fatal("unsafe write", scenario)
				}
			}
		}
	}
}
func TestCloudRestoreValidation(t *testing.T) {
	for key, values := range map[string][]any{"realm_id": {"", nil}, "storage_class": {"STANDARD"}, "tier_type": {"local"}, "retain_head_object": {"false", nil}, "allow_read_through": {"true", nil}, "read_through_restore_days": {-1, 1.5, 9007199254740992.0, "1", nil}, "restore_storage_class": {"", "a,b", "a=b", "a/b"}, "confirm_restore": {false, nil}} {
		for _, v := range values {
			p := cloudRestoreParams()
			p[key] = v
			if _, err := buildCloudRestore(p); err == nil {
				t.Fatal(key, v)
			}
		}
	}
	p := cloudRestoreParams()
	p["tier_type"] = "cloud-s3-glacier"
	p["glacier_restore_days"], p["glacier_restore_tier_type"] = 0, "Expedited"
	r := cloudRestoreFixture()
	group := periodDocument([]byte(strings.ReplaceAll(r.bodies["before"], `"tier_type":"cloud-s3"`, `"s3-glacier":{"glacier_restore_days":7,"glacier_restore_tier_type":"Standard"},"tier_type":"cloud-s3-glacier"`)))
	if _, ok := cloudRestoreExpected(group, p); !ok {
		t.Fatal("Glacier rejected")
	}
}

func TestCloudRestoreUpdatesGlacierAndInitializesDefault(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	p := cloudRestoreParams()
	p["tier_type"], p["expected_default_placement"] = "cloud-s3-glacier", ""
	p["glacier_restore_days"], p["glacier_restore_tier_type"] = 0, "Expedited"
	r := cloudRestoreFixture()
	for _, stage := range []string{"before", "recheck", "after"} {
		glacier := `"s3-glacier":{"glacier_restore_days":7,"glacier_restore_tier_type":"Standard","future":"preserved"},"tier_type":"cloud-s3-glacier"`
		if stage == "after" {
			glacier = strings.ReplaceAll(strings.ReplaceAll(glacier, `:7`, `:0`), `"Standard"`, `"Expedited"`)
		}
		r.bodies[stage] = strings.ReplaceAll(r.bodies[stage], `"tier_type":"cloud-s3"`, glacier)
		value := ""
		if stage == "after" {
			value = "p"
		}
		r.bodies[stage] = strings.ReplaceAll(r.bodies[stage], `"default_placement":"other"`, `"default_placement":"`+value+`"`)
	}
	s.executor = r
	if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.cloud_restore", Parameters: p}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(r.calls[3].Args[len(r.calls[3].Args)-1], ",glacier_restore_days=0,glacier_restore_tier_type=Expedited") {
		t.Fatal(r.calls[3])
	}
	if days, ok := cloudRestoreDays(9007199254740991); !ok || days != "9007199254740991" {
		t.Fatal("exact day boundary rejected")
	}
}

func TestGlacierRestoreValidation(t *testing.T) {
	for _, level := range []any{"Bulk", "standard", "Expedited,secret=x", "", nil} {
		p := cloudRestoreParams()
		p["tier_type"] = "cloud-s3-glacier"
		p["glacier_restore_days"] = 1
		p["glacier_restore_tier_type"] = level
		if _, err := buildCloudRestore(p); err == nil {
			t.Fatal("invalid level accepted", level)
		}
	}
	for _, days := range []any{nil, -1, 1.5, "1", 9007199254740992.0} {
		p := cloudRestoreParams()
		p["tier_type"] = "cloud-s3-glacier"
		p["glacier_restore_days"] = days
		p["glacier_restore_tier_type"] = "Standard"
		if _, err := buildCloudRestore(p); err == nil {
			t.Fatal("invalid days accepted", days)
		}
	}
	for _, key := range []string{"glacier_restore_days", "glacier_restore_tier_type"} {
		p := cloudRestoreParams()
		p[key] = nil
		if _, err := buildCloudRestore(p); err == nil {
			t.Fatal("S3 accepted Glacier field")
		}
	}
}
