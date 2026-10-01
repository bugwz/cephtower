package mutation

import (
	"encoding/base64"
	"encoding/json"
)

func smbShareCreateRequest(parameters map[string]any) Request {
	return Request{ResourceKey: "smb/share/" + base64.RawURLEncoding.EncodeToString([]byte(optional(parameters, "cluster")+"\x00"+optional(parameters, "name"))), Parameters: parameters}
}

func smbShareCreateJSON(parameters map[string]any, cluster, share, filesystem string) ([]byte, error) {
	p := make(map[string]any, len(parameters)+1)
	for key, value := range parameters {
		p[key] = value
	}
	if optional(p, "path") == "" {
		p["path"] = "/"
	}
	if value, exists := p["subvolume"]; exists && value == "" {
		return nil, invalid("subvolume must not be empty at creation")
	}
	base, _ := json.Marshal(map[string]any{"resource_type": "ceph.smb.share", "cluster_id": cluster, "share_id": share, "name": share, "intent": "present", "readonly": false, "browseable": true, "cephfs": map[string]any{"volume": filesystem, "path": "/", "provider": "samba-vfs"}})
	return smbShareUpdateJSON(base, smbShareCreateRequest(p))
}
