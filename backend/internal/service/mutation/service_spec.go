package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

// Export excludes specs marked deleted even while their daemons are being removed.
func serviceSpecAbsent(output []byte) bool {
	if strings.TrimSpace(string(output)) == "No services reported" {
		return true
	}
	var specs []json.RawMessage
	return json.Unmarshal(output, &specs) == nil && specs != nil && len(specs) == 0
}

// Export returns ServiceSpec objects, not ServiceDescription status wrappers.
// RawMessage retains fields unknown to this application and exact numeric values.
func mergeServiceSpec(exported, patch []byte, name string) ([]byte, error) {
	var rows []map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(exported))
	if decoder.Decode(&rows) != nil || decoder.Decode(new(any)) != io.EOF || len(rows) != 1 || rows[0] == nil {
		return nil, invalid("current service spec could not be uniquely identified; no update was applied")
	}
	var changes map[string]json.RawMessage
	if json.Unmarshal(patch, &changes) != nil {
		return nil, invalid("invalid service patch")
	}
	current := rows[0]
	var currentName, currentType, currentID, requestedType, requestedID string
	if json.Unmarshal(current["service_name"], &currentName) != nil || currentName != name || json.Unmarshal(current["service_type"], &currentType) != nil {
		return nil, invalid("exported service identity does not match the update target")
	}
	if raw, exists := current["service_id"]; exists && json.Unmarshal(raw, &currentID) != nil {
		return nil, invalid("invalid exported service id")
	}
	if json.Unmarshal(changes["service_type"], &requestedType) != nil {
		return nil, invalid("invalid requested service type")
	}
	if raw, exists := changes["service_id"]; exists && json.Unmarshal(raw, &requestedID) != nil {
		return nil, invalid("invalid requested service id")
	}
	derivedName := currentType
	if currentID != "" {
		derivedName += "." + currentID
	}
	if derivedName != name || currentType != requestedType || currentID != requestedID {
		return nil, invalid("exported service identity is inconsistent")
	}
	if placement, exists := changes["placement"]; exists {
		current["placement"] = placement
	}
	if unmanaged, exists := changes["unmanaged"]; exists {
		current["unmanaged"] = unmanaged
	}
	if networks, exists := changes["networks"]; exists {
		current["networks"] = networks
	}
	return json.Marshal(current)
}
