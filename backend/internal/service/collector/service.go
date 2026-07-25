package collector

import (
	"errors"
	"sync"

	"cephtower/backend/internal/store"
)

var ErrModuleRunning = errors.New("collector module is already running")

type Service struct {
	database func() *store.Database
	workDir  string
	runs     *runRegistry
}
type runRegistry struct {
	mu     sync.Mutex
	active map[string]struct{}
}

func New(database func() *store.Database, workDirs ...string) *Service {
	service := NewService(database, workDirs...)
	return &service
}

func NewService(database func() *store.Database, workDirs ...string) Service {
	workDir := ""
	if len(workDirs) > 0 {
		workDir = workDirs[0]
	}
	return Service{database: database, workDir: workDir, runs: &runRegistry{active: map[string]struct{}{}}}
}
