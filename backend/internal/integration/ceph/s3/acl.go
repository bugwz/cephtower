package s3

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type BucketACLConfiguration struct {
	Owner  BucketACLOwner   `json:"owner"`
	Grants []BucketACLGrant `json:"grants"`
}
type BucketACLOwner struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}
type BucketACLGrantee struct {
	Type         string `json:"type"`
	ID           string `json:"id"`
	DisplayName  string `json:"display_name"`
	URI          string `json:"uri"`
	EmailAddress string `json:"email_address"`
}
type BucketACLGrant struct {
	Grantee     BucketACLGrantee `json:"grantee"`
	Permissions []string         `json:"permissions"`
}
type aclNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Text     string     `xml:",chardata"`
	Children []aclNode  `xml:",any"`
}

// BucketACL preserves grants and all permission elements emitted by rgw_acl_s3.cc.
// It does not infer effective authorization or convert unknown values to private.
func BucketACL(body []byte) (BucketACLConfiguration, error) {
	result := BucketACLConfiguration{Grants: []BucketACLGrant{}}
	if err := validateConfigurationXML(body, "AccessControlPolicy"); err != nil {
		return result, err
	}
	var root aclNode
	if err := xml.Unmarshal(body, &root); err != nil {
		return result, err
	}
	invalid := fmt.Errorf("invalid native AccessControlPolicy structure")
	children := func(node aclNode, allowed map[string]bool) (map[string][]aclNode, bool) {
		if strings.TrimSpace(node.Text) != "" {
			return nil, false
		}
		fields := map[string][]aclNode{}
		for _, child := range node.Children {
			repeated, exists := allowed[child.XMLName.Local]
			if !exists || (!repeated && len(fields[child.XMLName.Local]) != 0) {
				return nil, false
			}
			fields[child.XMLName.Local] = append(fields[child.XMLName.Local], child)
		}
		return fields, true
	}
	scalar := func(fields map[string][]aclNode, key string) (string, bool) {
		nodes := fields[key]
		if len(nodes) == 0 {
			return "", true
		}
		return nodes[0].Text, len(nodes) == 1 && len(nodes[0].Children) == 0
	}
	fields, ok := children(root, map[string]bool{"Owner": false, "AccessControlList": false})
	if !ok || len(fields["Owner"]) != 1 || len(fields["AccessControlList"]) != 1 {
		return result, invalid
	}
	owner, ok := children(fields["Owner"][0], map[string]bool{"ID": false, "DisplayName": false})
	if !ok {
		return result, invalid
	}
	result.Owner.ID, ok = scalar(owner, "ID")
	if !ok || result.Owner.ID == "" {
		return result, invalid
	}
	result.Owner.DisplayName, ok = scalar(owner, "DisplayName")
	if !ok {
		return result, invalid
	}
	list, ok := children(fields["AccessControlList"][0], map[string]bool{"Grant": true})
	if !ok {
		return result, invalid
	}
	for _, node := range list["Grant"] {
		fields, ok := children(node, map[string]bool{"Grantee": false, "Permission": true})
		if !ok || len(fields["Grantee"]) != 1 {
			return result, invalid
		}
		grantee := fields["Grantee"][0]
		grant := BucketACLGrant{Permissions: []string{}}
		for _, attr := range grantee.Attrs {
			if attr.Name.Local == "type" && attr.Name.Space == "http://www.w3.org/2001/XMLSchema-instance" {
				if grant.Grantee.Type != "" {
					return result, invalid
				}
				grant.Grantee.Type = attr.Value
			}
		}
		if grant.Grantee.Type == "" {
			return result, invalid
		}
		identity, ok := children(grantee, map[string]bool{"ID": false, "DisplayName": false, "URI": false, "EmailAddress": false})
		if !ok {
			return result, invalid
		}
		for key, target := range map[string]*string{"ID": &grant.Grantee.ID, "DisplayName": &grant.Grantee.DisplayName, "URI": &grant.Grantee.URI, "EmailAddress": &grant.Grantee.EmailAddress} {
			*target, ok = scalar(identity, key)
			if !ok {
				return result, invalid
			}
		}
		if (grant.Grantee.Type == "CanonicalUser" && grant.Grantee.ID == "") || (grant.Grantee.Type == "Group" && grant.Grantee.URI == "") || (grant.Grantee.Type == "AmazonCustomerByEmail" && grant.Grantee.EmailAddress == "") {
			return result, invalid
		}
		for _, permission := range fields["Permission"] {
			if len(permission.Children) != 0 || permission.Text == "" {
				return result, invalid
			}
			grant.Permissions = append(grant.Permissions, permission.Text)
		}
		result.Grants = append(result.Grants, grant)
	}
	return result, nil
}
