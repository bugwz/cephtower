package mutation

import (
	"reflect"
	"strings"
	"unicode"
)

func syncPipeDestinationOwner(p map[string]any) (string, bool, error) {
	raw, present := p["dest_owner"]
	if !present {
		return "", false, nil
	}
	uid, ok := raw.(string)
	if !ok {
		return "", true, invalid("destination owner must be a string")
	}
	if uid == "" {
		return uid, true, nil
	}
	if !syncFlowToken(uid) || strings.IndexFunc(uid, unicode.IsSpace) >= 0 {
		return "", true, invalid("canonical destination owner UID is required")
	}
	// Mirror rgw_user::from_str/to_str without silently normalizing identity.
	parts := strings.SplitN(uid, "$", 3)
	if (len(parts) == 2 && (parts[0] == "" || parts[1] == "")) || (len(parts) == 3 && (parts[1] == "" || parts[2] == "")) {
		return "", true, invalid("canonical destination owner UID is required")
	}
	return uid, true, nil
}

func updateSyncPipeACL(params map[string]any, p map[string]any) (bool, error) {
	owner, present, err := syncPipeDestinationOwner(p)
	if err != nil || !present {
		return false, err
	}
	dest, ok := params["dest"].(map[string]any)
	if !ok {
		return false, invalid("pipe destination params unavailable")
	}
	old, exists := dest["acl_translation"]
	if owner == "" {
		delete(dest, "acl_translation")
		return exists, nil
	}
	desired := map[string]any{"owner": owner}
	dest["acl_translation"] = desired
	return !reflect.DeepEqual(old, desired), nil
}
