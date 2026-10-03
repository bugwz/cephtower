package s3

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

const iamMetadata = `<ResponseMetadata><RequestId>request</RequestId></ResponseMetadata>`
const iamRoleXML = `<Role><RoleId>id</RoleId><RoleName>role+name</RoleName><Arn>arn:aws:iam::RGW12345678901234567:role/role+name</Arn><Path>/</Path><CreateDate>now</CreateDate><Description/><MaxSessionDuration>3600</MaxSessionDuration><AssumeRolePolicyDocument>{}</AssumeRolePolicyDocument></Role>`

func TestIAMRoleProtocol(t *testing.T) {
	for _, action := range []string{"GetRole", "ListAttachedRolePolicies", "AttachRolePolicy", "DetachRolePolicy"} {
		t.Run(action, func(t *testing.T) {
			calls := 0
			client, err := New("https://rgw.example.test", Credentials{AccessKey: "key", SecretKey: "secret", SessionToken: "session"}, &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				body, _ := io.ReadAll(r.Body)
				if r.Method != "POST" || r.URL.Path != "/" || r.URL.RawQuery != "" || !strings.Contains(r.Header.Get("Authorization"), "/iam/aws4_request") || r.Header.Get("X-Amz-Security-Token") != "session" || r.Header.Get("X-Amz-Content-Sha256") != sha256Hex(body) {
					t.Fatal("wrong IAM transport/signing")
				}
				r.Body = io.NopCloser(strings.NewReader(string(body)))
				if r.ParseForm() != nil || r.PostForm.Get("Action") != action || r.PostForm.Get("Version") != "2010-05-08" || r.PostForm.Get("RoleName") != "role+name" {
					t.Fatal("wrong IAM form")
				}
				result := ""
				switch action {
				case "GetRole":
					result = "<GetRoleResult>" + iamRoleXML + "</GetRoleResult>"
				case "ListAttachedRolePolicies":
					result = `<ListAttachedRolePoliciesResult><AttachedPolicies><member><PolicyName>AmazonS3ReadOnlyAccess</PolicyName><PolicyArn>arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess</PolicyArn></member></AttachedPolicies></ListAttachedRolePoliciesResult>`
				default:
					if r.PostForm.Get("PolicyArn") != "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess" {
						t.Fatal("wrong policy")
					}
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("<" + action + "Response>" + iamMetadata + result + "</" + action + "Response>"))}, nil
			})})
			if err != nil {
				t.Fatal(err)
			}
			switch action {
			case "GetRole":
				var fields map[string]string
				fields, err = client.GetIAMRole(context.Background(), "role+name")
				if fields["RoleId"] != "id" {
					t.Fatal("identity lost")
				}
			case "ListAttachedRolePolicies":
				var arns []string
				arns, err = client.ListIAMRolePolicies(context.Background(), "role+name")
				if len(arns) != 1 {
					t.Fatal("policies missing")
				}
			default:
				err = client.SetIAMRolePolicy(context.Background(), "role+name", "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess", action == "AttachRolePolicy")
			}
			if err != nil || calls != 1 {
				t.Fatalf("IAM failed: %v calls=%d", err, calls)
			}
		})
	}
}

func TestIAMRoleFailClosed(t *testing.T) {
	for _, tc := range []struct {
		name, endpoint, role, body string
		status, calls              int
	}{
		{"http", "http://rgw.test", "role+name", "", 200, 0},
		{"bad name", "https://rgw.test", "bad/name", "", 200, 0},
		{"remote error", "https://rgw.test", "role+name", "secret-value", 403, 1},
		{"redirect", "https://rgw.test", "role+name", "", 307, 1},
		{"missing result", "https://rgw.test", "role+name", "<GetRoleResponse>" + iamMetadata + "</GetRoleResponse>", 200, 1},
		{"duplicate role", "https://rgw.test", "role+name", "<GetRoleResponse>" + iamMetadata + "<GetRoleResult>" + iamRoleXML + iamRoleXML + "</GetRoleResult></GetRoleResponse>", 200, 1},
		{"trailing xml", "https://rgw.test", "role+name", "<GetRoleResponse/> <Other/>", 200, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			client, _ := New(tc.endpoint, Credentials{AccessKey: "key", SecretKey: "secret"}, &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				return &http.Response{StatusCode: tc.status, Header: http.Header{"Location": {"https://other.test/"}}, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})})
			_, err := client.GetIAMRole(context.Background(), tc.role)
			if err == nil || strings.Contains(err.Error(), "secret-value") || calls != tc.calls {
				t.Fatalf("unsafe result: %v calls=%d", err, calls)
			}
		})
	}
}

func TestIAMPolicyListRejectsPartialOrAmbiguousResponses(t *testing.T) {
	member := `<member><PolicyName>ReadOnly</PolicyName><PolicyArn>arn:aws:iam::aws:policy/ReadOnly</PolicyArn></member>`
	for _, tc := range []struct {
		result string
		valid  bool
	}{
		{`<AttachedPolicies/>`, true},
		{`<AttachedPolicies>` + member + `</AttachedPolicies>`, true},
		{`<AttachedPolicies>` + member + member + `</AttachedPolicies>`, false},
		{`<AttachedPolicies/><IsTruncated>true</IsTruncated>`, false},
		{`<AttachedPolicies/><AttachedPolicies/>`, false},
		{`<AttachedPolicies>unknown</AttachedPolicies>`, false},
		{`<AttachedPolicies><member><PolicyArn>arn</PolicyArn></member></AttachedPolicies>`, false},
	} {
		client, _ := New("https://rgw.test", Credentials{AccessKey: "key", SecretKey: "secret"}, &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("<ListAttachedRolePoliciesResponse>" + iamMetadata + "<ListAttachedRolePoliciesResult>" + tc.result + "</ListAttachedRolePoliciesResult></ListAttachedRolePoliciesResponse>"))}, nil
		})})
		list, err := client.ListIAMRolePolicies(context.Background(), "role+name")
		if (err == nil) != tc.valid || (tc.valid && list == nil) {
			t.Fatalf("wrong list handling: %v", err)
		}
	}
}

func TestIAMPolicyMutationRejectsFalseSuccess(t *testing.T) {
	for _, attach := range []bool{false, true} {
		action := "DetachRolePolicy"
		if attach {
			action = "AttachRolePolicy"
		}
		for _, inner := range []string{"", "<Error>secret-value</Error>", iamMetadata + iamMetadata, iamMetadata + "<" + action + "Result/>", `<ResponseMetadata><RequestId/><RequestId>two</RequestId></ResponseMetadata>`} {
			client, _ := New("https://rgw.test", Credentials{AccessKey: "key", SecretKey: "secret"}, &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("<" + action + "Response>" + inner + "</" + action + "Response>"))}, nil
			})})
			err := client.SetIAMRolePolicy(context.Background(), "role+name", "arn:aws:iam::aws:policy/ReadOnly", attach)
			if err == nil || strings.Contains(err.Error(), "secret-value") {
				t.Fatal("ambiguous mutation accepted or exposed")
			}
		}
	}
	client, _ := New("https://rgw.test", Credentials{AccessKey: "key", SecretKey: "secret"}, &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) { t.Fatal("invalid ARN sent"); return nil, nil })})
	for _, arn := range []string{"", "short", strings.Repeat("a", 2049), "arn:aws:iam::aws:policy/bad\n"} {
		if client.SetIAMRolePolicy(context.Background(), "role", arn, true) == nil {
			t.Fatal("invalid ARN accepted")
		}
	}
}
