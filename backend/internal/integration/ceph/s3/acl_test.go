package s3

import (
	"strings"
	"testing"
)

const aclExample = `<AccessControlPolicy xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Owner><ID>team$owner</ID><DisplayName>A&amp;B</DisplayName></Owner><AccessControlList><Grant><Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="CanonicalUser"><ID>team$owner</ID></Grantee><Permission>READ</Permission><Permission>WRITE_ACP</Permission></Grant><Grant><Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="Group"><URI>http://acs.amazonaws.com/groups/global/AllUsers</URI></Grantee><Permission>READ</Permission></Grant></AccessControlList></AccessControlPolicy>`

func TestBucketACLNativeOutput(t *testing.T) {
	acl, err := BucketACL([]byte(aclExample))
	if err != nil {
		t.Fatal(err)
	}
	if acl.Owner.ID != "team$owner" || acl.Owner.DisplayName != "A&B" || len(acl.Grants) != 2 || len(acl.Grants[0].Permissions) != 2 || acl.Grants[0].Permissions[1] != "WRITE_ACP" || acl.Grants[1].Grantee.Type != "Group" {
		t.Fatalf("lost ACL: %#v", acl)
	}
	for _, body := range []string{
		strings.ReplaceAll(aclExample, "CanonicalUser", "FutureType"),
		strings.ReplaceAll(aclExample, ">READ<", ">FUTURE_PERMISSION<"),
		strings.Replace(aclExample, "<Permission>READ</Permission><Permission>WRITE_ACP</Permission>", "", 1),
		`<AccessControlPolicy><Owner><ID>owner</ID></Owner><AccessControlList/></AccessControlPolicy>`,
		strings.Replace(aclExample, `xsi:type="CanonicalUser"><ID>team$owner</ID>`, `xsi:type="AmazonCustomerByEmail"><EmailAddress>a@example.test</EmailAddress>`, 1),
	} {
		if _, err := BucketACL([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBucketACLRejectsAmbiguousResponses(t *testing.T) {
	for _, body := range []string{
		"", "<AccessControlPolicy/>", "<Error><Code>AccessDenied</Code></Error>",
		aclExample + aclExample, "<!DOCTYPE x>" + aclExample,
		strings.Replace(aclExample, "<ID>team$owner</ID>", "<ID>one</ID><ID>two</ID>", 1),
		strings.Replace(aclExample, "<Owner>", "<Owner>unexpected", 1),
		strings.Replace(aclExample, "<ID>team$owner</ID>", "<ID><nested/></ID>", 1),
		strings.ReplaceAll(aclExample, "xsi:type=", "type="),
		strings.ReplaceAll(aclExample, "http://www.w3.org/2001/XMLSchema-instance", "wrong"),
		strings.Replace(aclExample, "<URI>http://acs.amazonaws.com/groups/global/AllUsers</URI>", "", 1),
		strings.Replace(aclExample, "<Permission>READ</Permission>", "<Permission><nested/></Permission>", 1),
		strings.Replace(aclExample, "<Permission>READ</Permission>", "<Permission/>", 1),
	} {
		if _, err := BucketACL([]byte(body)); err == nil {
			t.Errorf("invalid ACL accepted: %s", body)
		}
	}
	if ValidateBucketConfiguration("acl", []byte(aclExample)) == nil || DeletableBucketConfiguration("acl") {
		t.Fatal("read-only ACL unexpectedly writable")
	}
}
