package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

func nfsClusterStateMatches(request Request, data []byte) bool {
	name := optional(request.Parameters, "name")
	if request.Action == "nfs_cluster.delete" {
		name = last(resourceTail(request.ResourceKey))
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
	if request.Action == "nfs_cluster.create" {
		return seen[name]
	}
	return request.Action == "nfs_cluster.delete" && !seen[name]
}
