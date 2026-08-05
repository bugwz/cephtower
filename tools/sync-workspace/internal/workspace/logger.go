package workspace

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type logger struct {
	output io.Writer
	mu     sync.Mutex
}

func newLogger(output io.Writer) *logger {
	return &logger{output: output}
}

func (l *logger) Infof(format string, args ...any) {
	l.write("INFO", format, args...)
}

func (l *logger) Errorf(format string, args ...any) {
	l.write("ERROR", format, args...)
}

func (l *logger) write(level, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.output, "[%s] %s %s\n", time.Now().Format(time.RFC3339), level, fmt.Sprintf(format, args...))
}
