package security

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRedactCommonSecretForms(t *testing.T) {
	cases := []struct{ input, secret string }{
		{"password=quoted-secret", "quoted-secret"},
		{"https://user:password@example.test/api", "user:password"},
		{"authorization: Bearer-token", "Bearer-token"},
		{"eyJhbGciOiJIUzI1NiJ9.cGF5bG9hZA.c2lnbmF0dXJl", "eyJ"},
		{"client.test { key = AQExampleCephKey== }", "AQExampleCephKey"},
	}
	for _, test := range cases {
		output := Redact(test.input)
		if strings.Contains(output, test.secret) {
			t.Fatalf("secret %q leaked in %q", test.secret, output)
		}
	}
}

func TestProtectJSONEncryptsAndRestoresNestedSecrets(t *testing.T) {
	key := "0123456789abcdefghijklmnopqrstuv"
	input := map[string]any{
		"name":       "target",
		"chap":       map[string]any{"username": "client", "password": "plain-password"},
		"access_key": "plain-access",
	}
	protected, err := ProtectJSON(input, key)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(protected)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"plain-password", "plain-access"} {
		if bytes.Contains(encoded, []byte(secret)) {
			t.Fatalf("protected JSON contains %q: %s", secret, encoded)
		}
	}
	restored, err := UnprotectJSON(protected, key)
	if err != nil {
		t.Fatal(err)
	}
	restoredJSON, _ := json.Marshal(restored)
	if !bytes.Contains(restoredJSON, []byte("plain-password")) || !bytes.Contains(restoredJSON, []byte("plain-access")) {
		t.Fatalf("restored JSON = %s", restoredJSON)
	}
	if _, err := UnprotectJSON(protected, "abcdefghijklmnopqrstuvwxyz012345"); err == nil {
		t.Fatal("protected JSON decrypted with the wrong key")
	}
}

func FuzzRedact(f *testing.F) {
	for _, seed := range []string{"password=secret", "https://u:p@example.test", "client.a { key = AQabc== }", "plain log line", "token=\"quoted value\""} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		output := Redact(value)
		if len(output) > len(value)+64 {
			t.Fatalf("redactor expanded input unexpectedly")
		}
	})
}
