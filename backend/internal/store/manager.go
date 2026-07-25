package store

import (
	"sync"
)

// Manager provides synchronized access to the active database. Database
// replacement is required by the first-run setup flow.
type Manager struct {
	mu sync.RWMutex
	db *Database
}

func NewManager(db *Database) *Manager {
	return &Manager{db: db}
}

func (m *Manager) Current() *Database {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db
}

func (m *Manager) Replace(db *Database) *Database {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	previous := m.db
	m.db = db
	return previous
}

func (m *Manager) Close() error {
	return Close(m.Current())
}
