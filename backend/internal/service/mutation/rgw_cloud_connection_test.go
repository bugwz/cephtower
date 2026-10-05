package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func connectionParams() map[string]any {
	p := groupClassParams("other")
	delete(p, "confirm_create")
	p["confirm_connection"], p["credentials_saved"] = true, true
	p["tier_type"], p["endpoint"] = "cloud-s3", "https://new.example:9443/s3"
	p["access_key"], p["secret"] = "new-access", `literal,{secret}"=true`
	return p
}

func connectionFixture() *placementExecutor {
	r := cloudRestoreFixture()
	for _, stage := range []string{"before", "recheck", "after"} {
		group := periodDocument([]byte(r.bodies["before"]))
		s3 := group["placement_targets"].([]any)[0].(map[string]any)["tier_targets"].([]any)[0].(map[string]any)["val"].(map[string]any)["s3"].(map[string]any)
		s3["access_key"] = "old-access"
		if stage == "after" {
			for _, key := range []string{"endpoint", "access_key", "secret"} {
				s3[key] = connectionParams()[key]
			}
		}
		raw, _ := json.Marshal(group)
		r.bodies[stage] = string(raw)
	}
	return r
}

func TestCloudConnectionChain(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	req := Request{ClusterID: cluster, Action: "rgw_zonegroup.cloud_connection", Parameters: connectionParams()}
	for _, tier := range []string{"cloud-s3", "cloud-s3-glacier"} {
		r := connectionFixture()
		req.Parameters["tier_type"] = tier
		for _, stage := range []string{"before", "recheck", "after"} {
			r.bodies[stage] = strings.ReplaceAll(r.bodies[stage], `"tier_type":"cloud-s3"`, `"s3-glacier":{"glacier_restore_days":7},"tier_type":"`+tier+`"`)
		}
		s.executor = r
		result, err := s.Execute(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(result)
		if strings.Contains(string(raw), "new-access") || strings.Contains(string(raw), "secret") || result.Details.(map[string]any)["remote_connection_tested"] != false || len(r.calls) != 9 {
			t.Fatal("unsafe connection result")
		}
		write := r.calls[3]
		want := `endpoint="https://new.example:9443/s3",access_key="new-access",secret="literal\u002c\u007bsecret\u007d\"=true"`
		if write.Args[len(write.Args)-1] != want || len(write.SensitiveArgs) != 1 {
			t.Fatal("unencoded or unmarked credentials")
		}
		if _, ok := write.SensitiveArgs[len(write.Args)-1]; !ok {
			t.Fatal("wrong sensitive index")
		}
		for i, c := range r.calls {
			if i != 3 && len(c.SensitiveArgs) > 0 {
				t.Fatal("sensitive indices propagated to unrelated stage")
			}
		}
	}
	req.Parameters = connectionParams()
	for _, stage := range []string{"before", "realm_before", "recheck", "add", "after", "period.pre_check", "period.commit", "period.realm_post_check", "period.period_post_check"} {
		for _, mode := range []string{"error", "exit"} {
			r := connectionFixture()
			s.executor = r
			if mode == "error" {
				r.fail = stage
			} else {
				r.codeStage = stage
			}
			_, err := s.Execute(context.Background(), req)
			var ae *cephdomain.ActionError
			if !errors.As(err, &ae) || ae.Retryable || strings.Contains(err.Error(), "private") || r.calls[len(r.calls)-1].ID != "rgw_zonegroup.cloud_connection."+stage {
				t.Fatal(stage, err)
			}
		}
	}
	for _, scenario := range []string{"wrong_tier", "missing_key", "drift", "unrelated_change", "no_change"} {
		r := connectionFixture()
		s.executor = r
		switch scenario {
		case "wrong_tier":
			r.bodies["before"] = strings.ReplaceAll(r.bodies["before"], "cloud-s3", "unknown")
		case "missing_key":
			r.bodies["before"] = strings.ReplaceAll(r.bodies["before"], `"access_key":"old-access",`, "")
		case "drift":
			r.bodies["recheck"] = strings.ReplaceAll(r.bodies["recheck"], "old-access", "concurrent")
		case "unrelated_change":
			r.bodies["after"] = strings.ReplaceAll(r.bodies["after"], "restricted", "changed")
		case "no_change":
			r.bodies["before"] = r.bodies["after"]
		}
		if _, err := s.Execute(context.Background(), req); err == nil {
			t.Fatal(scenario)
		}
		if scenario != "unrelated_change" {
			for _, c := range r.calls {
				if c.Mutating {
					t.Fatal("unsafe write", scenario)
				}
			}
		}
	}
}

func TestCloudConnectionValidation(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	r := connectionFixture()
	s.executor = r
	p := connectionParams()
	p["credentials_saved"] = false
	if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.cloud_connection", Parameters: p}); err == nil || len(r.calls) != 0 {
		t.Fatal("unconfirmed connection reached native execution")
	}
	for _, endpoint := range []string{"", "ftp://host", "https://user:pass@host", "https://@host", "https://host?", "https://host#", "https://host?token=x", "https://host\\path", "https://host\n", "https://host:0", "https://host:65536", "https://host:", "https:///path", "https://host:bad"} {
		p := connectionParams()
		p["endpoint"] = endpoint
		if _, err := buildCloudConnection(p); err == nil {
			t.Fatal("bad endpoint accepted", endpoint)
		}
	}
	for _, endpoint := range []string{"http://host/path", "https://host", "https://[::1]:443/path", "https://host/path%20name"} {
		if !cloudConnectionEndpoint(endpoint) {
			t.Fatal(endpoint)
		}
	}
	for _, key := range []string{"access_key", "secret"} {
		for _, value := range []any{nil, "", " ", "[REDACTED]", "a\n", strings.Repeat("x", 4097)} {
			p := connectionParams()
			p[key] = value
			if _, err := buildCloudConnection(p); err == nil {
				t.Fatal("bad credential accepted", key)
			}
		}
	}
	for key, value := range map[string]any{"realm_id": "", "tier_type": "local", "storage_class": "STANDARD", "confirm_connection": false, "credentials_saved": false} {
		p := connectionParams()
		p[key] = value
		if _, err := buildCloudConnection(p); err == nil {
			t.Fatal(key)
		}
	}
}

func TestCloudConnectionInitializesEmptyDefault(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	p := connectionParams()
	p["expected_default_placement"] = ""
	r := connectionFixture()
	for _, stage := range []string{"before", "recheck", "after"} {
		value := ""
		if stage == "after" {
			value = "p"
		}
		r.bodies[stage] = strings.ReplaceAll(r.bodies[stage], `"default_placement":"other"`, `"default_placement":"`+value+`"`)
	}
	s.executor = r
	if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.cloud_connection", Parameters: p}); err != nil {
		t.Fatal(err)
	}
}
