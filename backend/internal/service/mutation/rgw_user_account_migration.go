package mutation

import (
	"encoding/json"
	"regexp"
)

func rgwUserAccountMigrationRequested(request Request) bool {
	_, present := request.Parameters["target_account_id"]
	return request.Action == "rgw_user.update" && present
}

var rgwMigrationName = regexp.MustCompile(`^[A-Za-z0-9_+=,.@-]{1,64}$`)

func rgwMigrationUserTenant(raw []byte, uid string) (string, bool) {
	var user struct {
		UID     string  `json:"full_user_id"`
		Account *string `json:"account_id"`
		Tenant  *string `json:"tenant"`
		Name    string  `json:"display_name"`
		Type    string  `json:"type"`
	}
	if json.Unmarshal(raw, &user) != nil || user.UID != uid || user.Account == nil || *user.Account != "" || user.Tenant == nil || user.Type != "rgw" || !rgwMigrationName.MatchString(user.Name) {
		return "", false
	}
	return *user.Tenant, true
}

func rgwMigrationAccountMatches(raw []byte, id, tenant string) bool {
	var account struct {
		ID     string  `json:"id"`
		Tenant *string `json:"tenant"`
	}
	return json.Unmarshal(raw, &account) == nil && account.ID == id && account.Tenant != nil && *account.Tenant == tenant
}
