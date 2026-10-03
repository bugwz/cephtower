package external

import (
	"context"
	"reflect"
	"regexp"
	"sort"
	"strings"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/s3"
)

var iamAccountID = regexp.MustCompile(`^RGW[0-9]{17}$`)
var iamManagedRoleName = regexp.MustCompile(`^[A-Za-z0-9_+=,.@-]{1,64}$`)

func rolePolicySnapshot(value any) ([]string, bool) {
	items, ok := value.([]any)
	if !ok || items == nil {
		return nil, false
	}
	result := []string{}
	seen := map[string]bool{}
	for _, item := range items {
		arn, valid := item.(string)
		if !valid || arn == "" || seen[arn] {
			return nil, false
		}
		seen[arn] = true
		result = append(result, arn)
	}
	sort.Strings(result)
	return result, true
}

func (s *Service) roleManagedPolicy(ctx context.Context, request Request) (cephdomain.ActionResult, error) {
	p := request.Parameters
	text := func(key string) string { value, _ := p[key].(string); return value }
	account, name, mode, policy := text("account_id"), text("name"), text("mode"), text("policy_arn")
	roleID, roleARN := text("expected_role_id"), text("expected_role_arn")
	expected, valid := rolePolicySnapshot(p["expected_policies"])
	if !iamAccountID.MatchString(account) || !iamManagedRoleName.MatchString(name) || request.ResourceKey != "rgw/role/"+account+"/"+name || (mode != "attach" && mode != "detach") || len(policy) < 20 || len(policy) > 2048 || strings.ContainsAny(policy, "\x00\r\n") || !valid || roleID == "" || !strings.HasPrefix(roleARN, "arn:aws:iam::"+account+":role/") || !strings.HasSuffix(roleARN, "/"+name) || text("owner_uid") == "" {
		return cephdomain.ActionResult{}, failure("invalid_request", "account role identity and complete managed policy snapshot required", false)
	}
	desired := []string{}
	present := false
	for _, arn := range expected {
		if arn == policy {
			present = true
		} else {
			desired = append(desired, arn)
		}
	}
	if (mode == "attach" && present) || (mode == "detach" && !present) {
		return cephdomain.ActionResult{}, failure("invalid_request", "managed policy selection does not change the snapshot", false)
	}
	if mode == "attach" {
		desired = append(desired, policy)
	}
	sort.Strings(desired)
	endpoint, credential, client, err := s.httpClient(ctx, request.ClusterID, "s3")
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	if s.topicOwners == nil || credential.SessionToken != "" {
		return cephdomain.ActionResult{}, failure("capability_unavailable", "IAM role changes require native owner verification and permanent S3 credentials", false)
	}
	// The shared verifier proves an active permanent key belongs to the exact UID
	// and returns both scope and owner as the account ID for account users.
	scope, owner, err := s.topicOwners.VerifyTopicOwner(ctx, request.ClusterID, text("owner_uid"), credential.AccessKey, credential.SecretKey)
	if err != nil || scope != account || owner != account {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "configured S3 key does not belong to the requested account", false)
	}
	api, err := s3.New(endpoint.URL, s3.Credentials{AccessKey: credential.AccessKey, SecretKey: credential.SecretKey, Region: credential.Region}, client)
	if err != nil {
		return cephdomain.ActionResult{}, failure("invalid_credential", "IAM credentials or endpoint invalid", false)
	}
	before, err := api.GetIAMRole(ctx, name)
	if err != nil || before["RoleId"] != roleID || before["Arn"] != roleARN {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "IAM role identity could not be verified; no change submitted", false)
	}
	policies, err := api.ListIAMRolePolicies(ctx, name)
	sort.Strings(policies)
	if err != nil || !reflect.DeepEqual(policies, expected) {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "IAM managed policy snapshot changed or is unavailable; no change submitted", false)
	}
	if err := api.SetIAMRolePolicy(ctx, name, policy, mode == "attach"); err != nil {
		return cephdomain.ActionResult{}, failure("iam_failed", "managed policy write outcome is uncertain; refresh before another change", false)
	}
	after, err := api.GetIAMRole(ctx, name)
	if err != nil || !reflect.DeepEqual(before, after) {
		return cephdomain.ActionResult{}, failure("post_check_failed", "policy change submitted but complete role identity and attributes could not be verified", false)
	}
	actual, err := api.ListIAMRolePolicies(ctx, name)
	sort.Strings(actual)
	if err != nil || !reflect.DeepEqual(actual, desired) {
		return cephdomain.ActionResult{}, failure("post_check_failed", "policy change submitted but complete managed policy set could not be verified", false)
	}
	return cephdomain.ActionResult{}, nil
}
