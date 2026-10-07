package security

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestTLSServiceMaterialProtection(t *testing.T) {
	for _, field := range []string{"ssl_key", "ssl_cert", "SSL_KEY", "ingress_ssl_cert"} {
		input := map[string]any{"spec": map[string]any{field: "tls-fixture-secret\nsecond-line", "ssl": true, "frontend_port": "443"}}
		protected, err := ProtectJSON(input, testKey)
		if err != nil {
			t.Fatal(err)
		}
		encoded, _ := json.Marshal(protected)
		if strings.Contains(string(encoded), "tls-fixture-secret") || !strings.Contains(string(encoded), "$encrypted") {
			t.Fatal("TLS material was not encrypted")
		}
		plain, err := UnprotectJSON(protected, testKey)
		if err != nil || !reflect.DeepEqual(plain, input) {
			t.Fatalf("TLS material did not round trip: %v", err)
		}
		visible, err := RedactJSON(input)
		if err != nil {
			t.Fatal(err)
		}
		data := visible.(map[string]any)["spec"].(map[string]any)
		if data[field] != "[REDACTED]" || data["ssl"] != true || data["frontend_port"] != "443" {
			t.Fatal("incorrect TLS redaction")
		}
	}
}

func TestTLSMaterialTextRedaction(t *testing.T) {
	for _, input := range []string{
		`{"ssl_key":"tls-fixture-secret\nsecond-line"}`, `malformed {"ssl_cert":"tls-fixture-secret"}`, "ssl_key=tls-fixture-secret",
		"error: -----BEGIN PRIVATE KEY-----\ntls-fixture-secret\n-----END PRIVATE KEY----- failure",
		"-----BEGIN RSA PRIVATE KEY-----\ntls-fixture-secret\n-----END RSA PRIVATE KEY-----",
		"-----BEGIN ENCRYPTED PRIVATE KEY-----\ntls-fixture-secret\n-----END ENCRYPTED PRIVATE KEY-----",
		"-----BEGIN OPENSSH PRIVATE KEY-----\ntls-fixture-secret",
	} {
		output := Redact(input)
		if strings.Contains(output, "tls-fixture-secret") || Redact(output) != output {
			t.Fatal("TLS text redaction failed")
		}
	}
	if Redact("ssl=true frontend_port=443") != "ssl=true frontend_port=443" {
		t.Fatal("non-secret TLS settings redacted")
	}
}
