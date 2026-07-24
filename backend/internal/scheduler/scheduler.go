package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Schedule func(time.Time) time.Time
type Task func(context.Context)

var (
	systemTaskMu            sync.Mutex
	logRetentionCleanupTask Task
)

type Scheduler struct {
	mu       sync.Mutex
	tasks    map[string]*scheduledTask
	wake     chan struct{}
	stop     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

type scheduledTask struct {
	next     time.Time
	schedule Schedule
	task     Task
	running  bool
}

func New() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	scheduler := &Scheduler{
		tasks:  make(map[string]*scheduledTask),
		wake:   make(chan struct{}, 1),
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
		ctx:    ctx,
		cancel: cancel,
	}
	go scheduler.run()
	return scheduler
}

// Start creates the system scheduler and registers all recurring system tasks.
func RegisterLogRetentionCleanup(task Task) {
	systemTaskMu.Lock()
	defer systemTaskMu.Unlock()
	logRetentionCleanupTask = task
}

func Start() (*Scheduler, error) {
	scheduler := New()
	systemTaskMu.Lock()
	cleanupTask := logRetentionCleanupTask
	systemTaskMu.Unlock()
	if cleanupTask != nil {
		cleanupAt := nextDailyAt(1)
		if err := scheduler.Register(
			"log-retention-cleanup",
			cleanupAt(time.Now()),
			cleanupAt,
			cleanupTask,
		); err != nil {
			scheduler.Stop()
			return nil, err
		}
	}
	return scheduler, nil
}

func nextDailyAt(hour int) Schedule {
	return func(now time.Time) time.Time {
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
		if !next.After(now) {
			next = next.AddDate(0, 0, 1)
		}
		return next
	}
}

func (s *Scheduler) Register(name string, firstAt time.Time, schedule Schedule, task Task) error {
	if name == "" {
		return fmt.Errorf("scheduler task name is required")
	}
	if schedule == nil {
		return fmt.Errorf("scheduler task %q schedule is required", name)
	}
	if task == nil {
		return fmt.Errorf("scheduler task %q handler is required", name)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-s.stop:
		return fmt.Errorf("scheduler is stopped")
	default:
	}
	if _, exists := s.tasks[name]; exists {
		return fmt.Errorf("scheduler task %q is already registered", name)
	}
	s.tasks[name] = &scheduledTask{next: firstAt, schedule: schedule, task: task}
	s.signal()
	return nil
}

func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() {
		s.cancel()
		close(s.stop)
	})
	<-s.done
	s.wg.Wait()
}

func (s *Scheduler) signal() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *Scheduler) run() {
	defer close(s.done)
	for {
		delay := s.nextDelay()
		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
			s.runDue(time.Now())
		case <-s.wake:
			stopTimer(timer)
		case <-s.stop:
			stopTimer(timer)
			return
		}
	}
}

func (s *Scheduler) nextDelay() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	next := now.Add(24 * time.Hour)
	for _, task := range s.tasks {
		if task.next.Before(next) {
			next = task.next
		}
	}
	delay := time.Until(next)
	if delay < 0 {
		return 0
	}
	return delay
}

func (s *Scheduler) runDue(now time.Time) {
	s.mu.Lock()
	for _, task := range s.tasks {
		if task.next.After(now) {
			continue
		}
		task.next = task.schedule(now)
		if !task.next.After(now) {
			task.next = now.Add(time.Millisecond)
		}
		if task.running {
			continue
		}
		task.running = true
		s.wg.Add(1)
		go s.runTask(task)
	}
	s.mu.Unlock()
}

func (s *Scheduler) runTask(task *scheduledTask) {
	defer s.wg.Done()
	task.task(s.ctx)
	s.mu.Lock()
	task.running = false
	s.mu.Unlock()
	s.signal()
}

func stopTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}
