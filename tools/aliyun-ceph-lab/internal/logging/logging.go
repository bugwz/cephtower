package logging

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

var outputMu sync.Mutex

// Infof writes one timestamped informational log line to standard error.
func Infof(format string, args ...any) {
	writef("INFO", format, args...)
}

func Warnf(format string, args ...any) {
	writef("WARN", format, args...)
}

func Errorf(format string, args ...any) {
	writef("ERROR", format, args...)
}

func writef(level, format string, args ...any) {
	outputMu.Lock()
	defer outputMu.Unlock()
	_ = Writef(os.Stderr, level, format, args...)
	_ = os.Stderr.Sync()
}

func Writef(output io.Writer, level, format string, args ...any) error {
	line := formatLine(time.Now(), level, fmt.Sprintf(format, args...))
	_, err := io.WriteString(output, line)
	return err
}

func formatTimestamp(value time.Time) string { return value.Format(time.RFC3339) }

func formatLine(value time.Time, level, message string) string {
	return fmt.Sprintf("[%s] %s %s\n", formatTimestamp(value), level, message)
}
