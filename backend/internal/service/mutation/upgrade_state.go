package mutation

import (
	"bytes"
	"encoding/json"
	"io"
)

func upgradeStartMatches(target map[string]any, data []byte) bool {
	if !upgradeControlMatches("resume", data) {
		return false
	}
	var state struct {
		TargetImage string `json:"target_image"`
	}
	if json.Unmarshal(data, &state) != nil || state.TargetImage == "" {
		return false
	}
	return state.TargetImage == target["target_name"] || state.TargetImage == target["target_digest"]
}

func upgradeControlMatches(action string, data []byte) bool {
	var state struct {
		InProgress *bool `json:"in_progress"`
		Paused     *bool `json:"is_paused"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if decoder.Decode(&state) != nil || decoder.Decode(new(any)) != io.EOF || state.InProgress == nil || state.Paused == nil {
		return false
	}
	switch action {
	case "pause":
		return *state.InProgress && *state.Paused
	case "resume":
		return *state.InProgress && !*state.Paused
	case "stop":
		return !*state.InProgress && !*state.Paused
	default:
		return false
	}
}
