package external

import (
	"context"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	endpointservice "cephtower/backend/internal/service/endpoint"
)

func TestIAMRoleManagedPolicyVerifiedStages(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	_, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://rgw.test"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "private-key"}})
	if err != nil {
		t.Fatal(err)
	}
	const account = "RGW12345678901234567"
	const arn = "arn:aws:iam::RGW12345678901234567:role/path/role"
	const policy = "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
	for _, mode := range []string{"attach", "detach"} {
		for _, fault := range []string{"", "owner", "key", "role id", "role arn", "pre read", "pre list", "snapshot", "write", "post role", "role drift", "post list", "policy drift", "missing target", "unchanged"} {
			t.Run(mode+"/"+fault, func(t *testing.T) {
				calls, ownerCalls := 0, 0
				service.topicOwners = topicOwnerVerifierFunc(func(_ context.Context, id uint64, uid, key, secret string) (string, string, error) {
					ownerCalls++
					if id != cluster.ID || uid != "tenant$user" || key != "access" || secret != "private-key" {
						t.Fatal("wrong owner check")
					}
					if fault == "owner" {
						return account, "tenant$user", nil
					}
					return account, account, nil
				})
				service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
					calls++
					if r.ParseForm() != nil || !strings.Contains(r.Header.Get("Authorization"), "/iam/aws4_request") || r.Form.Get("RoleName") != "role" || r.URL.RawQuery != "" {
						t.Fatal("wrong IAM target")
					}
					action := r.Form.Get("Action")
					wantAction := "GetRole"
					if calls == 2 || calls == 5 {
						wantAction = "ListAttachedRolePolicies"
					}
					if calls == 3 {
						wantAction = "AttachRolePolicy"
						if mode == "detach" {
							wantAction = "DetachRolePolicy"
						}
						if r.Form.Get("PolicyArn") != policy {
							t.Fatal("wrong policy")
						}
					}
					if action != wantAction {
						t.Fatal("wrong request order")
					}
					status := 200
					body := "<" + action + "Response><ResponseMetadata><RequestId>id</RequestId></ResponseMetadata>"
					if action == "GetRole" {
						fields := map[string]string{"RoleId": "id", "RoleName": "role", "Arn": arn, "Path": "/path/", "CreateDate": "now", "Description": "old", "MaxSessionDuration": "3600", "AssumeRolePolicyDocument": "{}"}
						if fault == "role id" {
							fields["RoleId"] = "other"
						}
						if fault == "role arn" {
							fields["Arn"] = "other"
						}
						if calls == 4 && fault == "role drift" {
							fields["Description"] = "changed"
						}
						body += "<GetRoleResult><Role>"
						for key, value := range fields {
							var b strings.Builder
							_ = xml.EscapeText(&b, []byte(value))
							body += "<" + key + ">" + b.String() + "</" + key + ">"
						}
						body += "</Role></GetRoleResult>"
					} else if action == "ListAttachedRolePolicies" {
						present := mode == "detach"
						if calls == 5 {
							present = !present
						}
						if calls == 2 && fault == "snapshot" {
							present = !present
						}
						if calls == 5 && fault == "missing target" {
							present = !present
						}
						body += "<ListAttachedRolePoliciesResult><AttachedPolicies>"
						if present {
							body += "<member><PolicyName>AmazonS3ReadOnlyAccess</PolicyName><PolicyArn>" + policy + "</PolicyArn></member>"
						}
						if calls == 5 && fault == "policy drift" {
							body += "<member><PolicyName>Other</PolicyName><PolicyArn>arn:aws:iam::aws:policy/Other</PolicyArn></member>"
						}
						body += "</AttachedPolicies></ListAttachedRolePoliciesResult>"
					}
					body += "</" + action + "Response>"
					if (calls == 1 && fault == "pre read") || (calls == 2 && fault == "pre list") || (calls == 3 && fault == "write") || (calls == 4 && fault == "post role") || (calls == 5 && fault == "post list") {
						status = 403
						body = "private-key"
					}
					return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
				})
				expected := []any{}
				if mode == "detach" {
					expected = append(expected, policy)
				}
				if fault == "unchanged" {
					if mode == "attach" {
						expected = append(expected, policy)
					} else {
						expected = []any{}
					}
				}
				key := "rgw/role/" + account + "/role"
				if fault == "key" {
					key = "rgw/role/other/role"
				}
				_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_role.managed_policy", ResourceKey: key, Parameters: map[string]any{"account_id": account, "name": "role", "mode": mode, "policy_arn": policy, "owner_uid": "tenant$user", "expected_role_id": "id", "expected_role_arn": arn, "expected_policies": expected}})
				wantCalls, code := 5, "post_check_failed"
				switch fault {
				case "":
					code = ""
				case "owner":
					wantCalls, code = 0, "pre_check_failed"
				case "key", "unchanged":
					wantCalls, code = 0, "invalid_request"
				case "role id", "role arn", "pre read":
					wantCalls, code = 1, "pre_check_failed"
				case "pre list", "snapshot":
					wantCalls, code = 2, "pre_check_failed"
				case "write":
					wantCalls, code = 3, "iam_failed"
				case "post role", "role drift":
					wantCalls = 4
				}
				if calls != wantCalls {
					t.Fatalf("calls %d want %d", calls, wantCalls)
				}
				if (fault == "key" || fault == "unchanged") && ownerCalls != 0 {
					t.Fatal("invalid request reached owner check")
				}
				if code == "" {
					if err != nil {
						t.Fatal(err)
					}
				} else {
					var failure *cephdomain.ActionError
					if !errors.As(err, &failure) || failure.Code != code || failure.Retryable || strings.Contains(err.Error(), "private-key") {
						t.Fatalf("unsafe failure: %v", err)
					}
				}
			})
		}
	}
}

func TestIAMRoleManagedPolicyRequiresPermanentOwnerProof(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	_, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://rgw.test"})
	if err != nil {
		t.Fatal(err)
	}
	request := Request{ClusterID: cluster.ID, Action: "rgw_role.managed_policy", ResourceKey: "rgw/role/RGW12345678901234567/role", Parameters: map[string]any{"account_id": "RGW12345678901234567", "name": "role", "mode": "attach", "policy_arn": "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess", "owner_uid": "user", "expected_role_id": "id", "expected_role_arn": "arn:aws:iam::RGW12345678901234567:role/role", "expected_policies": []any{}}}
	service.transport = externalRoundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("unverified identity reached IAM")
		return nil, nil
	})
	for _, temporary := range []bool{false, true} {
		credential := map[string]any{"access_key": "key", "secret_key": "secret"}
		service.topicOwners = nil
		if temporary {
			credential["session_token"] = "temporary"
			service.topicOwners = topicOwnerVerifierFunc(func(context.Context, uint64, string, string, string) (string, string, error) {
				t.Fatal("temporary key accepted")
				return "", "", nil
			})
		}
		if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: credential}); err != nil {
			t.Fatal(err)
		}
		_, err := service.Execute(ctx, request)
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != "capability_unavailable" || failure.Retryable {
			t.Fatalf("unsafe owner guard: %v", err)
		}
	}
	for _, value := range []any{nil, []any{"a", "a"}, []any{42}, []any{""}, "[]"} {
		if _, ok := rolePolicySnapshot(value); ok {
			t.Fatal("invalid snapshot accepted")
		}
	}
}
