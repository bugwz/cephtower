package mutation

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

type rbdMirrorBootstrapExecutor struct {
	specs      []executor.CommandSpec
	poolMode   string
	token      string
	failCreate bool
	failImport bool
}

func (e *rbdMirrorBootstrapExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	spec.Stdin = append([]byte(nil), spec.Stdin...)
	e.specs = append(e.specs, spec)
	switch spec.ID {
	case "rbd_mirroring.bootstrap.pool_info", "rbd_mirroring.bootstrap.import.post_check":
		return executor.CommandResult{Stdout: []byte(`{"mode":"` + e.poolMode + `"}`)}, nil
	case "rbd_mirroring.bootstrap.create":
		if e.failCreate {
			return executor.CommandResult{}, fmt.Errorf("secret-create-error")
		}
		return executor.CommandResult{Stdout: []byte(e.token + "\n")}, nil
	case "rbd_mirroring.bootstrap.import":
		if e.failImport {
			return executor.CommandResult{}, fmt.Errorf("secret-import-error")
		}
	}
	return executor.CommandResult{}, nil
}

func newRBDMirrorBootstrapService(t *testing.T) (*Service, *rbdMirrorBootstrapExecutor, uint64) {
	t.Helper()
	service, _, clusterID := newCephUserService(t)
	runner := &rbdMirrorBootstrapExecutor{
		poolMode: "disabled",
		token:    base64.StdEncoding.EncodeToString([]byte(`{"fsid":"fixture","client_id":"rbd-mirror-peer","key":"fixture-key","mon_host":"mon-a:6789"}`)),
	}
	service.executor = runner
	return service, runner, clusterID
}

func TestCreateRBDMirrorBootstrapTokenEnablesPoolAndReturnsToken(t *testing.T) {
	service, runner, clusterID := newRBDMirrorBootstrapService(t)
	token, err := service.CreateRBDMirrorBootstrapToken(context.Background(), clusterID, "images", "site-a")
	if err != nil || token != runner.token {
		t.Fatalf("token=%q err=%v", token, err)
	}
	if len(runner.specs) != 3 {
		t.Fatalf("commands=%d", len(runner.specs))
	}
	want := [][]string{
		{"mirror", "pool", "info", "images", "--format", "json"},
		{"mirror", "pool", "enable", "images", "image"},
		{"mirror", "pool", "peer", "bootstrap", "create", "images", "--site-name=site-a"},
	}
	for index, args := range want {
		if runner.specs[index].Binary != executor.BinaryRBD || !reflect.DeepEqual(runner.specs[index].Args, args) {
			t.Fatalf("command %d=%+v", index, runner.specs[index])
		}
	}
	if !runner.specs[1].Mutating || !runner.specs[2].Mutating {
		t.Fatal("mutating commands were not marked mutating")
	}
}

func TestImportRBDMirrorBootstrapTokenUsesStdinAndPostChecks(t *testing.T) {
	service, runner, clusterID := newRBDMirrorBootstrapService(t)
	runner.poolMode = "image"
	if err := service.ImportRBDMirrorBootstrapToken(context.Background(), clusterID, "images", "site-b", "rx-tx", runner.token); err != nil {
		t.Fatal(err)
	}
	if len(runner.specs) != 3 {
		t.Fatalf("commands=%d", len(runner.specs))
	}
	importSpec := runner.specs[1]
	wantArgs := []string{"mirror", "pool", "peer", "bootstrap", "import", "images", "-", "--site-name=site-b", "--direction=rx-tx"}
	if !reflect.DeepEqual(importSpec.Args, wantArgs) || !importSpec.Mutating {
		t.Fatalf("import=%+v", importSpec)
	}
	if strings.Contains(strings.Join(importSpec.Args, " "), runner.token) {
		t.Fatal("token leaked into process arguments")
	}
	if string(importSpec.Stdin) != runner.token+"\n" {
		t.Fatalf("stdin=%q", importSpec.Stdin)
	}
	if runner.specs[2].ID != "rbd_mirroring.bootstrap.import.post_check" {
		t.Fatalf("post check=%+v", runner.specs[2])
	}
}

func TestRBDMirrorBootstrapRejectsInvalidInputBeforeExecution(t *testing.T) {
	service, runner, clusterID := newRBDMirrorBootstrapService(t)
	for _, test := range []struct {
		name string
		run  func() error
	}{
		{"cluster", func() error {
			_, err := service.CreateRBDMirrorBootstrapToken(context.Background(), 0, "images", "site-a")
			return err
		}},
		{"pool", func() error {
			_, err := service.CreateRBDMirrorBootstrapToken(context.Background(), clusterID, "--help", "site-a")
			return err
		}},
		{"site", func() error {
			_, err := service.CreateRBDMirrorBootstrapToken(context.Background(), clusterID, "images", "")
			return err
		}},
		{"direction", func() error {
			return service.ImportRBDMirrorBootstrapToken(context.Background(), clusterID, "images", "site-a", "tx-only", runner.token)
		}},
		{"token", func() error {
			return service.ImportRBDMirrorBootstrapToken(context.Background(), clusterID, "images", "site-a", "rx-only", "not-a-token")
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); err == nil {
				t.Fatal("invalid input accepted")
			}
		})
	}
	if len(runner.specs) != 0 {
		t.Fatal("invalid input reached executor")
	}
}

func TestRBDMirrorBootstrapHidesExecutorErrorsAndRejectsInvalidOutput(t *testing.T) {
	service, runner, clusterID := newRBDMirrorBootstrapService(t)
	runner.poolMode = "image"
	runner.failCreate = true
	if _, err := service.CreateRBDMirrorBootstrapToken(context.Background(), clusterID, "images", "site-a"); err == nil || strings.Contains(err.Error(), "secret-create-error") {
		t.Fatalf("unsafe create error: %v", err)
	}
	runner.failCreate = false
	runner.token = "invalid-output"
	if _, err := service.CreateRBDMirrorBootstrapToken(context.Background(), clusterID, "images", "site-a"); err == nil {
		t.Fatal("invalid command output accepted")
	}
	runner.token = base64.StdEncoding.EncodeToString([]byte(`{"fsid":"fixture","client_id":"rbd-mirror-peer","key":"fixture-key","mon_host":"mon-a:6789"}`))
	runner.failImport = true
	if err := service.ImportRBDMirrorBootstrapToken(context.Background(), clusterID, "images", "site-a", "rx-only", runner.token); err == nil || strings.Contains(err.Error(), "secret-import-error") {
		t.Fatalf("unsafe import error: %v", err)
	}
}
