package s3

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCannedACLHeaderParticipatesInSignature(t *testing.T) {
	client := &Client{credentials: Credentials{AccessKey: "access", SecretKey: "secret", Region: "us-east-1"}}
	if err := client.PutBucketCannedACL(context.Background(), "bucket", "public-read\r\nx-amz-test: true"); err == nil {
		t.Fatal("invalid header accepted")
	}
	signatures := map[string]bool{}
	for _, preset := range []string{"private", "public-read", "public-read-write", "authenticated-read"} {
		req, _ := http.NewRequest("PUT", "https://s3.example.test/team:bucket?acl=", nil)
		req.Header.Set("X-Amz-Acl", preset)
		client.sign(req, time.Unix(0, 0).UTC(), sha256Hex(nil))
		signature := req.Header.Get("Authorization")
		if !strings.Contains(signature, "SignedHeaders=host;x-amz-acl;") || signatures[signature] {
			t.Fatal("ACL value does not affect signature")
		}
		signatures[signature] = true
	}
}

func TestBucketCannedACLMatchesNativePermissionSets(t *testing.T) {
	owner := BucketACLGrantee{Type: "CanonicalUser", ID: "owner"}
	group := BucketACLGrantee{Type: "Group", URI: "http://acs.amazonaws.com/groups/global/AllUsers"}
	full := BucketACLGrant{Grantee: owner, Permissions: []string{"FULL_CONTROL"}}
	for _, canned := range []string{"private", "public-read", "public-read-write", "authenticated-read"} {
		actual := BucketACLConfiguration{Owner: BucketACLOwner{ID: "owner"}, Grants: []BucketACLGrant{full}}
		if canned != "private" {
			recipient := group
			if canned == "authenticated-read" {
				recipient.URI = "http://acs.amazonaws.com/groups/global/AuthenticatedUsers"
			}
			actual.Grants = append(actual.Grants, BucketACLGrant{Grantee: recipient, Permissions: []string{"READ"}})
			if canned == "public-read-write" {
				actual.Grants = append(actual.Grants, BucketACLGrant{Grantee: recipient, Permissions: []string{"WRITE"}})
			}
		}
		if !BucketCannedACLMatches("owner", canned, actual) {
			t.Fatal("native preset rejected: " + canned)
		}
		actual.Owner.ID = "different"
		if BucketCannedACLMatches("owner", canned, actual) {
			t.Fatal("owner drift accepted")
		}
	}
	for _, grants := range [][]BucketACLGrant{
		nil,
		{{Grantee: owner, Permissions: []string{"READ"}}},
		{{Grantee: owner, Permissions: []string{"FUTURE"}}},
		{full, {Grantee: group, Permissions: []string{"READ"}}},
		{full, {Grantee: owner, Permissions: []string{}}},
		{{Grantee: BucketACLGrantee{Type: "CanonicalUser", ID: "other"}, Permissions: []string{"FULL_CONTROL"}}},
	} {
		if BucketCannedACLMatches("owner", "private", BucketACLConfiguration{Owner: BucketACLOwner{ID: "owner"}, Grants: grants}) {
			t.Fatal("unexpected private ACL accepted")
		}
	}
	expanded := BucketACLConfiguration{Owner: BucketACLOwner{ID: "owner", DisplayName: "changed"}, Grants: []BucketACLGrant{{Grantee: owner, Permissions: []string{"WRITE_ACP", "READ", "WRITE", "READ_ACP", "READ"}}}}
	if !BucketCannedACLMatches("owner", "private", expanded) {
		t.Fatal("equivalent set rejected")
	}
	for _, invalid := range []string{"", "PUBLIC-READ", "public-read\n", "bucket-owner-read", "unknown"} {
		if ValidBucketCannedACL(invalid) || BucketCannedACLMatches("owner", invalid, expanded) {
			t.Fatal("unsupported preset accepted")
		}
	}
}
