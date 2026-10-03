package mutation

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

var topicOwnerUID = regexp.MustCompile(`^[A-Za-z0-9_.$@-]{1,512}$`)
var topicAccountID = regexp.MustCompile(`^RGW[0-9]{17}$`)

// Verify the configured permanent S3 key in memory, without exporting user keys.
// ListBuckets reports the user UID, not the account owner used for SNS topics.
func (s *Service) VerifyTopicOwner(ctx context.Context, clusterID uint64, uid, accessKey, secretKey string) (string, string, error) {
	invalid := fmt.Errorf("configured S3 credential owner could not be verified")
	if !topicOwnerUID.MatchString(uid) || accessKey == "" || secretKey == "" {
		return "", "", invalid
	}
	access, err := s.clusters.Access(ctx, clusterID)
	if err != nil {
		return "", "", invalid
	}
	defer func() { access.ClientKey = "" }()
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: "rgw_topic.owner_check", Binary: executor.BinaryRGWAdmin, Args: []string{"user", "info", "--uid=" + uid, "--format", "json"}, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	if err != nil || result.ExitCode != 0 {
		return "", "", invalid
	}
	return verifiedTopicOwner(result.Stdout, uid, accessKey, secretKey)
}

func verifiedTopicOwner(raw []byte, uid, accessKey, secretKey string) (string, string, error) {
	invalid := fmt.Errorf("configured S3 credential owner could not be verified")
	var info struct {
		UID       string          `json:"full_user_id"`
		Tenant    *string         `json:"tenant"`
		Account   *string         `json:"account_id"`
		Suspended *int            `json:"suspended"`
		Keys      []rgwSubuserKey `json:"keys"`
	}
	if json.Unmarshal(raw, &info) != nil || info.UID != uid || info.Tenant == nil || info.Account == nil || info.Suspended == nil || *info.Suspended != 0 || info.Keys == nil {
		return "", "", invalid
	}
	scope := *info.Tenant
	owner := uid
	if strings.ContainsAny(scope, ":$\x00\r\n") {
		return "", "", invalid
	}
	if *info.Account != "" {
		if !topicAccountID.MatchString(*info.Account) {
			return "", "", invalid
		}
		scope = *info.Account
		owner = scope
	}
	matched := 0
	for _, key := range info.Keys {
		if key.Access == accessKey {
			if key.User != uid || key.Active == nil || !*key.Active || subtle.ConstantTimeCompare([]byte(key.Secret), []byte(secretKey)) != 1 {
				return "", "", invalid
			}
			matched++
		}
	}
	if matched != 1 {
		return "", "", invalid
	}
	return scope, owner, nil
}
