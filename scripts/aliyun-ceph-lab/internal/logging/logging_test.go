package logging

import (
	"testing"
	"time"
)

func TestFormatTimestampIncludesCurrentOffset(t *testing.T) {
	zone := time.FixedZone("test-local", 8*60*60)
	value := time.Date(2026, 7, 26, 18, 42, 10, 0, zone)
	if got, want := formatTimestamp(value), "2026-07-26T18:42:10+08:00"; got != want {
		t.Fatalf("formatTimestamp() = %q, want %q", got, want)
	}
}

func TestFormatLineIncludesLevelAndMessage(t *testing.T) {
	zone := time.FixedZone("test-local", 8*60*60)
	value := time.Date(2026, 7, 26, 18, 42, 10, 0, zone)
	want := "[2026-07-26T18:42:10+08:00] WARN instance is still starting\n"
	if got := formatLine(value, "WARN", "instance is still starting"); got != want {
		t.Fatalf("formatLine() = %q, want %q", got, want)
	}
}
