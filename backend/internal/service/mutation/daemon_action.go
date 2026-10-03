package mutation

import (
	"encoding/json"
	"regexp"
	"strings"
)

// A matching inventory record confirms visibility, not asynchronous action completion.
func daemonActionTargetVisible(output []byte, name string) bool {
	var rows []struct {
		Name string `json:"daemon_name"`
		Type string `json:"daemon_type"`
		ID   string `json:"daemon_id"`
	}
	if json.Unmarshal(output, &rows) != nil || len(rows) != 1 {
		return false
	}
	return rows[0].Name == name && rows[0].Type+"."+rows[0].ID == name
}

// Cephadm acknowledges scheduling, not completion of the daemon lifecycle change.
func daemonActionScheduled(output []byte, name, action string) bool {
	prefix := "Scheduled to " + action + " " + name + " on host '"
	return regexp.MustCompile(`^` + regexp.QuoteMeta(prefix) + `[^'\s]+'$`).MatchString(strings.TrimSpace(string(output)))
}
