package mutation

import (
	"reflect"
	"strings"
	"testing"
)

func encryptionPatch(encryptionType, provider string, values map[string]any) map[string]any {
	return map[string]any{"entity": "client.rgw.a", "encryption_type": encryptionType, "provider": provider, "expected_backend": "barbican", "values": values, "confirm_disruption": true, "confirm_credentials_saved": true}
}

func TestRGWEncryptionPatchCommands(t *testing.T) {
	for _, profile := range [][2]string{{"kms", "vault"}, {"kms", "kmip"}, {"s3", "vault"}} {
		values := map[string]any{"addr": "https://vault.example:8200", "auth": "agent", "prefix": " /v1/transit/ ", "secret_engine": "transit", "namespace": "", "token_file": "/etc/rgw/token", "ssl_cacert": "/etc/ca.pem", "ssl_clientcert": "/etc/client.pem", "ssl_clientkey": "/etc/key.pem", "verify_ssl": false}
		if profile[1] == "kmip" {
			values = map[string]any{"addr": "kmip.example:5696", "username": "", "password": " -private-password ", "client_cert": "/etc/client.pem", "client_key": "/etc/key.pem", "ca_path": "/etc/ca", "kms_key_template": "$keyid", "s3_key_template": "--literal"}
		}
		if profile[0] == "s3" {
			values["key_template"] = "%bucket_id"
		}
		p := encryptionPatch(profile[0], profile[1], values)
		plan, err := planRGWEncryption(p)
		if err != nil {
			t.Fatal(err)
		}
		if len(plan.changes) != len(values) || plan.expectedBackend != "barbican" || plan.entity != "client.rgw.a" {
			t.Fatalf("invalid plan: %+v", plan)
		}
		prefix, backend := "rgw_crypt_"+profile[1]+"_", "rgw_crypt_s3_kms_backend"
		if profile[0] == "s3" {
			prefix = "rgw_crypt_sse_s3_vault_"
			backend = "rgw_crypt_sse_s3_backend"
		}
		if plan.backendKey != backend {
			t.Fatal("wrong backend key")
		}
		for i, change := range plan.changes {
			if i > 0 && plan.changes[i-1].field >= change.field {
				t.Fatal("nondeterministic field order")
			}
			option := prefix + change.field
			if change.field == "key_template" {
				option = "rgw_crypt_sse_s3_key_template"
			}
			want := []string{"config", "set", "client.rgw.a", option, "--value=" + change.value}
			if change.value == "" {
				want = []string{"--", "config", "set", "client.rgw.a", option, "--value", ""}
			}
			if change.option != option || !reflect.DeepEqual(change.command.args, want) {
				t.Fatalf("wrong command: %v", change.command.args)
			}
			if _, ok := change.command.sensitive[len(want)-1]; !ok {
				t.Fatal("value not marked sensitive")
			}
			if change.field == "verify_ssl" && change.value != "false" {
				t.Fatal("false lost")
			}
			if change.field == "prefix" && change.value != " /v1/transit/ " {
				t.Fatal("spaces trimmed")
			}
		}
	}
	p := encryptionPatch("kms", "vault", map[string]any{"prefix": ""})
	plan, err := planRGWEncryption(p)
	if err != nil || len(plan.changes) != 1 || plan.changes[0].field != "prefix" {
		t.Fatal("partial patch expands unspecified fields")
	}
	if _, err := planRGWEncryption(encryptionPatch("kms", "vault", map[string]any{"secret_engine": "kv-v2"})); err != nil {
		t.Fatal(err)
	}
}

func TestRGWEncryptionPatchRejectsInvalidValues(t *testing.T) {
	for _, values := range []map[string]any{nil, {}, {"backend": "vault"}, {"unique_id": "id"}, {"encryption_type": "s3"}, {"unknown": "x"}, {"verify_ssl": "false"}, {"auth": ""}, {"auth": "password"}, {"secret_engine": "other"}, {"prefix": true}, {"prefix": "x\n"}, {"prefix": string([]byte{0xff})}, {"prefix": strings.Repeat("x", 4097)}, {"password": "secret"}, {"addr": "https://user:secret@vault.example"}, {"addr": "https://vault.example?secret=x"}, {"addr": "https://vault.example:0"}, {"addr": "https://vault.example:65536"}, {"addr": "https://vault.example:"}, {"addr": "file:///secret"}, {"addr": "https://vault.example\\x"}, {"token_file": "[REDACTED]"}} {
		if _, err := planRGWEncryption(encryptionPatch("kms", "vault", values)); err == nil {
			t.Fatalf("invalid values accepted: keys=%v", reflect.ValueOf(values).MapKeys())
		}
	}
	for _, value := range []string{"[::1]:5696", "host:0", "host:65536", "host:", ":5696", "https://host", "host?secret=x", "host:05696"} {
		if _, err := planRGWEncryption(encryptionPatch("kms", "kmip", map[string]any{"addr": value})); err == nil {
			t.Fatal("unsupported KMIP address accepted")
		}
	}
	for _, edit := range []map[string]any{{"entity": "client.admin"}, {"entity": "client.rgw.a/x"}, {"provider": "unknown"}, {"encryption_type": "other"}, {"expected_backend": ""}, {"expected_backend": "a\x00"}, {"confirm_disruption": false}, {"confirm_disruption": "true"}, {"confirm_credentials_saved": "true"}, {"extra": "x"}} {
		p := encryptionPatch("kms", "vault", map[string]any{"prefix": ""})
		for key, value := range edit {
			p[key] = value
		}
		if _, err := planRGWEncryption(p); err == nil {
			t.Fatal("invalid scope accepted")
		}
	}
	for _, p := range []map[string]any{encryptionPatch("s3", "kmip", map[string]any{"addr": "host"}), encryptionPatch("s3", "vault", map[string]any{"secret_engine": "kv-v2"})} {
		if _, err := planRGWEncryption(p); err == nil {
			t.Fatal("unsupported profile accepted")
		}
	}
	p := encryptionPatch("kms", "kmip", map[string]any{"password": "private-password"})
	delete(p, "confirm_credentials_saved")
	if _, err := planRGWEncryption(p); err == nil || strings.Contains(err.Error(), "private-password") {
		t.Fatal("unconfirmed password accepted or leaked")
	}
}
