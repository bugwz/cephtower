package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

func nfsBucketOwner(data []byte, bucket, tenant string) (string, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var record map[string]any
	if decoder.Decode(&record) != nil || record == nil {
		return "", invalid("invalid RGW bucket owner response")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return "", invalid("invalid RGW bucket owner response")
	}
	if record["bucket"] != bucket || record["tenant"] != tenant {
		return "", invalid("RGW bucket identity mismatch")
	}
	owner, ok := record["owner"].(string)
	if !ok || !nfsRGWUserID.MatchString(owner) || strings.HasPrefix(owner, "-") {
		return "", invalid("RGW bucket owner is not a supported user identity")
	}
	ownerTenant := ""
	if prefix, _, found := strings.Cut(owner, "$"); found {
		ownerTenant = prefix
	}
	if ownerTenant != tenant {
		return "", invalid("RGW bucket owner tenant mismatch")
	}
	return owner, nil
}
