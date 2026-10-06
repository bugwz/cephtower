package mutation

import (
	"encoding/json"
	"slices"
	"strings"
)

func hostActionStateMatches(raw []byte, hostname, action string) bool {
	var hosts []struct {
		Hostname *string  `json:"hostname"`
		Status   *string  `json:"status"`
		Labels   []string `json:"labels"`
	}
	if json.Unmarshal(raw, &hosts) != nil || hosts == nil || hostname == "" {
		return false
	}
	seen := map[string]bool{}
	found := false
	for _, host := range hosts {
		if host.Hostname == nil || *host.Hostname == "" || strings.TrimSpace(*host.Hostname) != *host.Hostname || seen[*host.Hostname] {
			return false
		}
		seen[*host.Hostname] = true
		if *host.Hostname != hostname {
			continue
		}
		found = true
		switch action {
		case "maintenance_enter":
			if host.Status == nil || *host.Status != "maintenance" {
				return false
			}
		case "maintenance_exit":
			if host.Status == nil || *host.Status != "" {
				return false
			}
		case "drain", "stop_drain":
			if host.Labels == nil {
				return false
			}
			labels := map[string]bool{}
			for _, label := range host.Labels {
				if label == "" || strings.TrimSpace(label) != label || labels[label] {
					return false
				}
				labels[label] = true
			}
			for _, label := range []string{"_no_schedule", "_no_conf_keyring"} {
				if slices.Contains(host.Labels, label) != (action == "drain") {
					return false
				}
			}
		default:
			return false
		}
	}
	return found
}
