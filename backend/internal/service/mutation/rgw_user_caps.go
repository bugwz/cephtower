package mutation

import "encoding/json"

// Native RGWUserCaps::dump emits canonical permission strings, with * for both bits.
func rgwUserCaps(raw []byte, uid string) (map[string]uint8, bool) {
	var info struct {
		UID  string `json:"full_user_id"`
		Caps []struct {
			Type string `json:"type"`
			Perm string `json:"perm"`
		} `json:"caps"`
	}
	if json.Unmarshal(raw, &info) != nil || info.UID != uid || info.Caps == nil {
		return nil, false
	}
	caps := make(map[string]uint8, len(info.Caps))
	for _, cap := range info.Caps {
		if cap.Type == "" {
			return nil, false
		}
		if _, exists := caps[cap.Type]; exists {
			return nil, false
		}
		var bits uint8
		switch cap.Perm {
		case "<none>":
		case "read":
			bits = 1
		case "write":
			bits = 2
		case "*":
			bits = 3
		default:
			return nil, false
		}
		caps[cap.Type] = bits
	}
	return caps, true
}

func rgwExpectedCaps(raw []byte, p map[string]any) (map[string]uint8, bool) {
	caps, ok := rgwUserCaps(raw, rawText(p, "uid"))
	if !ok {
		return nil, false
	}
	bits := map[string]uint8{"read": 1, "write": 2, "read,write": 3, "*": 3}[rawText(p, "permission")]
	kind := rawText(p, "type")
	if rawText(p, "action") == "replace" {
		if _, exists := caps[kind]; !exists {
			return nil, false
		}
		caps[kind] = bits
	} else if rawText(p, "action") == "add" {
		caps[kind] |= bits
	} else {
		caps[kind] &^= bits
		if caps[kind] == 0 {
			delete(caps, kind)
		}
	}
	return caps, true
}

func rgwUserCapsMatch(raw []byte, uid string, expected map[string]uint8) bool {
	actual, ok := rgwUserCaps(raw, uid)
	if !ok || expected == nil || len(actual) != len(expected) {
		return false
	}
	for kind, bits := range expected {
		if value, exists := actual[kind]; !exists || value != bits {
			return false
		}
	}
	return true
}
