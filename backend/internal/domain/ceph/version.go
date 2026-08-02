package ceph

import (
	"regexp"
	"strings"
)

var versionPattern = regexp.MustCompile(`(?i)\b(\d+\.\d+\.\d+)\b(?:\s+\(([0-9a-f]{7,40})\))?`)

func NormalizeVersion(value string) string {
	match := versionPattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(match) == 0 {
		return strings.TrimSpace(value)
	}
	if match[2] == "" {
		return match[1]
	}
	return match[1] + " (" + match[2] + ")"
}

func IsVersion(value string) bool {
	return versionPattern.MatchString(strings.TrimSpace(value))
}

func VersionHasCommit(value string) bool {
	match := versionPattern.FindStringSubmatch(strings.TrimSpace(value))
	return len(match) > 2 && match[2] != ""
}

func NormalizeVersionPointer(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := NormalizeVersion(*value)
	return &normalized
}
