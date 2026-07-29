package task

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerRunsTaskAndStops(t *testing.T) {
	scheduler := NewScheduler()
	defer scheduler.Stop()

	var runs atomic.Int32
	done := make(chan struct{})
	err := scheduler.Register("test", time.Now(), Every(time.Hour), func(context.Context) {
		runs.Add(1)
		close(done)
	})
	if err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduled task did not run")
	}
	if runs.Load() != 1 {
		t.Fatalf("task runs = %d, want 1", runs.Load())
	}
}

func TestSchedulerDoesNotRunSameTaskConcurrently(t *testing.T) {
	scheduler := NewScheduler()
	defer scheduler.Stop()

	started := make(chan struct{})
	release := make(chan struct{})
	var startedOnce sync.Once
	var running atomic.Int32
	var maxRunning atomic.Int32
	err := scheduler.Register("test", time.Now(), Every(10*time.Millisecond), func(context.Context) {
		current := running.Add(1)
		defer running.Add(-1)
		for {
			previous := maxRunning.Load()
			if current <= previous || maxRunning.CompareAndSwap(previous, current) {
				break
			}
		}
		startedOnce.Do(func() {
			close(started)
		})
		<-release
	})
	if err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("scheduled task did not start")
	}
	time.Sleep(20 * time.Millisecond)
	if maxRunning.Load() != 1 {
		t.Fatalf("max concurrent task runs = %d, want 1", maxRunning.Load())
	}
	close(release)
}

func TestDailyAtSchedulesOneAM(t *testing.T) {
	location := time.FixedZone("CST", 8*60*60)
	start := time.Date(2026, 3, 5, 5, 0, 0, 0, location)
	want := time.Date(2026, 3, 6, 1, 0, 0, 0, location)
	if got := DailyAt(1)(start); !got.Equal(want) {
		t.Fatalf("DailyAt() = %s, want %s", got, want)
	}
}
