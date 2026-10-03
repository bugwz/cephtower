package mutation

import (
	"encoding/json"
	"regexp"
)

func rgwUserCreateKeyArgs(p map[string]any) ([]string, error) {
	access, hasAccess := p["access_key"]
	_, hasSecret := p["secret_key"]
	if !hasAccess && !hasSecret {
		return []string{"--generate-key=false"}, nil
	}
	id, ok := access.(string)
	secret, err := rgwSubuserSecret(p)
	if !hasAccess || !hasSecret || !ok || !regexp.MustCompile(`^[A-Za-z0-9]{1,128}$`).MatchString(id) || err != nil {
		return nil, invalid("user creation requires both a valid access_key and secret_key, or neither")
	}
	return []string{"--key-type=s3", "--access-key=" + id, "--secret-key=" + secret}, nil
}

func rgwUserCreateKeysMatch(raw []byte, p map[string]any) bool {
	var info struct {
		UID       string          `json:"full_user_id"`
		Keys      []rgwSubuserKey `json:"keys"`
		SwiftKeys []rgwSubuserKey `json:"swift_keys"`
	}
	if json.Unmarshal(raw, &info) != nil || info.UID != rawText(p, "uid") || info.Keys == nil || info.SwiftKeys == nil || len(info.SwiftKeys) != 0 {
		return false
	}
	access, provided := p["access_key"].(string)
	if !provided {
		return len(info.Keys) == 0
	}
	if len(info.Keys) != 1 {
		return false
	}
	key := info.Keys[0]
	secret, _ := p["secret_key"].(string)
	return key.User == info.UID && key.Access == access && key.Secret == secret && key.Active != nil && *key.Active
}
