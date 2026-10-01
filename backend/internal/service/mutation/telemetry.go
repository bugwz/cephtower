package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func telemetryCommand(p map[string]any) (command, error) {
	enabled, ok := p["enabled"].(bool)
	if !ok {
		return command{}, invalid("enabled must be boolean")
	}
	args := []string{"telemetry", "off"}
	if enabled {
		if p["license"] != "sharing-1-0" {
			return command{}, invalid("enabling telemetry requires explicit sharing-1-0 license acceptance")
		}
		args = []string{"telemetry", "on", "--license", "sharing-1-0"}
	} else if _, present := p["license"]; present {
		return command{}, invalid("license acceptance must not accompany telemetry disable")
	}
	return command{binary: executor.BinaryCeph, args: args, check: []string{"telemetry", "status", "--format", "json"}, timeout: time.Minute}, nil
}

func telemetryStateMatches(data []byte, enabled bool) bool {
	var state struct {
		Enabled *bool `json:"enabled"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(&state) == nil && decoder.Decode(new(any)) == io.EOF && state.Enabled != nil && *state.Enabled == enabled
}

func telemetryChannelCommand(p map[string]any) (command, error) {
	channel, _ := p["channel"].(string)
	switch channel {
	case "basic", "ident", "crash", "device", "perf":
	default:
		return command{}, invalid("unknown telemetry channel")
	}
	enabled, ok := p["enabled"].(bool)
	if !ok {
		return command{}, invalid("enabled must be boolean")
	}
	action := "disable"
	if enabled {
		action = "enable"
	}
	return command{binary: executor.BinaryCeph, args: []string{"telemetry", action, "channel", channel}, check: []string{"telemetry", "status", "--format", "json"}, timeout: time.Minute}, nil
}

func telemetryChannelMatches(data []byte, p map[string]any) bool {
	var state map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(data))
	if decoder.Decode(&state) != nil || decoder.Decode(new(any)) != io.EOF {
		return false
	}
	var active, enabled *bool
	channel, _ := p["channel"].(string)
	return json.Unmarshal(state["enabled"], &active) == nil && active != nil && *active &&
		json.Unmarshal(state["channel_"+channel], &enabled) == nil && enabled != nil && *enabled == p["enabled"]
}
