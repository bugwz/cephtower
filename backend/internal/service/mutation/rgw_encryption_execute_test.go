package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type encryptionExecutor struct {
	values     map[string]string
	calls      []executor.CommandSpec
	buffers    [][]byte
	failAt     int
	exitOnly   bool
	drift      string
	omitStored bool
	writes     int
}

func (e *encryptionExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.calls = append(e.calls, spec)
	if len(e.calls) == e.failAt {
		result := executor.CommandResult{Stdout: []byte("private-stdout"), Stderr: []byte("private-stderr"), ExitCode: 5}
		e.buffers = append(e.buffers, result.Stdout, result.Stderr)
		if e.exitOnly {
			return result, nil
		}
		return result, errors.New("private-error")
	}
	args := spec.Args
	if args[0] == "--" {
		args = args[1:]
	}
	body := ""
	switch args[1] {
	case "get":
		if e.drift != "" && strings.Contains(spec.ID, e.drift) {
			e.values["rgw_crypt_vault_namespace"] = "external-change"
		}
		body = e.values[args[3]] + "\n"
	case "set":
		e.writes++
		e.values[args[3]] = strings.TrimPrefix(args[len(args)-1], "--value=")
		body = "private-diagnostic"
	case "dump":
		rows := []map[string]string{}
		if !e.omitStored {
			for option, value := range e.values {
				rows = append(rows, map[string]string{"section": "client.rgw.a", "name": option, "value": value})
			}
		}
		encoded, _ := json.Marshal(rows)
		body = string(encoded)
	default:
		panic("unexpected command")
	}
	result := executor.CommandResult{Stdout: []byte(body), Stderr: []byte("private-diagnostic")}
	e.buffers = append(e.buffers, result.Stdout, result.Stderr)
	return result, nil
}

func newEncryptionExecutor() *encryptionExecutor {
	return &encryptionExecutor{values: map[string]string{"rgw_crypt_s3_kms_backend": "barbican", "rgw_crypt_sse_s3_backend": "vault"}}
}

func TestExecuteRGWEncryptionVerifiesEachWrite(t *testing.T) {
	s, _, id := newCephUserService(t)
	for _, profile := range [][2]string{{"kms", "vault"}, {"kms", "kmip"}, {"s3", "vault"}} {
		values := map[string]any{"namespace": "space", "verify_ssl": false}
		if profile[1] == "kmip" {
			values = map[string]any{"password": "private-password", "username": ""}
		}
		p := encryptionPatch(profile[0], profile[1], values)
		if profile[0] == "s3" {
			p["expected_backend"] = "vault"
		}
		r := Request{ClusterID: id, Action: "rgw_encryption.update", ResourceKey: "rgw/encryption/client.rgw.a", Parameters: p}
		e := newEncryptionExecutor()
		s.executor = e
		result, err := s.Execute(context.Background(), r)
		if err != nil {
			t.Fatal(err)
		}
		if !Supports(r.Action) || e.writes != 2 {
			t.Fatal("action not dispatched")
		}
		encoded, _ := json.Marshal(result)
		if strings.Contains(string(encoded), "private-") || !strings.Contains(string(encoded), `"runtime_verified":false`) {
			t.Fatalf("unsafe result: %s", encoded)
		}
		for _, spec := range e.calls {
			if spec.Binary != executor.BinaryCeph {
				t.Fatal("wrong binary")
			}
			if spec.Mutating {
				if _, ok := spec.SensitiveArgs[len(spec.Args)-1]; !ok {
					t.Fatal("unprotected write argument")
				}
			} else if len(spec.SensitiveArgs) != 0 {
				t.Fatal("write argument indexes propagated to reads")
			}
		}
		for _, buffer := range e.buffers {
			for _, b := range buffer {
				if b != 0 {
					t.Fatal("raw command output retained")
				}
			}
		}
	}
}

