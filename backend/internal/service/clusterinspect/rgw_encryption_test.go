package clusterinspect

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWEncryptionReadClearsCommandBuffers(t *testing.T) {
	for _, mode := range []string{"success", "error", "exit", "invalid", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			s, runner, id := testInspection(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var buffers [][]byte
			runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
				// Each previous command must be cleared before another command starts.
				for _, buffer := range buffers {
					if !bytes.Equal(buffer, make([]byte, len(buffer))) {
						t.Fatal("previous output retained")
					}
				}
				value := "\n"
				password := spec.Args[3] == "rgw_crypt_kmip_password"
				if password {
					value = "private-password\n"
				}
				if password && mode == "invalid" {
					value = "private-password\x00\n"
				}
				result := executor.CommandResult{Stdout: []byte(value), Stderr: []byte("private-diagnostic")}
				buffers = append(buffers, result.Stdout, result.Stderr)
				if password {
					switch mode {
					case "error":
						return result, errors.New("private-error")
					case "exit":
						result.ExitCode = 2
					case "cancel":
						cancel()
					}
				}
				return result, nil
			}
			config, err := s.RGWEncryptionConfiguration(ctx, id, "client.rgw.a", "kms", "kmip")
			if (err == nil) != (mode == "success") {
				t.Fatal("unexpected outcome")
			}
			encoded, _ := json.Marshal(config)
			if strings.Contains(string(encoded), "private-") || err != nil && strings.Contains(err.Error(), "private-") {
				t.Fatal("private output leaked")
			}
			for _, buffer := range buffers {
				if !bytes.Equal(buffer, make([]byte, len(buffer))) {
					t.Fatal("command output retained")
				}
			}
		})
	}
}

func TestRGWEncryptionConfigurationProfiles(t *testing.T) {
	for _, profile := range [][2]string{{"kms", "vault"}, {"kms", "kmip"}, {"s3", "vault"}} {
		t.Run(strings.Join(profile[:], "/"), func(t *testing.T) {
			s, runner, id := testInspection(t)
			runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
				key := spec.Args[3]
				value := ""
				switch {
				case strings.HasSuffix(key, "_backend"):
					value = "barbican"
				case strings.HasSuffix(key, "_password"):
					value = "private-password"
				case strings.HasSuffix(key, "_addr"):
					if profile[1] == "vault" {
						value = "https://user:private-password@vault.example:8200/path?secret=private-query#private-fragment"
					} else {
						value = "kmip.example:5696"
					}
				case strings.HasSuffix(key, "_prefix"):
					value = " /preserve spaces/ "
				case strings.HasSuffix(key, "_verify_ssl"):
					value = "false"
				case strings.HasSuffix(key, "_key_template"):
					value = "%bucket_id"
				}
				return executor.CommandResult{Stdout: []byte(value + "\n")}, nil
			}
			config, err := s.RGWEncryptionConfiguration(context.Background(), id, "client.rgw.a", profile[0], profile[1])
			if err != nil {
				t.Fatal(err)
			}
			if config.Backend != "barbican" || config.Provider != profile[1] || config.Entity != "client.rgw.a" || config.ObservedAt.IsZero() {
				t.Fatalf("config=%+v", config)
			}
			count := 10
			if profile[1] == "kmip" {
				count = 8
			}
			if profile[0] == "s3" {
				count++
			}
			if len(config.Fields) != count || len(runner.specs) != count+2 {
				t.Fatalf("fields=%d calls=%d", len(config.Fields), len(runner.specs))
			}
			for _, spec := range runner.specs {
				if spec.Binary != executor.BinaryCeph || spec.Mutating || len(spec.Args) != 4 || !reflect.DeepEqual(spec.Args[:3], []string{"config", "get", "client.rgw.a"}) {
					t.Fatalf("unexpected command=%+v", spec)
				}
				if strings.Contains(spec.Args[3], "unique_id") || strings.Contains(spec.Args[3], "encryption_type") || strings.HasSuffix(spec.Args[3], "vault_backend") {
					t.Fatal("synthetic reference field became a native option")
				}
			}
			if !reflect.DeepEqual(runner.specs[0].Args, runner.specs[len(runner.specs)-1].Args) {
				t.Fatal("backend not rechecked")
			}
			for _, field := range config.Fields {
				if field.Name == "prefix" && field.Value != " /preserve spaces/ " {
					t.Fatal("trimmed native value")
				}
				if field.Name == "verify_ssl" && field.Value != "false" {
					t.Fatal("lost false")
				}
				if field.Name == "addr" && profile[1] == "vault" && (field.Value != "https://vault.example:8200/path" || !field.Redacted) {
					t.Fatal("unsafe address")
				}
				if field.Name == "password" && (field.Value != "[REDACTED]" || !field.Redacted) {
					t.Fatal("unsafe password")
				}
				if field.Name == "key_template" && field.Option != "rgw_crypt_sse_s3_key_template" {
					t.Fatal("incorrect S3 template key")
				}
			}
			encoded, _ := json.Marshal(config)
			if strings.Contains(string(encoded), "private-") {
				t.Fatal("secret leaked")
			}
		})
	}
}

