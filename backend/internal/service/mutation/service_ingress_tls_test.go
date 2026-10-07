package mutation

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"reflect"
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

func TestIngressTLSUpdateChain(t *testing.T) {
	cert, key := ingressTestKeyPair(t)
	current := `[{"service_name":"ingress.rgw.a","service_type":"ingress","service_id":"rgw.a","spec":{"backend_service":"rgw.a","virtual_ip":"192.0.2.10/24","frontend_port":443,"ssl":true,"ssl_cert":"old-cert","ssl_key":"old-key","future":18446744073709551615}}]`
	for _, mode := range []string{"preserve", "replace", "disable"} {
		p := map[string]any{"service_type": "ingress"}
		if mode == "replace" {
			p["ssl"], p["ssl_cert"], p["ssl_key"] = true, cert, key
		}
		if mode == "disable" {
			p["ssl"] = false
		}
		s, _, id := newCephUserService(t)
		e := &directoryRenameExecutor{outputs: map[string]string{"service.update.pre_check": current, "service.update": "Scheduled ingress.rgw.a update..."}}
		s.executor = e
		if _, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "service.update", ResourceKey: "service/ingress.rgw.a", Parameters: p}); err != nil {
			t.Fatal(err)
		}
		if len(e.specs) != 3 || e.specs[0].Mutating || !e.specs[1].Mutating || !reflect.DeepEqual(e.specs[1].Args, []string{"orch", "apply", "-i", "-"}) {
			t.Fatal("wrong TLS update command chain")
		}
		var outer map[string]json.RawMessage
		var spec map[string]json.RawMessage
		json.Unmarshal(e.specs[1].Stdin, &outer)
		json.Unmarshal(outer["spec"], &spec)
		if string(spec["future"]) != "18446744073709551615" || string(spec["backend_service"]) != `"rgw.a"` || string(spec["virtual_ip"]) != `"192.0.2.10/24"` || string(spec["frontend_port"]) != "443" {
			t.Fatal("unrelated ingress settings changed")
		}
		switch mode {
		case "preserve":
			if string(spec["ssl_key"]) != `"old-key"` {
				t.Fatal("default changed credentials")
			}
		case "disable":
			if string(spec["ssl"]) != "false" || spec["ssl_key"] != nil || spec["ssl_cert"] != nil {
				t.Fatal("TLS disable retained credentials")
			}
		case "replace":
			var saved string
			json.Unmarshal(spec["ssl_key"], &saved)
			if saved != key {
				t.Fatal("replacement key lost")
			}
		}
	}
	for _, raw := range []string{`null`, `[]`, `"invalid"`} {
		broken := strings.Replace(current, `{"backend_service":"rgw.a","virtual_ip":"192.0.2.10/24","frontend_port":443,"ssl":true,"ssl_cert":"old-cert","ssl_key":"old-key","future":18446744073709551615}`, raw, 1)
		if _, err := mergeServiceSpec([]byte(broken), []byte(`{"service_type":"ingress","service_id":"rgw.a","spec":{"ssl":false}}`), "ingress.rgw.a"); err == nil {
			t.Fatal("invalid export accepted for TLS edit")
		}
	}
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
