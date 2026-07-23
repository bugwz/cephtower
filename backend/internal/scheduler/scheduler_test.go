package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerRunsTaskAndStops(t *testing.T) {
	scheduler := New()
	defer scheduler.Stop()

	var runs atomic.Int32
	done := make(chan struct{})
	err := scheduler.Register("test", time.Now(), func(now time.Time) time.Time {
		return now.Add(time.Hour)
	}, func(ctx context.Context) {
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
	scheduler := New()
	defer scheduler.Stop()

	started := make(chan struct{})
	release := make(chan struct{})
	var running atomic.Int32
	var maxRunning atomic.Int32
	err := scheduler.Register("test", time.Now(), func(now time.Time) time.Time {
		return now.Add(10 * time.Millisecond)
	}, func(ctx context.Context) {
		current := running.Add(1)
		for {
			previous := maxRunning.Load()
			if current <= previous || maxRunning.CompareAndSwap(previous, current) {
				break
			}
		}
		close(started)
		<-release
		running.Add(-1)
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

func TestNextDailyAtSchedulesOneAM(t *testing.T) {
	location := time.FixedZone("CST", 8*60*60)
	start := time.Date(2026, 3, 5, 5, 0, 0, 0, location)
	want := time.Date(2026, 3, 6, 1, 0, 0, 0, location)
	if got := nextDailyAt(1)(start); !got.Equal(want) {
		t.Fatalf("nextDailyAt() = %s, want %s", got, want)
	}
}
