package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

// MgrMonitor's module ls separates always-on modules from explicitly enabled
// modules, and retains force-disabled modules in the always-on list.
func managerModuleStateMatches(data []byte, name string, enabled bool) bool {
	var state struct {
		Always   []string `json:"always_on_modules"`
		Enabled  []string `json:"enabled_modules"`
		Forced   []string `json:"force_disabled_modules"`
		Disabled []struct {
			Name string `json:"name"`
		} `json:"disabled_modules"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if decoder.Decode(&state) != nil || decoder.Decode(new(any)) != io.EOF || state.Always == nil || state.Enabled == nil || state.Forced == nil || state.Disabled == nil {
		return false
	}
	sets := make([]map[string]bool, 4)
	disabled := make([]string, 0, len(state.Disabled))
	for _, module := range state.Disabled {
		disabled = append(disabled, module.Name)
	}
	for index, list := range [][]string{state.Always, state.Enabled, state.Forced, disabled} {
		sets[index] = map[string]bool{}
		for _, module := range list {
			if module == "" || strings.TrimSpace(module) != module || sets[index][module] {
				return false
			}
			sets[index][module] = true
		}
	}
	always, explicit, forced, inactive := sets[0], sets[1], sets[2], sets[3]
	for module := range explicit {
		if always[module] || inactive[module] {
			return false
		}
	}
	for module := range inactive {
		if always[module] {
			return false
		}
	}
	for module := range forced {
		if !always[module] {
			return false
		}
	}
	if !always[name] && !explicit[name] && !inactive[name] {
		return false
	}
	return ((always[name] || explicit[name]) && !forced[name]) == enabled
}
