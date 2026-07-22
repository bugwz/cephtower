package command

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCommandRunBuildsCephArgs(t *testing.T) {
	var gotBin string
	var gotArgs []string
	client := NewCommandClientWithRunner(Config{
		Bin:     "/usr/bin/ceph",
		Cluster: "prod",
		Conf:    "/etc/ceph/prod.conf",
		MonHost: "10.0.0.11:6789,10.0.0.12:6789",
		Name:    "client.cephtower",
		Keyring: "/etc/ceph/ceph.client.cephtower.keyring",
		Timeout: time.Second,
	}, func(_ context.Context, bin string, args []string, _ []byte) (CommandResult, error) {
		gotBin = bin
		gotArgs = append([]string(nil), args...)
		return CommandResult{Bin: bin, Args: args}, nil
	})

	_, err := client.Run(context.Background(), CommandRequest{
		Args:   []string{"status"},
		Format: "json",
	})
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if gotBin != "/usr/bin/ceph" {
		t.Fatalf("bin = %q, want configured ceph binary", gotBin)
	}
	wantArgs := []string{
		"--cluster", "prod",
		"--conf", "/etc/ceph/prod.conf",
		"--mon-host", "10.0.0.11:6789,10.0.0.12:6789",
		"--name", "client.cephtower",
		"--keyring", "/etc/ceph/ceph.client.cephtower.keyring",
		"status",
		"--format", "json",
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
}

func TestCommandJSONDecodesStdout(t *testing.T) {
	client := NewCommandClientWithRunner(Config{Bin: "ceph"}, func(_ context.Context, bin string, args []string, _ []byte) (CommandResult, error) {
		return CommandResult{
			Bin:    bin,
			Args:   args,
			Stdout: []byte(`{"health":{"status":"HEALTH_OK"}}`),
		}, nil
	})

	status, err := client.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() returned error: %v", err)
	}

	health, ok := status["health"].(map[string]any)
	if !ok || health["status"] != "HEALTH_OK" {
		t.Fatalf("status = %#v, want decoded health status", status)
	}
}

func TestCommandRunDoesNotDuplicateFormatArg(t *testing.T) {
	var gotArgs []string
	client := NewCommandClientWithRunner(Config{Bin: "ceph"}, func(_ context.Context, bin string, args []string, _ []byte) (CommandResult, error) {
		gotArgs = append([]string(nil), args...)
		return CommandResult{Bin: bin, Args: args}, nil
	})

	_, err := client.Run(context.Background(), CommandRequest{
		Args:   []string{"status", "--format", "json-pretty"},
		Format: "json",
	})
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	wantArgs := []string{"status", "--format", "json-pretty"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
}

func TestCommandRunReturnsCommandErrorForNonZeroExit(t *testing.T) {
	client := NewCommandClientWithRunner(Config{Bin: "ceph"}, func(_ context.Context, bin string, args []string, _ []byte) (CommandResult, error) {
		return CommandResult{
			Bin:      bin,
			Args:     args,
			Stderr:   []byte("permission denied"),
			ExitCode: 13,
		}, nil
	})

	_, err := client.Run(context.Background(), CommandRequest{Args: []string{"status"}})
	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("Run() error = %T, want *CommandError", err)
	}
	if commandErr.Result.ExitCode != 13 || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("CommandError = %#v, message %q", commandErr, err.Error())
	}
}

