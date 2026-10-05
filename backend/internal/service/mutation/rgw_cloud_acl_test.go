package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func aclParams() map[string]any {
	p := groupClassParams("other")
	delete(p, "confirm_create")
	p["tier_type"] = "cloud-s3"
	p["confirm_acl"] = true
	p["confirm_clear"] = false
	p["expected_acls"] = []cloudACL{{"keep", "old", "id"}, {"remove", "removed", "email"}}
	p["acls"] = []cloudACL{{"keep", "new", "uri"}, {"true", "", "id"}}
	return p
}
func aclFixture() *placementExecutor {
	r := cloudRestoreFixture()
	before := strings.Replace(r.bodies["before"], `"acl_mappings":[]`, `"acl_mappings":[{"key":"keep","val":{"source_id":"keep","dest_id":"old","type":"id"}},{"key":"remove","val":{"source_id":"remove","dest_id":"removed","type":"email"}}]`, 1)
	r.bodies["before"], r.bodies["recheck"] = before, before
	r.bodies["after"] = strings.Replace(before, `{"key":"remove","val":{"source_id":"remove","dest_id":"removed","type":"email"}}`, `{"key":"true","val":{"source_id":"true","dest_id":"","type":"id"}}`, 1)
	r.bodies["after"] = strings.Replace(r.bodies["after"], `"dest_id":"old","type":"id"`, `"dest_id":"new","type":"uri"`, 1)
	return r
}
func TestCloudACLChain(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	p := aclParams()
	req := Request{ClusterID: cluster, Action: "rgw_zonegroup.cloud_acl", Parameters: p}
	r := aclFixture()
	s.executor = r
	result, err := s.Execute(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	d := result.Details.(map[string]any)
	if d["acl_configuration_verified"] != true || d["period_published"] != true || d["existing_object_acls_rewritten"] != false {
		t.Fatal(d)
	}
	args := r.calls[3].Args
	if len(r.calls) != 9 || !reflect.DeepEqual(args, []string{"zonegroup", "placement", "modify", "--zonegroup-id", "g", "--placement-id", "p", "--storage-class", "COLD", "--format", "json", "--tier-config", `acls[0].source_id="keep",acls[0].dest_id="new",acls[0].type="uri",acls[1].source_id="true",acls[1].dest_id="",acls[1].type="id"`, "--tier-config-rm", `acls[0].source_id="remove"`}) {
		t.Fatal(args)
	}
	for _, stage := range []string{"before", "realm_before", "recheck", "add", "after", "period.pre_check", "period.commit", "period.realm_post_check", "period.period_post_check"} {
		for _, mode := range []string{"error", "exit"} {
			r = aclFixture()
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
			if r.calls[len(r.calls)-1].ID != "rgw_zonegroup.cloud_acl."+stage {
				t.Fatal("continued", stage)
			}
		}
	}
	for _, scenario := range []string{"old_changed", "wrong_type", "wrong_key", "malformed_key", "duplicate", "recheck_changed", "secret_changed"} {
		r = aclFixture()
		s.executor = r
		switch scenario {
		case "old_changed":
			r.bodies["before"] = strings.ReplaceAll(r.bodies["before"], `"dest_id":"old"`, `"dest_id":"concurrent"`)
		case "wrong_type":
			r.bodies["before"] = strings.ReplaceAll(r.bodies["before"], "cloud-s3", "unknown")
		case "wrong_key":
			r.bodies["before"] = strings.ReplaceAll(r.bodies["before"], `"key":"keep"`, `"key":"wrong"`)
		case "malformed_key":
			r.bodies["before"] = strings.ReplaceAll(strings.ReplaceAll(r.bodies["before"], `"key":"keep"`, `"key":{}`), `"source_id":"keep"`, `"source_id":{}`)
		case "duplicate":
			r.bodies["before"] = strings.ReplaceAll(r.bodies["before"], `"remove"`, `"keep"`)
		case "recheck_changed":
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
func TestCloudACLExplicitClearAndGlacier(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	p := aclParams()
	p["acls"] = []cloudACL{}
	p["confirm_clear"] = true
	p["tier_type"] = "cloud-s3-glacier"
	r := aclFixture()
	for _, stage := range []string{"before", "recheck"} {
		r.bodies[stage] = strings.ReplaceAll(r.bodies[stage], `"tier_type":"cloud-s3"`, `"s3-glacier":{"glacier_restore_days":7,"glacier_restore_tier_type":"Standard"},"tier_type":"cloud-s3-glacier"`)
	}
	after := periodDocument([]byte(r.bodies["before"]))
	tier := after["placement_targets"].([]any)[0].(map[string]any)["tier_targets"].([]any)[0].(map[string]any)["val"].(map[string]any)
	tier["s3"].(map[string]any)["acl_mappings"] = []any{}
	raw, _ := json.Marshal(after)
	r.bodies["after"] = string(raw)
	s.executor = r
	if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.cloud_acl", Parameters: p}); err != nil {
		t.Fatal(err)
	}
	args := r.calls[3].Args
	if len(args) != 13 || args[11] != "--tier-config-rm" || args[12] != `acls[0].source_id="keep",acls[1].source_id="remove"` {
		t.Fatal(args)
	}
}
func TestCloudACLValidationAndLiteralStrings(t *testing.T) {
	for key, values := range map[string][]any{"realm_id": {""}, "storage_class": {"STANDARD"}, "tier_type": {"local"}, "confirm_acl": {false, nil}, "confirm_clear": {nil}, "acls": {nil, []cloudACL{{"", "x", "id"}}, []cloudACL{{"a", "x", "ID"}}, []cloudACL{{"a", "x", "id"}, {"a", "y", "uri"}}, []cloudACL{{"a,b", "x", "id"}}, []cloudACL{{"a", "{x}", "id"}}, []cloudACL{{"a", "\n", "id"}}}} {
		for _, v := range values {
			p := aclParams()
			p[key] = v
			if _, err := buildCloudACL(p); err == nil {
				t.Fatal(key, v)
			}
		}
	}
	p := aclParams()
	p["acls"] = p["expected_acls"]
	if _, err := buildCloudACL(p); err == nil {
		t.Fatal("no-op accepted")
	}
	p = aclParams()
	p["acls"] = []cloudACL{}
	if _, err := buildCloudACL(p); err == nil {
		t.Fatal("unconfirmed clear")
	}
	p = aclParams()
	p["expected_acls"] = []cloudACL{}
	p["acls"] = []cloudACL{{`001`, `quote"slash\=`, "id"}}
	spec, err := buildCloudACL(p)
	if err != nil {
		t.Fatal(err)
	}
	if spec.args[len(spec.args)-1] != `acls[0].source_id="001",acls[0].dest_id="quote\"slash\\=",acls[0].type="id"` {
		t.Fatal(spec.args)
	}
	for _, entries := range [][]cloudACL{{{strings.Repeat("x", 4097), "d", "id"}}, make([]cloudACL, 257)} {
		p = aclParams()
		p["acls"] = entries
		if _, err := buildCloudACL(p); err == nil {
			t.Fatal("oversized ACL accepted")
		}
	}
	large := []cloudACL{}
	for i := 0; i < 20; i++ {
		large = append(large, cloudACL{strings.Repeat("x", i+1), strings.Repeat("d", 4096), "id"})
	}
	p = aclParams()
	p["acls"] = large
	if _, err := buildCloudACL(p); err == nil {
		t.Fatal("oversized command accepted")
	}
}
