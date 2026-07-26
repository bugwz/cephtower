package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"cephtower/scripts/aliyun-ceph-lab/internal/state"
)

func TestHelpFlags(t *testing.T) {
	tests := [][]string{
		{"-h"},
		{"--help"},
		{"help"},
		{"help", "create"},
		{"create", "-h"},
		{"create", "--help"},
		{"list", "-h"},
		{"list", "--help"},
		{"delete", "-h"},
		{"delete", "--help"},
	}
	for _, args := range tests {
		if err := run(args); err != nil {
			t.Fatalf("run(%q) returned help error: %v", args, err)
		}
	}
}

func TestRemovedCommandsAreRejected(t *testing.T) {
	for _, command := range []string{"status", "destroy"} {
		if err := run([]string{command}); err == nil {
			t.Fatalf("run(%q) accepted removed command", command)
		}
	}
}

func TestResourceManagementCommands(t *testing.T) {
	for _, command := range []string{"create", "list", "delete"} {
		if !validCommand(command) {
			t.Fatalf("validCommand(%q) = false", command)
		}
	}
}

func TestCreateDoesNotRequireYes(t *testing.T) {
	if usage := commandUsage("create"); strings.Contains(usage, "--yes") {
		t.Fatalf("create usage still requires --yes: %q", usage)
	}
}

func TestDeleteSupportsOptionalYes(t *testing.T) {
	if usage := commandUsage("delete"); !strings.Contains(usage, "--yes") {
		t.Fatalf("delete usage does not document --yes: %q", usage)
	} else if strings.Contains(usage, "delete --yes") {
		t.Fatalf("delete usage still requires --yes: %q", usage)
	}
}

func TestCreateOutputIncludesSSHCredentialsForEveryNode(t *testing.T) {
	current := &state.State{
		Version: state.Version,
		Nodes: []state.Node{
			{Name: "node-1", PublicIP: "203.0.113.10", SSH: state.SSH{
				Host: "203.0.113.10", Port: 22, User: "root", Password: "CephTower#123",
				PasswordGenerated: true, LogPath: "/tmp/log/203.0.113.10.log",
			}},
			{Name: "node-2", PublicIP: "203.0.113.11", SSH: state.SSH{
				Host: "203.0.113.11", Port: 22, User: "root", Password: "CephTower#123",
				PasswordGenerated: true, LogPath: "/tmp/log/203.0.113.11.log",
			}},
		},
	}
	output := newCreateOutput(current)
	if len(output.Nodes) != 2 {
		t.Fatalf("create output has %d nodes, want 2", len(output.Nodes))
	}
	for _, node := range output.Nodes {
		if node.SSHUser != "root" || node.SSHPassword != "CephTower#123" ||
			node.SSHPort != 22 || !node.SSHPasswordGenerated || node.LogPath == "" {
			t.Fatalf("node %s has unexpected SSH credentials: %#v", node.Name, node)
		}
	}
	raw, err := json.Marshal(output)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"ssh_user":"root"`) ||
		!strings.Contains(string(raw), `"ssh_password":"CephTower#123"`) {
		t.Fatalf("create output JSON is missing SSH credentials: %s", raw)
	}
}

func TestDeletePreviewIncludesNodesWithoutSSHCredentials(t *testing.T) {
	current := &state.State{
		ClusterName: "ceph-dev",
		RegionID:    "ap-southeast-1",
		Nodes: []state.Node{{
			Name: "node-1", InstanceID: "i-example", Status: "Running",
			PublicIP: "203.0.113.10", PrivateIP: "172.31.0.10",
			SSH: state.SSH{User: "root", Password: "CephTower#123"},
		}},
	}
	raw, err := json.Marshal(newDeletePreview(current))
	if err != nil {
		t.Fatal(err)
	}
	output := string(raw)
	for _, want := range []string{
		`"cluster_name":"ceph-dev"`, `"instance_id":"i-example"`,
		`"public_ip":"203.0.113.10"`, `"private_ip":"172.31.0.10"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("delete preview is missing %s: %s", want, output)
		}
	}
	if strings.Contains(output, "password") || strings.Contains(output, "CephTower#123") {
		t.Fatalf("delete preview exposed SSH credentials: %s", output)
	}
}

func TestConfirmDeleteAcceptsSupportedAnswers(t *testing.T) {
	for _, test := range []struct {
		input string
		want  bool
	}{
		{input: "y\n", want: true},
		{input: " YES \n", want: true},
		{input: "n\n", want: false},
		{input: "No\n", want: false},
	} {
		var output bytes.Buffer
		got, err := confirmDelete(strings.NewReader(test.input), &output)
		if err != nil {
			t.Fatalf("confirmDelete(%q) returned error: %v", test.input, err)
		}
		if got != test.want {
			t.Fatalf("confirmDelete(%q) = %t, want %t", test.input, got, test.want)
		}
		if !strings.HasSuffix(output.String(), "Enter y/yes or n/no:\n") {
			t.Fatalf("confirmation prompt is not on its own line: %q", output.String())
		}
	}
}

func TestConfirmDeleteRetriesInvalidAnswer(t *testing.T) {
	var output bytes.Buffer
	confirmed, err := confirmDelete(strings.NewReader("maybe\nyes\n"), &output)
	if err != nil {
		t.Fatal(err)
	}
	if !confirmed || !strings.Contains(output.String(), "Invalid response") {
		t.Fatalf("invalid response was not retried: confirmed=%t output=%q", confirmed, output.String())
	}
}
