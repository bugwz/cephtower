package mutation

import (
	"encoding/json"
	"strings"
)

func hostRemovalConfirmed(raw []byte, hostname string) bool {
	if hostname == "" || strings.TrimSpace(hostname) != hostname {
		return false
	}
	var hosts []struct {
		Hostname *string `json:"hostname"`
	}
	if json.Unmarshal(raw, &hosts) != nil || hosts == nil {
		return false
	}
	seen := map[string]bool{}
	for _, host := range hosts {
		if host.Hostname == nil {
			return false
		}
		name := *host.Hostname
		if name == "" || strings.TrimSpace(name) != name || seen[name] || name == hostname {
			return false
		}
		seen[name] = true
	}
	return true
}
