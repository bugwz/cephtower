package mutation

import (
	"encoding/json"
	"strings"
)

func hostUpdateMatches(raw []byte, host string, parameters map[string]any) bool {
	var rows []struct {
		Hostname *string  `json:"hostname"`
		Address  *string  `json:"addr"`
		Labels   []string `json:"labels"`
	}
	if json.Unmarshal(raw, &rows) != nil || rows == nil || host == "" {
		return false
	}
	seen := map[string]bool{}
	found := false
	for _, row := range rows {
		if row.Hostname == nil || *row.Hostname == "" || strings.TrimSpace(*row.Hostname) != *row.Hostname || seen[*row.Hostname] {
			return false
		}
		seen[*row.Hostname] = true
		if *row.Hostname != host {
			continue
		}
		if address := optional(parameters, "address"); address != "" && (row.Address == nil || *row.Address != address) {
			return false
		}
		add, _ := stringSlice(parameters["labels_add"])
		remove, _ := stringSlice(parameters["labels_remove"])
		if len(add) == 0 && len(remove) == 0 {
			found = true
			continue
		}
		if row.Labels == nil {
			return false
		}
		labels := map[string]bool{}
		for _, label := range row.Labels {
			if label == "" || strings.TrimSpace(label) != label || labels[label] {
				return false
			}
			labels[label] = true
		}
		for _, field := range []string{"labels_add", "labels_remove"} {
			requested, valid := stringSlice(parameters[field])
			if parameters[field] != nil && !valid {
				return false
			}
			for _, label := range requested {
				if labels[label] != (field == "labels_add") {
					return false
				}
			}
		}
		found = true
	}
	return found
}
