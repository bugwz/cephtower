package ceph

import (
	"context"
	"strings"
	"time"
)

func (p *NativeProvider) collectCephFSClients(ctx context.Context, access ClusterAccess, filesystem string, now time.Time) []Observation {
	if strings.TrimSpace(filesystem) == "" {
		markCollectionUnavailable(ctx, "collect.cephfs_client")
		return nil
	}
	var payload any
	args := []string{"tell", "mds." + filesystem + ":0", "session", "ls", "--format", "json"}
	if err := p.runInto(ctx, access, "collect.cephfs_client", args, &payload); err != nil {
		return nil
	}
	sessions, valid := cephFSSessionList(payload)
	if !valid {
		markCollectionUnavailable(ctx, "collect.cephfs_client")
		return nil
	}
	rows := make([]Observation, 0, len(sessions))
	for _, session := range sessions {
		clientID := textField(session, "id")
		if clientID == "" {
			markCollectionUnavailable(ctx, "collect.cephfs_client")
			continue
		}
		details := make(map[string]any, len(session)+7)
		for key, value := range session {
			details[key] = value
		}
		details["fs"] = filesystem
		details["filesystem"] = filesystem
		details["client_id"] = clientID
		metadata, _ := session["client_metadata"].(map[string]any)
		details["type"] = "unknown"
		details["version"] = ""
		if metadata != nil {
			details["hostname"] = textField(metadata, "hostname")
			details["root"] = textField(metadata, "root")
			switch {
			case textField(metadata, "ceph_version") != "":
				details["type"] = "userspace"
				details["version"] = textField(metadata, "ceph_version")
			case textField(metadata, "kernel_version") != "":
				details["type"] = "kernel"
				details["version"] = textField(metadata, "kernel_version")
			}
		}
		state := strings.TrimSpace(textField(session, "state"))
		rows = append(rows, Observation{
			Kind:       "cephfs_client",
			NaturalKey: filesystem + "/" + clientID,
			ParentKind: "filesystem",
			ParentKey:  filesystem,
			Name:       clientID,
			Status:     state,
			Source:     "ceph_cli",
			Payload:    details,
			ObservedAt: now,
		})
	}
	return rows
}

func cephFSSessionList(payload any) ([]map[string]any, bool) {
	switch value := payload.(type) {
	case []any:
		return cephFSSessionItems(value)
	case map[string]any:
		sessions, exists := value["sessions"]
		if !exists {
			return nil, false
		}
		items, ok := sessions.([]any)
		if !ok {
			return nil, false
		}
		return cephFSSessionItems(items)
	default:
		return nil, false
	}
}

func cephFSSessionItems(items []any) ([]map[string]any, bool) {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		session, ok := item.(map[string]any)
		if !ok {
			return nil, false
		}
		result = append(result, session)
	}
	return result, true
}
