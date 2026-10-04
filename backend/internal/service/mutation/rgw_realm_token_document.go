package mutation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/url"
	"unicode/utf8"
)

type rgwRealmTokenDocument struct {
	RealmID, RealmName, Endpoint, AccessKey, Secret string
}

// Parse the reference RealmToken's five string fields without JSON's usual
// duplicate-key or invalid-UTF-8 normalization. Errors never include the input.
func parseRGWRealmToken(encoded string) (rgwRealmTokenDocument, error) {
	failure := func() (rgwRealmTokenDocument, error) {
		return rgwRealmTokenDocument{}, invalid("invalid realm bootstrap token document")
	}
	if encoded == "" || len(encoded) > 64<<10 {
		return failure()
	}
	raw, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil {
		return failure()
	}
	defer clear(raw)
	if !utf8.Valid(raw) || base64.StdEncoding.EncodeToString(raw) != encoded {
		return failure()
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	start, err := d.Token()
	if err != nil || start != json.Delim('{') {
		return failure()
	}
	fields := map[string]*string{}
	var token rgwRealmTokenDocument
	fields["realm_id"], fields["realm_name"], fields["endpoint"], fields["access_key"], fields["secret"] = &token.RealmID, &token.RealmName, &token.Endpoint, &token.AccessKey, &token.Secret
	seen := map[string]bool{}
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return failure()
		}
		name, ok := key.(string)
		target := fields[name]
		if !ok || target == nil || seen[name] {
			return failure()
		}
		seen[name] = true
		value, err := d.Token()
		if err != nil {
			return failure()
		}
		text, ok := value.(string)
		if !ok || text == "" {
			return failure()
		}
		*target = text
	}
	end, err := d.Token()
	if err != nil || end != json.Delim('}') || len(seen) != len(fields) {
		return failure()
	}
	if _, err = d.Token(); err != io.EOF {
		return failure()
	}
	if !syncFlowToken(token.RealmID) || !syncFlowToken(token.RealmName) {
		return failure()
	}
	endpoint, err := url.Parse(token.Endpoint)
	if err != nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.Hostname() == "" || endpoint.User != nil || endpoint.Fragment != "" || endpoint.Opaque != "" {
		return failure()
	}
	return token, nil
}
