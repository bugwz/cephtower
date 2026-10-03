package mutation

import "encoding/json"

func rgwUserAccountRootMatches(raw []byte, uid, account string, root *bool) bool {
	var info struct {
		UID     string `json:"full_user_id"`
		Account string `json:"account_id"`
		Type    string `json:"type"`
	}
	if json.Unmarshal(raw, &info) != nil || uid == "" || account == "" || info.UID != uid || info.Account != account || (info.Type != "root" && info.Type != "rgw") {
		return false
	}
	return root == nil || (info.Type == "root") == *root
}
