package mutation

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestRGWRealmTokenDocument(t *testing.T) {
	base := map[string]any{"realm_id": "id", "realm_name": "realm", "endpoint": "https://rgw.example:443/path", "access_key": "access", "secret": "private-secret"}
	encode := func(v any) string {
		b, e := json.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		return base64.StdEncoding.EncodeToString(b)
	}
	for _, endpoint := range []string{"http://rgw.example:80", "https://rgw.example/path", "https://[2001:db8::1]:443"} {
		base["endpoint"] = endpoint
		token, err := parseRGWRealmToken(encode(base))
		if err != nil || token.Endpoint != endpoint || token.Secret != "private-secret" || token.RealmID != "id" || token.RealmName != "realm" {
			t.Fatal("valid native fields were not preserved")
		}
	}
	base["endpoint"] = "https://rgw.example"
	for _, key := range []string{"realm_id", "realm_name", "endpoint", "access_key", "secret"} {
		for _, value := range []any{nil, "", false, 12, []string{"private-secret"}} {
			copy := map[string]any{}
			for k, v := range base {
				copy[k] = v
			}
			copy[key] = value
			if _, err := parseRGWRealmToken(encode(copy)); err == nil || strings.Contains(err.Error(), "private-secret") {
				t.Fatalf("invalid %s accepted or leaked", key)
			}
		}
		copy := map[string]any{}
		for k, v := range base {
			if k != key {
				copy[k] = v
			}
		}
		if _, err := parseRGWRealmToken(encode(copy)); err == nil {
			t.Fatalf("missing %s accepted", key)
		}
	}
	for _, endpoint := range []string{"/relative", "rgw.example", "file:///secret", "https://u:private-secret@rgw.example", "https://rgw.example/#fragment", "https://rgw.example:bad", "https://"} {
		base["endpoint"] = endpoint
		if _, err := parseRGWRealmToken(encode(base)); err == nil || strings.Contains(err.Error(), "private-secret") {
			t.Fatal("invalid endpoint accepted or leaked")
		}
	}
	base["endpoint"] = "https://rgw.example"
	raw, _ := json.Marshal(base)
	for _, tail := range []string{`,"realm_id":"other"}`, `,"secret":"replacement"}`, `,"unknown":"private-secret"}`} {
		bad := base64.StdEncoding.EncodeToString(append(append([]byte{}, raw[:len(raw)-1]...), []byte(tail)...))
		if _, err := parseRGWRealmToken(bad); err == nil {
			t.Fatal("ambiguous or unknown field accepted")
		}
		response, _ := json.Marshal([]map[string]string{{"realm": "realm", "token": bad}})
		if _, err := selectRGWRealmToken(response, "id", "realm"); err == nil {
			t.Fatal("invalid document exported")
		}
	}
	for _, bad := range []string{"", encode(base) + "\n", strings.Repeat("A", 65537), base64.StdEncoding.EncodeToString(append(raw, raw...)), base64.StdEncoding.EncodeToString([]byte{0xff}), encode([]any{base})} {
		if _, err := parseRGWRealmToken(bad); err == nil {
			t.Fatal("invalid encoding accepted")
		}
	}
}
