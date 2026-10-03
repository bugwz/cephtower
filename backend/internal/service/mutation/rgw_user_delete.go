package mutation

import "encoding/json"

// Unbounded native user list returns an array after walking every metadata page.
// Do not accept paginated objects as evidence of global user absence.
func rgwUserPresence(raw []byte, uid string, present bool) bool {
	var users []string
	if uid == "" || json.Unmarshal(raw, &users) != nil || users == nil {
		return false
	}
	seen := make(map[string]bool, len(users))
	for _, user := range users {
		if user == "" || seen[user] {
			return false
		}
		seen[user] = true
	}
	return seen[uid] == present
}
