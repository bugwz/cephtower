package security

import (
	"encoding/json"
	"strings"
	"testing"
)

const testKey = "0123456789abcdefghijklmnopqrstuv"

func TestSecretboxRoundTripAndRandomNonce(t *testing.T) {
	one, err := Encrypt([]byte("sensitive-value"), testKey)
	if err != nil {
		t.Fatal(err)
	}
	two, err := Encrypt([]byte("sensitive-value"), testKey)
	if err != nil {
		t.Fatal(err)
	}
	if one == two || strings.Contains(one, "=") {
		t.Fatalf("non-random or padded ciphertext: %q", one)
	}
	plain, err := Decrypt(one, testKey)
	if err != nil || string(plain) != "sensitive-value" {
		t.Fatalf("Decrypt() = %q, %v", plain, err)
	}
}

func TestRedactJSONPreservesStructure(t *testing.T) {
	value, err := RedactJSON(map[string]any{
		"password": "quoted-\"secret",
		"nested":   map[string]any{"access_token": "token-value", "name": "safe"},
		"items":    []any{map[string]any{"client_key": "ceph-key", "count": 2}},
	})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	for _, secret := range []string{"quoted", "token-value", "ceph-key"} {
		if strings.Contains(text, secret) {
			t.Fatalf("secret %q leaked in %s", secret, text)
		}
	}
	if !strings.Contains(text, `"name":"safe"`) || !strings.Contains(text, `"count":2`) {
		t.Fatalf("non-secret structure was lost: %s", text)
	}
}

func TestSecretboxRejectsInvalidInputs(t *testing.T) {
	if _, err := Encrypt(nil, testKey); err == nil {
		t.Fatal("Encrypt accepted empty plaintext")
	}
	value, err := Encrypt([]byte("secret"), testKey)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ value, key string }{
		{"", testKey}, {value + "=", testKey}, {"not+base64", testKey},
		{value[:len(value)-2], testKey}, {value, "bad-key"},
		{value, "1123456789abcdefghijklmnopqrstuv"},
	}
	for _, tc := range cases {
		if _, err := Decrypt(tc.value, tc.key); err == nil {
			t.Fatalf("Decrypt(%q) unexpectedly succeeded", tc.value)
		}
	}
}
