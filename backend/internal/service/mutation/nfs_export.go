package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

func nfsExportPseudo(data []byte, cluster, exportID string) (string, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var exports []map[string]any
	if decoder.Decode(&exports) != nil {
		return "", invalid("invalid native NFS export list")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return "", invalid("invalid native NFS export list")
	}
	pseudo := ""
	for _, export := range exports {
		id, ok := export["export_id"].(json.Number)
		if !ok || id.String() != exportID {
			continue
		}
		path, ok := export["pseudo"].(string)
		if !ok || !strings.HasPrefix(path, "/") || strings.ContainsAny(path, "\x00\r\n") || pseudo != "" {
			return "", invalid("invalid or ambiguous NFS export identity")
		}
		if value, exists := export["cluster_id"]; exists && value != cluster {
			return "", invalid("NFS export cluster mismatch")
		}
		pseudo = path
	}
	if pseudo == "" {
		return "", invalid("NFS export no longer exists")
	}
	return pseudo, nil
}
