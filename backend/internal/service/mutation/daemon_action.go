package mutation

import (
	"regexp"
	"strings"
)

// Cephadm acknowledges scheduling, not completion of the daemon lifecycle change.
func daemonActionScheduled(output []byte, name, action string) bool {
	prefix := "Scheduled to " + action + " " + name + " on host '"
	return regexp.MustCompile(`^` + regexp.QuoteMeta(prefix) + `[^'\s]+'$`).MatchString(strings.TrimSpace(string(output)))
}
