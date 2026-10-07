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
	if patch, exists := changes["spec"]; exists {
		var proposed map[string]any
		var original map[string]json.RawMessage
		if currentType != "ingress" || json.Unmarshal(patch, &proposed) != nil || proposed == nil || json.Unmarshal(current["spec"], &original) != nil || original == nil {
			return nil, invalid("cannot safely merge ingress settings")
		}
		for key := range proposed {
			if key != "ssl" && key != "ssl_cert" && key != "ssl_key" && key != "virtual_ip" && key != "frontend_port" && key != "monitor_port" {
				return nil, invalid("unsupported ingress spec update")
			}
		}
		listener, err := ingressListenerParameters(proposed)
		if err != nil {
			return nil, err
		}
		if len(listener) > 0 {
			var keepaliveOnly bool
			if raw, exists := original["keepalive_only"]; exists && (json.Unmarshal(raw, &keepaliveOnly) != nil || string(raw) == "null") {
				return nil, invalid("invalid ingress mode")
			}
			if keepaliveOnly || (original["virtual_ips_list"] != nil && string(original["virtual_ips_list"]) != "null") {
				return nil, invalid("listener editing is not supported for keepalive-only or multiple-VIP services")
			}
		}
		validated, err := ingressTLSParameters(proposed, listener)
		if err != nil {
			return nil, err
		}
		if validated["ssl"] == false {
			delete(original, "ssl_cert")
			delete(original, "ssl_key")
		}
		for key, value := range validated {
			original[key], _ = json.Marshal(value)
		}
		current["spec"], _ = json.Marshal(original)
	}
	return json.Marshal(current)
}
