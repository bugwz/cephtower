package app

import (
	"context"
	"sync/atomic"
	"testing"

	"cephtower/backend/internal/task"
)

func TestCloseIsIdempotent(t *testing.T) {
	var logCloses atomic.Int32
	application := &App{
		tasks: task.NewManager(1),
		closeLog: func() error {
			logCloses.Add(1)
			return nil
		},
	}
	if err := application.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := application.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := logCloses.Load(); got != 1 {
		t.Fatalf("log close count = %d, want 1", got)
	}
}