func TestCommandConvenienceMethodsBuildDocumentedCommands(t *testing.T) {
	var calls [][]string
	client := NewCommandClientWithRunner(Config{Bin: "ceph"}, func(_ context.Context, bin string, args []string, _ []byte) (CommandResult, error) {
		calls = append(calls, append([]string(nil), args...))
		return CommandResult{Bin: bin, Args: args, Stdout: []byte(`[]`)}, nil
	})

	if _, err := client.OrchHosts(context.Background(), OrchHostListOptions{
		HostPattern: "node-*",
		Label:       "mon",
		HostStatus:  "online",
		Detail:      true,
	}); err != nil {
		t.Fatalf("OrchHosts() returned error: %v", err)
	}
	if _, err := client.OrchPS(context.Background(), OrchPSOptions{
		Hostname:    "node-1",
		DaemonType:  "osd",
		DaemonID:    "1",
		ServiceName: "osd",
		SortBy:      "daemon_name",
		Refresh:     true,
	}); err != nil {
		t.Fatalf("OrchPS() returned error: %v", err)
	}

	want := [][]string{
		{"orch", "host", "ls", "--host-pattern", "node-*", "--label", "mon", "--host-status", "online", "--detail", "--format", "json"},
		{"orch", "ps", "node-1", "--daemon-type", "osd", "--daemon-id", "1", "--service-name", "osd", "--sort-by", "daemon_name", "--refresh", "--format", "json"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

func TestGenericCommandHelpersCoverUnwrappedCommands(t *testing.T) {
	var calls [][]string
	client := NewCommandClientWithRunner(Config{Bin: "ceph"}, func(_ context.Context, bin string, args []string, _ []byte) (CommandResult, error) {
		calls = append(calls, append([]string(nil), args...))
		return CommandResult{Bin: bin, Args: args, Stdout: []byte(`{"ok":true}`)}, nil
	})

	if _, err := client.JSONCommand(context.Background(), "nvmeof", "subsystem", "list"); err != nil {
		t.Fatalf("JSONCommand() returned error: %v", err)
	}
	if _, err := client.Text(context.Background(), "auth", "get-key", "client.admin"); err != nil {
		t.Fatalf("Text() returned error: %v", err)
	}

	want := [][]string{
		{"nvmeof", "subsystem", "list", "--format", "json"},
		{"auth", "get-key", "client.admin"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

func TestCommandWriteOperationsBuildArgsAndInput(t *testing.T) {
	var gotArgs []string
	var gotInput []byte
	client := NewCommandClientWithRunner(Config{Bin: "ceph"}, func(_ context.Context, bin string, args []string, input []byte) (CommandResult, error) {
		gotArgs = append([]string(nil), args...)
		gotInput = append([]byte(nil), input...)
		return CommandResult{Bin: bin, Args: args}, nil
	})

	_, err := client.ConfigKeySet(context.Background(), "mgr/dashboard/cert", []byte("secret"))
	if err != nil {
		t.Fatalf("ConfigKeySet() returned error: %v", err)
	}

	wantArgs := []string{"config-key", "set", "mgr/dashboard/cert", "-i", "-"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
	if string(gotInput) != "secret" {
		t.Fatalf("input = %q, want secret", string(gotInput))
	}
}

func TestExpandedCommandMethodsBuildExpectedArgs(t *testing.T) {
	var calls [][]string
	client := NewCommandClientWithRunner(Config{Bin: "ceph"}, func(_ context.Context, bin string, args []string, _ []byte) (CommandResult, error) {
		calls = append(calls, append([]string(nil), args...))
		return CommandResult{Bin: bin, Args: args, Stdout: []byte(`{}`)}, nil
	})

	_, _ = client.OSDReweightByUtilization(context.Background(), intPtr(120), floatPtr(0.05), intPtr(4), true)
	_, _ = client.OrchUpgradeStart(context.Background(), "quay.io/ceph/ceph:v20", "", []string{"mon", "mgr"}, []string{"node-1"}, nil, "2")
	_, _ = client.FSSubvolumeCreate(context.Background(), "cephfs", "vol-a", FSSubvolumeOptions{
		Size:      "1G",
		GroupName: "group-a",
		Pool:      "data",
		UID:       "1000",
		GID:       "1000",
		Mode:      "0755",
	})

	want := [][]string{
		{"osd", "reweight-by-utilization", "120", "0.05", "4", "--no-increasing"},
		{"orch", "upgrade", "start", "quay.io/ceph/ceph:v20", "--daemon-types", "mon", "--daemon-types", "mgr", "--hosts", "node-1", "--limit", "2"},
		{"fs", "subvolume", "create", "cephfs", "vol-a", "1G", "group-a", "data", "1000", "1000", "0755"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

func intPtr(value int) *int {
	return &value
}

func floatPtr(value float64) *float64 {
	return &value
}
