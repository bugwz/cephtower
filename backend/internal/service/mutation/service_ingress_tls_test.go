package mutation

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"
)

func ingressTestKeyPair(t *testing.T) (string, string) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, template, template, public, private)
	if err != nil {
		t.Fatal(err)
	}
	key, err := x509.MarshalPKCS8PrivateKey(private)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})), string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}))
}

func TestIngressTLSCreation(t *testing.T) {
	cert, key := ingressTestKeyPair(t)
	_, otherKey := ingressTestKeyPair(t)
	for _, tc := range []struct {
		ssl       any
		cert, key any
		valid     bool
	}{
		{true, cert, key, true}, {false, nil, nil, true}, {nil, nil, nil, true},
		{true, cert, otherKey, false}, {true, cert, nil, false}, {true, nil, key, false},
		{true, "invalid-cert", key, false}, {true, cert, "invalid-key", false},
		{true, strings.Repeat("x", 65537), key, false}, {true, cert, strings.Repeat("x", 65537), false},
		{false, cert, key, false}, {nil, cert, key, false}, {"true", cert, key, false},
	} {
		p := map[string]any{"service_type": "ingress", "service_id": "rgw.a", "backend_service": "rgw.a", "virtual_ip": "192.0.2.10/24", "frontend_port": 8080, "monitor_port": 9000}
		if tc.ssl != nil {
			p["ssl"] = tc.ssl
		}
		if tc.cert != nil {
			p["ssl_cert"] = tc.cert
		}
		if tc.key != nil {
			p["ssl_key"] = tc.key
		}
		cmd, err := build(Request{Action: "service.create"}, p)
		if (err == nil) != tc.valid {
			t.Fatalf("TLS validation outcome mismatch: %v", err)
		}
		if err != nil {
			if strings.Contains(err.Error(), key) || strings.Contains(err.Error(), "invalid-key") {
				t.Fatal("TLS error leaked input")
			}
			continue
		}
		var spec map[string]any
		json.Unmarshal(cmd.stdin, &spec)
		detail := spec["spec"].(map[string]any)
		if tc.ssl == true && (detail["ssl"] != true || detail["ssl_cert"] != cert || detail["ssl_key"] != key) {
			t.Fatal("native TLS data changed")
		}
		if tc.ssl == false && detail["ssl"] != false {
			t.Fatal("explicit TLS false lost")
		}
		if tc.ssl != true && (detail["ssl_cert"] != nil || detail["ssl_key"] != nil) {
			t.Fatal("disabled TLS retained credentials")
		}
	}
}
