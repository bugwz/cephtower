package workspace

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestLoggerUsesDeployToMachineFormat(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	newLogger(&output).Infof("service: started pid=%d", 123)
	pattern := regexp.MustCompile(`^\[[^]]+\] INFO service: started pid=123\n$`)
	if !pattern.MatchString(output.String()) {
		t.Fatalf("unexpected log format %q", output.String())
	}
}

func TestLogSynchronizationResultListsFiles(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	logSynchronizationResult(newLogger(&output), synchronizationResult{
		Uploaded: []string{"backend/cmd/main.go"},
		Deleted:  []string{"frontend/src/old.ts"},
	})
	for _, want := range []string{
		"INFO sync: uploaded backend/cmd/main.go -> /root/cephtower/backend/cmd/main.go",
		"INFO sync: deleted /root/cephtower/frontend/src/old.ts",
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("log output missing %q:\n%s", want, output.String())
		}
	}
}
