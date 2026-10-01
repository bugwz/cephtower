package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

func smbShareStateMatches(request Request, data []byte) bool {
	if request.Action != "smb_share.delete" {
		return false
	}
	_, share, err := decodePair(last(resourceTail(request.ResourceKey)))
	return err == nil && smbResourcePresenceMatches(data, share, false)
}

func smbClusterStateMatches(request Request, data []byte) bool {
	switch request.Action {
	case "smb_cluster.create":
		return smbResourcePresenceMatches(data, optional(request.Parameters, "name"), true)
	case "smb_cluster.delete":
		return smbResourcePresenceMatches(data, last(resourceTail(request.ResourceKey)), false)
	default:
		return false
	}
}

func smbResourcePresenceMatches(data []byte, name string, present bool) bool {
	if name == "" {
		return false
	}
	var names []string
	decoder := json.NewDecoder(bytes.NewReader(data))
	if decoder.Decode(&names) != nil || names == nil {
		return false
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return false
	}
	seen := map[string]bool{}
	for _, entry := range names {
		if strings.TrimSpace(entry) == "" || seen[entry] {
			return false
		}
		seen[entry] = true
	}
	return seen[name] == present
}
