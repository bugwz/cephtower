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
	verb, err := enum(p, "action", "modify", "rm")
	if err != nil {
		return command{}, err
	}
	if rawText(p, "confirm_subuser") != uid+":"+name {
		return command{}, invalid("confirm_subuser must match the full subuser id")
	}
	args := []string{"subuser", verb, "--uid", uid, "--subuser=" + name}
	if verb == "modify" {
		subuser_permission, err := enum(p, "subuser_permission", "none", "read", "write", "readwrite", "full")
		if err != nil {
			return command{}, err
		}
		if subuser_permission == "none" {
			subuser_permission = ""
		}
		args = append(args, "--access="+subuser_permission)
	} else if _, present := p["subuser_permission"]; present {
		return command{}, invalid("subuser_permission is not accepted for removal")
	}
	return rgw(args, []string{"user", "info", "--uid", uid}), nil
}

// Verify native full identities, not cached names. Missing arrays are not empty lists.
func rgwSubuserMatches(raw []byte, p map[string]any, before bool) bool {
	var info struct {
		UID      string `json:"full_user_id"`
		Subusers []struct {
			ID          string  `json:"id"`
			Permissions *string `json:"permissions"`
		} `json:"subusers"`
		Keys []struct {
			User string `json:"user"`
		} `json:"keys"`
		SwiftKeys []struct {
			User string `json:"user"`
		} `json:"swift_keys"`
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
