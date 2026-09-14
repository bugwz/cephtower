package ceph

import (
	"context"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

// module ls supplies release-specific activation state; mgr dump includes
// option schemas and load errors for enabled modules as well as disabled ones.
func (p *NativeProvider) collectManagerModules(ctx context.Context, access ClusterAccess, now time.Time) []Observation {
	var states, metadata map[string]any
	if !p.optional(ctx, access, executor.BinaryCeph, "collect.mgr_module", []string{"mgr", "module", "ls", "--format", "json"}, &states) {
		return nil
	}
	if !p.optional(ctx, access, executor.BinaryCeph, "collect.mgr_module_metadata", []string{"mgr", "dump", "--format", "json"}, &metadata) {
		return nil
	}
	for _, key := range []string{"enabled_modules", "always_on_modules", "force_disabled_modules", "disabled_modules"} {
		if _, ok := states[key].([]any); !ok {
			markCollectionUnavailable(ctx, "collect.mgr_module")
			return nil
		}
	}
	if _, ok := metadata["available_modules"].([]any); !ok {
		markCollectionUnavailable(ctx, "collect.mgr_module_metadata")
		return nil
	}
	contains := func(key, name string) bool {
		for _, value := range stringList(states[key], "") {
			if value == name {
				return true
			}
		}
		return false
	}
	var rows []Observation
	for _, module := range objectList(metadata["available_modules"]) {
		name := textField(module, "name")
		if name == "" {
			markCollectionUnavailable(ctx, "collect.mgr_module_metadata")
			return nil
		}
		if name == "selftest" {
			continue
		}
		always := contains("always_on_modules", name)
		forced := contains("force_disabled_modules", name)
		module["enabled"] = (always || contains("enabled_modules", name)) && !forced
		module["always_on"] = always
		module["force_disabled"] = forced
		module["options"] = module["module_options"]
		rows = append(rows, observation("mgr_module", name, name, "ceph_cli", module, now))
	}
	return rows
}
