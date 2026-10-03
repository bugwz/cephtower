package mutation

import (
	"encoding/json"
	"regexp"
	"strings"
)

var rgwSubuserName = regexp.MustCompile(`^[A-Za-z0-9_.@-]+$`)
var rgwSubuserPermissions = map[string]string{"none": "<none>", "read": "read", "write": "write", "readwrite": "read-write", "full": "full-control"}

func buildRGWSubuser(p map[string]any, rgw func([]string, []string) command) (command, error) {
	for _, field := range []string{"uid", "subuser", "action", "confirm_subuser", "subuser_permission"} {
		if value, present := p[field]; present {
			text, ok := value.(string)
			if !ok || text == "" || text != rawText(p, field) {
				return command{}, invalid(field + " must be an exact nonempty string")
			}
		}
	}
	uid, name := rawText(p, "uid"), rawText(p, "subuser")
	if !regexp.MustCompile(`^[A-Za-z0-9_.:@$-]+$`).MatchString(uid) || strings.HasPrefix(uid, "-") || !rgwSubuserName.MatchString(name) || strings.HasPrefix(name, "-") {
		return command{}, invalid("uid or local subuser name is invalid; qualified subuser names are not accepted")
	}
	verb, err := enum(p, "action", "create", "modify", "rm")
	if err != nil {
		return command{}, err
	}
	if rawText(p, "confirm_subuser") != uid+":"+name {
		return command{}, invalid("confirm_subuser must match the full subuser id")
	}
	args := []string{"subuser", verb, "--uid", uid, "--subuser=" + name}
	if verb != "rm" {
		permission, err := enum(p, "subuser_permission", "none", "read", "write", "readwrite", "full")
		if err != nil {
			return command{}, err
		}
		if permission == "none" {
			permission = ""
		}
		args = append(args, "--access="+permission)
	} else if _, present := p["subuser_permission"]; present {
		return command{}, invalid("subuser_permission is not accepted for removal")
	}
	if verb != "create" {
		for _, field := range []string{"key_type", "access_key", "secret_key"} {
			if _, present := p[field]; present {
				return command{}, invalid("key fields are only accepted for subuser creation")
			}
		}
		return rgw(args, []string{"user", "info", "--uid", uid}), nil
	}
	kind, ok := p["key_type"].(string)
	if !ok || (kind != "s3" && kind != "swift") {
		return command{}, invalid("key_type must be s3 or swift")
	}
	secret, ok := p["secret_key"].(string)
	if !ok || len(secret) == 0 || len(secret) > 256 || secret != strings.TrimSpace(secret) || regexp.MustCompile(`[\x00-\x1f\x7f]`).MatchString(secret) {
		return command{}, invalid("secret_key must be an explicit nonempty credential without surrounding whitespace or control characters")
	}
	args = append(args, "--key-type="+kind, "--secret-key="+secret)
	sensitive := map[int]struct{}{len(args) - 1: {}}
	if kind == "s3" {
		key, ok := p["access_key"].(string)
		if !ok || !regexp.MustCompile(`^[A-Za-z0-9]{1,128}$`).MatchString(key) {
			return command{}, invalid("access_key must contain 1 to 128 alphanumeric characters")
		}
		args = append(args, "--access-key="+key)
		sensitive[len(args)-1] = struct{}{}
	} else if _, present := p["access_key"]; present {
		return command{}, invalid("access_key is not accepted for Swift subusers")
	}
	cmd := rgw(args, []string{"user", "info", "--uid", uid})
	cmd.sensitive = sensitive
	return cmd, nil
}

type rgwSubuserKey struct {
	User   string `json:"user"`
	Access string `json:"access_key"`
	Secret string `json:"secret_key"`
	Active *bool  `json:"active"`
}

// Verify native full identities, not cached names. Missing arrays are not empty lists.
func rgwSubuserMatches(raw []byte, p map[string]any, before bool) bool {
	var info struct {
		UID      string `json:"full_user_id"`
		Subusers []struct {
			ID          string  `json:"id"`
			Permissions *string `json:"permissions"`
		} `json:"subusers"`
		Keys      []rgwSubuserKey `json:"keys"`
		SwiftKeys []rgwSubuserKey `json:"swift_keys"`
	}
	uid, id := rawText(p, "uid"), rawText(p, "confirm_subuser")
	if json.Unmarshal(raw, &info) != nil || info.UID != uid || uid == "" || info.Subusers == nil {
		return false
	}
	seen, found := map[string]bool{}, false
	for _, sub := range info.Subusers {
		if sub.ID == "" || seen[sub.ID] {
			return false
		}
		seen[sub.ID] = true
		if sub.ID == id {
			found = true
			if !before && (rawText(p, "action") == "rm" || sub.Permissions == nil || *sub.Permissions != rgwSubuserPermissions[rawText(p, "subuser_permission")]) {
				return false
			}
		}
	}
	if rawText(p, "action") == "create" {
		if info.Keys == nil || info.SwiftKeys == nil {
			return false
		}
		if before {
			if found {
				return false
			}
			for _, key := range append(info.Keys, info.SwiftKeys...) {
				if key.User == "" || key.User == id || (rawText(p, "key_type") == "s3" && key.Access == rawText(p, "access_key")) {
					return false
				}
			}
			return true
		}
		if !found {
			return false
		}
		keys := info.SwiftKeys
		if rawText(p, "key_type") == "s3" {
			keys = info.Keys
		}
		matches := 0
		secret, _ := p["secret_key"].(string)
		for _, key := range keys {
			if key.User == id {
				if key.Secret != secret || key.Active == nil || !*key.Active || (rawText(p, "key_type") == "s3" && key.Access != rawText(p, "access_key")) {
					return false
				}
				matches++
			}
		}
		return matches == 1
	}
	if before || rawText(p, "action") == "modify" {
		return found
	}
	if found || info.Keys == nil || info.SwiftKeys == nil {
		return false
	}
	for _, key := range append(info.Keys, info.SwiftKeys...) {
		if key.User == "" || key.User == id {
			return false
		}
	}
	return true
}
