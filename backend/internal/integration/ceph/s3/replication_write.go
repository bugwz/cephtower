package s3

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Dashboard replication is a same-name bucket rule. RGW derives the destination
// tenant from the authenticated identity, not the account field of the ARN.
func (c *Client) PutDashboardBucketReplication(ctx context.Context, bucket string) error {
	parts := strings.Split(bucket, ":")
	if len(parts) != 2 || parts[1] == "" || strings.ContainsAny(bucket, "/\\\x00") {
		return fmt.Errorf("explicit tenant and bucket required")
	}
	if c.credentials.SessionToken != "" {
		return fmt.Errorf("replication preparation requires permanent RGW user credentials")
	}
	body, _, err := c.requestTarget(ctx, http.MethodGet, "", nil, nil, nil)
	if err != nil {
		return fmt.Errorf("cannot verify S3 credential tenant: %w", err)
	}
	if err := validateConfigurationXML(body, "ListAllMyBucketsResult"); err != nil {
		return err
	}
	var root lifecycleField
	if err := xml.Unmarshal(body, &root); err != nil {
		return err
	}
	owners := 0
	uid := ""
	for _, node := range root.Children {
		if node.XMLName.Local != "Owner" {
			continue
		}
		owners++
		ids := 0
		for _, field := range node.Children {
			if field.XMLName.Local == "ID" {
				ids++
				if len(field.Children) != 0 {
					return fmt.Errorf("invalid credential owner")
				}
				uid = field.Text
			}
		}
		if ids != 1 {
			return fmt.Errorf("ambiguous credential owner")
		}
	}
	if owners != 1 || uid == "" || strings.TrimSpace(uid) != uid {
		return fmt.Errorf("credential owner unavailable")
	}
	uidParts := strings.Split(uid, "$")
	tenant := ""
	if len(uidParts) > 1 {
		tenant = uidParts[0]
	}
	if len(uidParts) > 3 || uidParts[len(uidParts)-1] == "" || tenant != parts[0] {
		return fmt.Errorf("S3 credential tenant does not match the target bucket")
	}
	// Serialize fixed fields, never accept arbitrary XML for this operation.
	escape := func(value string) string {
		var b strings.Builder
		_ = xml.EscapeText(&b, []byte(value))
		return b.String()
	}
	document := `<ReplicationConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Role>` + escape(parts[1]+"_replication_role") + `</Role><Rule><ID>dashboard_admin_pipe</ID><Status>Enabled</Status><Priority>0</Priority><Filter><Prefix/></Filter><Destination><Bucket>` + escape("arn:aws:s3:::"+parts[1]) + `</Bucket></Destination></Rule></ReplicationConfiguration>`
	_, _, err = c.request(ctx, http.MethodPut, bucket, url.Values{"replication": {""}}, []byte(document))
	return err
}

// RGW discards Role and the empty filter, adds priority, and serializes the
// inferred tenant in the destination ARN. Reject additional effective fields.
func DashboardBucketReplicationMatches(body []byte, bucket string) bool {
	parts := strings.Split(bucket, ":")
	if len(parts) != 2 || parts[1] == "" || validateConfigurationXML(body, "ReplicationConfiguration") != nil {
		return false
	}
	var root lifecycleField
	if xml.Unmarshal(body, &root) != nil {
		return false
	}
	check := func(node lifecycleField, fields map[string]string) bool {
		if strings.TrimSpace(node.Text) != "" {
			return false
		}
		seen := map[string]bool{}
		for _, field := range node.Children {
			name := field.XMLName.Local
			value, ok := fields[name]
			if !ok || seen[name] || len(field.Children) != 0 || field.Text != value {
				return false
			}
			seen[name] = true
		}
		return len(seen) == len(fields)
	}
	if strings.TrimSpace(root.Text) != "" {
		return false
	}
	rules, roles := 0, 0
	for _, node := range root.Children {
		switch node.XMLName.Local {
		case "Role":
			roles++
			if node.Text != "" || len(node.Children) != 0 {
				return false
			}
		case "Rule":
			rules++
			if strings.TrimSpace(node.Text) != "" {
				return false
			}
			flat := lifecycleField{}
			destinations := 0
			for _, field := range node.Children {
				if field.XMLName.Local == "Destination" {
					destinations++
					if !check(field, map[string]string{"Bucket": "arn:aws:s3::" + parts[0] + ":" + parts[1]}) {
						return false
					}
				} else {
					flat.Children = append(flat.Children, field)
				}
			}
			if destinations != 1 || !check(flat, map[string]string{"ID": "dashboard_admin_pipe", "Status": "Enabled", "Priority": "0"}) {
				return false
			}
		default:
			return false
		}
	}
	return rules == 1 && roles == 1
}
