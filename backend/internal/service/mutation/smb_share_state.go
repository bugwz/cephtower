package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

func smbShareStateMatches(request Request, data []byte) bool {
	name := optional(request.Parameters, "name")
	switch request.Action {
	case "smb_share.create":
	case "smb_share.delete":
		_, share, err := decodePair(last(resourceTail(request.ResourceKey)))
		if err != nil {
			return false
		}
		name = share
	default:
		return false
	}
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
	return seen[name] == (request.Action == "smb_share.create")
}
