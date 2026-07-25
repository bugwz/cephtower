package task

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestManagerRejectsDuplicateRunningTask(t *testing.T) {
	manager := NewManager(2)
	defer manager.Stop()

	started := make(chan struct{})
	release := make(chan struct{})
	if err := manager.Submit("collect", func(context.Context) error {
		close(started)
		<-release
		return nil
	}); err != nil {
		t.Fatalf("Submit() returned error: %v", err)
	}
	<-started
	if err := manager.Submit("collect", func(context.Context) error { return nil }); !errors.Is(err, ErrTaskRunning) {
		t.Fatalf("duplicate Submit() error = %v, want ErrTaskRunning", err)
	}
	close(release)
}

func TestManagerCancelsAndWaitsForTasks(t *testing.T) {
	manager := NewManager(1)
	finished := make(chan struct{})
	if err := manager.Submit("wait", func(ctx context.Context) error {
		<-ctx.Done()
		close(finished)
		return ctx.Err()
	}); err != nil {
		t.Fatalf("Submit() returned error: %v", err)
	}

	manager.Stop()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("Stop() did not wait for canceled task")
	}
	if err := manager.Submit("late", func(context.Context) error { return nil }); !errors.Is(err, ErrManagerStopped) {
		t.Fatalf("Submit() after Stop error = %v, want ErrManagerStopped", err)
	}
}

func TestManagerRecoversPanicAndReleasesTaskKey(t *testing.T) {
	manager := NewManager(1)
	done := make(chan struct{})
	if err := manager.Submit("panic", func(context.Context) error {
		defer close(done)
		panic("boom")
	}); err != nil {
		t.Fatal(err)
	}
	<-done
	deadline := time.Now().Add(time.Second)
	for {
		err := manager.Submit("panic", func(context.Context) error { return nil })
		if err == nil {
			break
		}
		if !errors.Is(err, ErrTaskRunning) || time.Now().After(deadline) {
			t.Fatalf("task key was not released: %v", err)
		}
		time.Sleep(time.Millisecond)
	}
	manager.Stop()
}
