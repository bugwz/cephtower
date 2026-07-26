package task

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"cephtower/backend/internal/logging"
)

var (
	ErrManagerStopped = errors.New("task manager is stopped")
	ErrTaskRunning    = errors.New("task is already running")
	ErrTaskCapacity   = errors.New("task concurrency limit reached")
)

type Job func(context.Context) error

type Manager struct {
	scheduler *Scheduler
	ctx       context.Context
	cancel    context.CancelFunc

	mu            sync.Mutex
	stopped       bool
	running       map[string]struct{}
	maxConcurrent int
	wg            sync.WaitGroup
	stopOnce      sync.Once
}

func NewManager(maxConcurrent int) *Manager {
	if maxConcurrent <= 0 {
		maxConcurrent = 4
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		scheduler:     NewScheduler(),
		ctx:           ctx,
		cancel:        cancel,
		running:       make(map[string]struct{}),
		maxConcurrent: maxConcurrent,
	}
}

func (m *Manager) Register(name string, firstAt time.Time, schedule Schedule, job Job) error {
	if job == nil {
		return fmt.Errorf("task %q handler is required", name)
	}
	return m.scheduler.Register(name, firstAt, schedule, func(context.Context) {
		if err := m.Submit(name, job); err != nil && !errors.Is(err, ErrTaskRunning) {
			logging.Warnf("scheduled task submission failed: task_name=%q error=%v", name, err)
		}
	})
}

func (m *Manager) Submit(name string, job Job) error {
	if name == "" {
		return fmt.Errorf("task name is required")
	}
	if job == nil {
		return fmt.Errorf("task %q handler is required", name)
	}

	m.mu.Lock()
	if m.stopped {
		m.mu.Unlock()
		return ErrManagerStopped
	}
	if _, exists := m.running[name]; exists {
		m.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrTaskRunning, name)
	}
	if len(m.running) >= m.maxConcurrent {
		m.mu.Unlock()
		return ErrTaskCapacity
	}
	m.running[name] = struct{}{}
	m.wg.Add(1)
	m.mu.Unlock()

	go m.run(name, job)
	return nil
}

func (m *Manager) Stop() {
	m.stopOnce.Do(func() {
		m.mu.Lock()
		m.stopped = true
		m.mu.Unlock()
		m.cancel()
		m.scheduler.Stop()
		m.wg.Wait()
	})
}

func (m *Manager) run(name string, job Job) {
	defer m.wg.Done()
	startedAt := time.Now()
	defer func() {
		if recovered := recover(); recovered != nil {
			logging.Errorf(
				"task panicked: task_name=%q duration_ms=%d panic=%v",
				name,
				time.Since(startedAt).Milliseconds(),
				recovered,
			)
		}
		m.mu.Lock()
		delete(m.running, name)
		m.mu.Unlock()
	}()

	logging.Debugf("task started: task_name=%q", name)
	if err := job(m.ctx); err != nil {
		durationMS := time.Since(startedAt).Milliseconds()
		if errors.Is(err, context.Canceled) && m.ctx.Err() != nil {
			logging.Debugf(
				"task canceled during shutdown: task_name=%q duration_ms=%d error=%v",
				name, durationMS, err,
			)
		} else {
			logging.Warnf(
				"task failed: task_name=%q duration_ms=%d error=%v",
				name, durationMS, err,
			)
		}
		return
	}
	logging.Debugf(
		"task finished: task_name=%q duration_ms=%d",
		name, time.Since(startedAt).Milliseconds(),
	)
}