func TestExecuteRGWEncryptionStopsOnFailures(t *testing.T) {
	s, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "rgw_encryption.update", ResourceKey: "rgw/encryption/client.rgw.a", Parameters: encryptionPatch("kms", "vault", map[string]any{"namespace": "new"})}
	e := newEncryptionExecutor()
	s.executor = e
	if _, err := s.Execute(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	count := len(e.calls)
	for stage := 1; stage <= count; stage++ {
		for _, exitOnly := range []bool{false, true} {
			e = newEncryptionExecutor()
			e.failAt = stage
			e.exitOnly = exitOnly
			s.executor = e
			_, err := s.Execute(context.Background(), r)
			var failure *cephdomain.ActionError
			if !errors.As(err, &failure) || failure.Retryable || strings.Contains(err.Error(), "private-") || len(e.calls) != stage {
				t.Fatalf("stage=%d err=%v calls=%d", stage, err, len(e.calls))
			}
		}
	}
	for _, stage := range []string{"recheck_namespace", "after_namespace"} {
		e = newEncryptionExecutor()
		e.drift = stage
		s.executor = e
		if _, err := s.Execute(context.Background(), r); err == nil {
			t.Fatal("concurrent configuration change accepted")
		}
		if stage == "recheck_namespace" && e.writes != 0 {
			t.Fatal("wrote after baseline changed")
		}
	}
	e = newEncryptionExecutor()
	e.omitStored = true
	s.executor = e
	if _, err := s.Execute(context.Background(), r); err == nil {
		t.Fatal("effective value without explicit stored entry accepted")
	}
	e = newEncryptionExecutor()
	e.values["rgw_crypt_s3_kms_backend"] = "kmip"
	s.executor = e
	if _, err := s.Execute(context.Background(), r); err == nil || e.writes != 0 {
		t.Fatal("unexpected backend accepted")
	}
	e = newEncryptionExecutor()
	s.executor = e
	r.ResourceKey = "rgw/encryption/client.rgw.other"
	if _, err := s.Execute(context.Background(), r); err == nil || len(e.calls) != 0 {
		t.Fatal("resource identity mismatch executed")
	}
}

func TestRGWEncryptionStoredScope(t *testing.T) {
	for _, body := range []string{`[]`, `null`, `{}`, `[{"section":"client","name":"option","value":"value"}]`, `[{"section":"client.rgw.a","name":"option","value":"other"}]`, `[{"section":"client.rgw.a","name":"option","value":"value","mask":"host:x"}]`, `[{"section":"client.rgw.a","name":"option","value":"value"},{"section":"client.rgw.a","name":"option","value":"value"}]`, `[{"section":"client.rgw.a","name":"option"}]`} {
		if rgwEncryptionStored([]byte(body), "client.rgw.a", "option", "value") {
			t.Fatal("invalid stored configuration accepted")
		}
	}
	if !rgwEncryptionStored([]byte(`[{"section":"client.rgw.a","name":"option","value":"value"}]`), "client.rgw.a", "option", "value") {
		t.Fatal("explicit entry rejected")
	}
}

func TestExecuteRGWEncryptionPartialFailureDoesNotRollback(t *testing.T) {
	s, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "rgw_encryption.update", ResourceKey: "rgw/encryption/client.rgw.a", Parameters: encryptionPatch("kms", "vault", map[string]any{"namespace": "new", "verify_ssl": false})}
	e := newEncryptionExecutor()
	s.executor = e
	if _, err := s.Execute(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	secondSet := 0
	for i, spec := range e.calls {
		if strings.HasSuffix(spec.ID, ".set_verify_ssl") {
			secondSet = i + 1
		}
	}
	if secondSet == 0 {
		t.Fatal("missing second write")
	}
	e = newEncryptionExecutor()
	e.failAt = secondSet
	s.executor = e
	if _, err := s.Execute(context.Background(), r); err == nil || e.writes != 1 || e.values["rgw_crypt_vault_namespace"] != "new" || len(e.calls) != secondSet {
		t.Fatal("partial failure retried, rolled back or lost state")
	}
	e = newEncryptionExecutor()
	s.executor = e
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.Execute(ctx, r); err == nil || len(e.calls) != 0 {
		t.Fatal("cancelled mutation executed")
	}
}
