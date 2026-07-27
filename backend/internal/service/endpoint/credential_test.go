package endpoint

import "testing"

func TestValidateCredentialUsesKindSpecificWhitelist(t *testing.T) {
	if err := validateCredential("s3", map[string]any{"access_key": "access", "secret_key": "secret", "password": "forbidden"}); err == nil {
		t.Fatal("unknown credential field was accepted")
	}
	if err := validateCredential("s3", map[string]any{"access_key": "access"}); err == nil {
		t.Fatal("incomplete S3 credential was accepted")
	}
	if err := validateCredential("nvmeof", map[string]any{"client_certificate": "certificate"}); err == nil {
		t.Fatal("unpaired client certificate was accepted")
	}
	if err := validateCredential("alertmanager", map[string]any{"token": "token"}); err != nil {
		t.Fatal(err)
	}
}