func TestRGWEncryptionConfigurationRejectsIncompleteReads(t *testing.T) {
	s, runner, id := testInspection(t)
	for _, scope := range [][3]string{{"client.admin", "kms", "vault"}, {"client.rgw.a\n", "kms", "vault"}, {"client.rgw.a", "s3", "kmip"}, {"client.rgw.a", "kms", "--help"}} {
		if _, err := s.RGWEncryptionConfiguration(context.Background(), id, scope[0], scope[1], scope[2]); err == nil {
			t.Fatal("invalid scope accepted")
		}
	}
	if len(runner.specs) != 0 {
		t.Fatal("invalid scope issued command")
	}
	for stage := 0; stage < 12; stage++ {
		runner.specs = nil
		runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
			if len(runner.specs) == stage+1 {
				return executor.CommandResult{Stdout: []byte("private-output"), ExitCode: 5}, errors.New("private-error")
			}
			return executor.CommandResult{Stdout: []byte("\n")}, nil
		}
		config, err := s.RGWEncryptionConfiguration(context.Background(), id, "client.rgw.a", "kms", "vault")
		if err == nil || strings.Contains(err.Error(), "private-") || len(config.Fields) != 0 || len(runner.specs) != stage+1 {
			t.Fatalf("stage=%d config=%+v err=%v", stage, config, err)
		}
	}
	for _, output := range []string{"missing newline", "a\n\n", "a\x00\n", string([]byte{0xff, '\n'}), strings.Repeat("a", 16384) + "\n"} {
		runner.run = nil
		runner.output = output
		if _, err := s.RGWEncryptionConfiguration(context.Background(), id, "client.rgw.a", "kms", "vault"); err == nil {
			t.Fatal("malformed scalar accepted")
		}
	}
	runner.specs = nil
	runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
		value := "\n"
		if len(runner.specs) == 12 {
			value = "changed\n"
		}
		return executor.CommandResult{Stdout: []byte(value)}, nil
	}
	if _, err := s.RGWEncryptionConfiguration(context.Background(), id, "client.rgw.a", "kms", "vault"); err == nil {
		t.Fatal("backend drift accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	runner.specs = nil
	if _, err := s.RGWEncryptionConfiguration(ctx, id, "client.rgw.a", "kms", "vault"); err == nil || len(runner.specs) != 0 {
		t.Fatal("cancelled read executed")
	}
}

func TestRGWEncryptionConfigurationCancellationAndUnsafeAddress(t *testing.T) {
	s, runner, id := testInspection(t)
	for _, address := range []string{"invalid", "file:///private-secret", "https://vault.example\\private-secret", "https://vault.example?"} {
		runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
			value := ""
			if strings.HasSuffix(spec.Args[3], "_addr") {
				value = address
			}
			return executor.CommandResult{Stdout: []byte(value + "\n")}, nil
		}
		config, err := s.RGWEncryptionConfiguration(context.Background(), id, "client.rgw.a", "kms", "vault")
		if err != nil || !config.Fields[0].Redacted || strings.Contains(config.Fields[0].Value, "private-secret") {
			t.Fatalf("unsafe address accepted: %+v %v", config, err)
		}
	}
	runner.run = func(executor.CommandSpec) (executor.CommandResult, error) {
		return executor.CommandResult{ExitCode: 5, Stdout: []byte("private-secret\n")}, nil
	}
	if _, err := s.RGWEncryptionConfiguration(context.Background(), id, "client.rgw.a", "kms", "vault"); err == nil || strings.Contains(err.Error(), "private-secret") {
		t.Fatal("nonzero exit was accepted or leaked")
	}
	ctx, cancel := context.WithCancel(context.Background())
	runner.specs = nil
	runner.run = func(executor.CommandSpec) (executor.CommandResult, error) {
		cancel()
		return executor.CommandResult{Stdout: []byte("\n")}, nil
	}
	if config, err := s.RGWEncryptionConfiguration(ctx, id, "client.rgw.a", "kms", "vault"); err == nil || len(config.Fields) != 0 || len(runner.specs) != 1 {
		t.Fatal("read continued after cancellation")
	}
}
