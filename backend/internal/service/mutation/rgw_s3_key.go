package mutation

import (
	"encoding/json"
	"regexp"
	"strings"
)

func buildRGWS3KeyCreate(p map[string]any, rgw func([]string, []string) command) (command, error) {
	uid, _ := p["uid"].(string)
	if !regexp.MustCompile(`^[A-Za-z0-9_.:@$-]+$`).MatchString(uid) || strings.HasPrefix(uid, "-") {
		return command{}, invalid("uid is invalid")
	}
	owner := uid
	args := []string{"key", "create", "--uid", uid, "--key-type=s3"}
	if value, present := p["subuser"]; present {
		name, ok := value.(string)
		if !ok || !rgwSubuserName.MatchString(name) || strings.HasPrefix(name, "-") {
			return command{}, invalid("subuser must be a local subuser name")
		}
		owner += ":" + name
		args = append(args, "--subuser="+name)
	}
	if confirmed, _ := p["confirm_owner"].(string); confirmed != owner {
		return command{}, invalid("confirm_owner must match the full credential owner")
	}
	key, ok := p["access_key"].(string)
	if !ok || !regexp.MustCompile(`^[A-Za-z0-9]{1,128}$`).MatchString(key) {
		return command{}, invalid("access_key must contain 1 to 128 alphanumeric characters")
	}
	secret, err := rgwSubuserSecret(p)
	if err != nil {
		return command{}, err
	}
	args = append(args, "--access-key="+key, "--secret-key="+secret)
	cmd := rgw(args, []string{"user", "info", "--uid", uid})
	cmd.sensitive = map[int]struct{}{len(args) - 2: {}, len(args) - 1: {}}
	return cmd, nil
}

func rgwS3KeyCreationMatches(raw []byte, p map[string]any, before bool) bool {
	var info struct {
		UID      string `json:"full_user_id"`
		Subusers []struct {
			ID string `json:"id"`
		} `json:"subusers"`
		Keys []rgwSubuserKey `json:"keys"`
	}
	uid, _ := p["uid"].(string)
	if json.Unmarshal(raw, &info) != nil || uid == "" || info.UID != uid || info.Keys == nil {
		return false
	}
	owner := uid
	if name, present := p["subuser"]; present {
		local, ok := name.(string)
		if !ok {
			return false
		}
		owner += ":" + local
		count := 0
		for _, sub := range info.Subusers {
			if sub.ID == owner {
				count++
			}
		}
		if count != 1 {
			return false
		}
	}
	keyID, _ := p["access_key"].(string)
	secret, _ := p["secret_key"].(string)
	count := 0
	for _, key := range info.Keys {
		if key.Access == "" || key.User == "" {
			return false
		}
		if key.Access == keyID {
			if before || key.User != owner || key.Secret != secret || key.Active == nil || !*key.Active {
				return false
			}
			count++
		}
	}
	return before || count == 1
}
