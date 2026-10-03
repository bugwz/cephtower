package s3

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var iamRoleName = regexp.MustCompile(`^[A-Za-z0-9_+=,.@-]{1,64}$`)

// IAM scope comes from the authenticated account, not a caller-supplied tenant.
// Callers must verify the returned RoleId/Arn before authorizing a mutation.
func (c *Client) iamRoleRequest(ctx context.Context, action, role, policy string) ([]byte, error) {
	if c.base.Scheme != "https" || !iamRoleName.MatchString(role) {
		return nil, fmt.Errorf("IAM requires HTTPS and a valid role name")
	}
	values := url.Values{"Action": {action}, "Version": {"2010-05-08"}, "RoleName": {role}}
	if policy != "" {
		values.Set("PolicyArn", policy)
	}
	body, _, err := c.requestSignedTarget(ctx, http.MethodPost, "", nil, []byte(values.Encode()), http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}, "iam")
	if err != nil {
		return nil, fmt.Errorf("IAM role request failed")
	}
	return body, nil
}

func iamResponse(body []byte, action string) (lifecycleField, error) {
	invalid := fmt.Errorf("invalid IAM role response")
	if validateConfigurationXML(body, action+"Response") != nil {
		return lifecycleField{}, invalid
	}
	var root lifecycleField
	if xml.Unmarshal(body, &root) != nil || strings.TrimSpace(root.Text) != "" {
		return lifecycleField{}, invalid
	}
	seen := map[string]bool{}
	for _, child := range root.Children {
		name := child.XMLName.Local
		if seen[name] || (name != "ResponseMetadata" && name != action+"Result") || strings.TrimSpace(child.Text) != "" {
			return lifecycleField{}, invalid
		}
		seen[name] = true
		if name == "ResponseMetadata" && (len(child.Children) != 1 || child.Children[0].XMLName.Local != "RequestId" || len(child.Children[0].Children) != 0 || child.Children[0].Text == "") {
			return lifecycleField{}, invalid
		}
	}
	if !seen["ResponseMetadata"] {
		return lifecycleField{}, invalid
	}
	return root, nil
}

func iamOnlyChild(node lifecycleField, name string) (lifecycleField, error) {
	var found lifecycleField
	count := 0
	for _, child := range node.Children {
		if child.XMLName.Local == name {
			found = child
			count++
		}
	}
	if count != 1 {
		return lifecycleField{}, fmt.Errorf("incomplete IAM role response")
	}
	return found, nil
}

func (c *Client) GetIAMRole(ctx context.Context, role string) (map[string]string, error) {
	body, err := c.iamRoleRequest(ctx, "GetRole", role, "")
	if err != nil {
		return nil, err
	}
	root, err := iamResponse(body, "GetRole")
	if err != nil {
		return nil, err
	}
	result, err := iamOnlyChild(root, "GetRoleResult")
	if err != nil {
		return nil, err
	}
	node, err := iamOnlyChild(result, "Role")
	if err != nil || len(result.Children) != 1 || strings.TrimSpace(node.Text) != "" {
		return nil, fmt.Errorf("invalid IAM role result")
	}
	fields := map[string]string{}
	for _, child := range node.Children {
		key := child.XMLName.Local
		if _, exists := fields[key]; exists || len(child.Children) != 0 {
			return nil, fmt.Errorf("ambiguous IAM role result")
		}
		fields[key] = child.Text
	}
	for _, key := range []string{"RoleId", "RoleName", "Arn", "Path", "CreateDate", "Description", "MaxSessionDuration", "AssumeRolePolicyDocument"} {
		if _, ok := fields[key]; !ok {
			return nil, fmt.Errorf("incomplete IAM role result")
		}
	}
	if fields["RoleName"] != role || fields["RoleId"] == "" || fields["Arn"] == "" {
		return nil, fmt.Errorf("IAM role identity mismatch")
	}
	return fields, nil
}

func (c *Client) ListIAMRolePolicies(ctx context.Context, role string) ([]string, error) {
	body, err := c.iamRoleRequest(ctx, "ListAttachedRolePolicies", role, "")
	if err != nil {
		return nil, err
	}
	root, err := iamResponse(body, "ListAttachedRolePolicies")
	if err != nil {
		return nil, err
	}
	result, err := iamOnlyChild(root, "ListAttachedRolePoliciesResult")
	if err != nil {
		return nil, err
	}
	list, err := iamOnlyChild(result, "AttachedPolicies")
	if err != nil || len(result.Children) != 1 || strings.TrimSpace(list.Text) != "" {
		return nil, fmt.Errorf("incomplete IAM policy list")
	}
	policies := []string{}
	seen := map[string]bool{}
	for _, member := range list.Children {
		if member.XMLName.Local != "member" || len(member.Children) != 2 || strings.TrimSpace(member.Text) != "" {
			return nil, fmt.Errorf("invalid IAM policy member")
		}
		arn, arnErr := iamOnlyChild(member, "PolicyArn")
		name, nameErr := iamOnlyChild(member, "PolicyName")
		if arnErr != nil || nameErr != nil || len(arn.Children) != 0 || len(name.Children) != 0 || arn.Text == "" || name.Text == "" || seen[arn.Text] {
			return nil, fmt.Errorf("ambiguous IAM policy member")
		}
		seen[arn.Text] = true
		policies = append(policies, arn.Text)
	}
	return policies, nil
}

// This is a protocol primitive, not a verified high-level operation. The caller
// must check account/role identity and policy snapshots before and after writes.
func (c *Client) SetIAMRolePolicy(ctx context.Context, role, policy string, attach bool) error {
	if len(policy) < 20 || len(policy) > 2048 || strings.ContainsAny(policy, "\x00\r\n") {
		return fmt.Errorf("invalid IAM policy ARN")
	}
	action := "DetachRolePolicy"
	if attach {
		action = "AttachRolePolicy"
	}
	body, err := c.iamRoleRequest(ctx, action, role, policy)
	if err != nil {
		return err
	}
	root, err := iamResponse(body, action)
	if err != nil {
		return err
	}
	for _, child := range root.Children {
		if child.XMLName.Local != "ResponseMetadata" {
			return fmt.Errorf("unexpected IAM mutation result")
		}
	}
	return nil
}
