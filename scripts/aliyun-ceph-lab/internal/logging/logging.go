package logging

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var outputMu sync.Mutex

// Infof writes one timestamped log line to standard output. time.Now uses the
// process's current local timezone, including any TZ environment override.
func Infof(format string, args ...any) {
	line := fmt.Sprintf("[%s] %s\n", formatTimestamp(time.Now()), fmt.Sprintf(format, args...))
	outputMu.Lock()
	defer outputMu.Unlock()
	// os.Stdout is written directly without a buffered writer. Keep each log as
	// one Write call, then request an immediate best-effort flush for regular
	// files. Sync can be unsupported by terminals and pipes, where Write itself
	// already hands the line to the operating system immediately.
	_, _ = os.Stdout.Write([]byte(line))
	_ = os.Stdout.Sync()
}

func formatTimestamp(value time.Time) string { return value.Format(time.RFC3339) }
