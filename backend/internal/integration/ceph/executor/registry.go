package executor

import (
	"fmt"
	"sort"
	"sync"
)

type BuildInput struct{ Parameters map[string]any }
type Builder func(BuildInput) (CommandSpec, error)
type Registry struct {
	mu       sync.RWMutex
	builders map[string]Builder
}

func NewRegistry() *Registry { return &Registry{builders: map[string]Builder{}} }
func (r *Registry) Register(action string, builder Builder) error {
	if action == "" || builder == nil {
		return fmt.Errorf("action and builder are required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.builders[action]; exists {
		return fmt.Errorf("action %q is already registered", action)
	}
	r.builders[action] = builder
	return nil
}
func (r *Registry) Build(action string, input BuildInput) (CommandSpec, error) {
	r.mu.RLock()
	builder := r.builders[action]
	r.mu.RUnlock()
	if builder == nil {
		return CommandSpec{}, fmt.Errorf("unsupported action %q", action)
	}
	spec, err := builder(input)
	if err != nil {
		return CommandSpec{}, err
	}
	spec.ID = action
	return spec, nil
}
func (r *Registry) Actions() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	actions := make([]string, 0, len(r.builders))
	for action := range r.builders {
		actions = append(actions, action)
	}
	sort.Strings(actions)
	return actions
}
