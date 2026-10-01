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
