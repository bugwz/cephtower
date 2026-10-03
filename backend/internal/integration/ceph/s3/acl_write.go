package s3

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
)

func ValidBucketCannedACL(value string) bool {
	return value == "private" || value == "public-read" || value == "public-read-write" || value == "authenticated-read"
}

func (c *Client) PutBucketCannedACL(ctx context.Context, bucket, value string) error {
	if !ValidBucketCannedACL(value) {
		return fmt.Errorf("unsupported bucket canned ACL")
	}
	_, _, err := c.requestWithHeaders(ctx, http.MethodPut, bucket, url.Values{"acl": {""}}, nil, http.Header{"X-Amz-Acl": {value}})
	return err
}

// Compare permission sets, accepting separate or combined native grants while
// rejecting extra recipients, unknown permissions and owner changes.
func BucketCannedACLMatches(ownerID, canned string, actual BucketACLConfiguration) bool {
	if ownerID == "" || actual.Owner.ID != ownerID || !ValidBucketCannedACL(canned) {
		return false
	}
	type identity struct{ Type, ID, URI, Email string }
	owner := identity{Type: "CanonicalUser", ID: ownerID}
	permissions := func(values ...string) map[string]bool {
		result := map[string]bool{}
		for _, value := range values {
			result[value] = true
		}
		return result
	}
	wanted := map[identity]map[string]bool{owner: permissions("READ", "WRITE", "READ_ACP", "WRITE_ACP")}
	group := identity{Type: "Group", URI: "http://acs.amazonaws.com/groups/global/AllUsers"}
	switch canned {
	case "public-read":
		wanted[group] = permissions("READ")
	case "public-read-write":
		wanted[group] = permissions("READ", "WRITE")
	case "authenticated-read":
		group.URI = "http://acs.amazonaws.com/groups/global/AuthenticatedUsers"
		wanted[group] = permissions("READ")
	}
	observed := map[identity]map[string]bool{}
	for _, grant := range actual.Grants {
		id := identity{grant.Grantee.Type, grant.Grantee.ID, grant.Grantee.URI, grant.Grantee.EmailAddress}
		if _, exists := wanted[id]; !exists || len(grant.Permissions) == 0 {
			return false
		}
		if observed[id] == nil {
			observed[id] = map[string]bool{}
		}
		for _, value := range grant.Permissions {
			switch value {
			case "FULL_CONTROL":
				for _, part := range []string{"READ", "WRITE", "READ_ACP", "WRITE_ACP"} {
					observed[id][part] = true
				}
			case "READ", "WRITE", "READ_ACP", "WRITE_ACP":
				observed[id][value] = true
			default:
				return false
			}
		}
	}
	return reflect.DeepEqual(wanted, observed)
}
