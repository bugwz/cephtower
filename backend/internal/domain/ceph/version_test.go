package ceph

import "testing"

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "full ceph version",
			value: "ceph version 20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc) tentacle (stable - RelWithDebInfo)",
			want:  "20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc)",
		},
		{name: "version without hash", value: "ceph version 20.2.2", want: "20.2.2"},
		{name: "already normalized", value: "20.2.2", want: "20.2.2"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NormalizeVersion(test.value); got != test.want {
				t.Fatalf("NormalizeVersion() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestIsVersion(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "20.2.2", want: true},
		{value: "ceph version 20.2.2", want: true},
		{value: "grafana 12", want: false},
		{value: "", want: false},
	}
	for _, test := range tests {
		if got := IsVersion(test.value); got != test.want {
			t.Fatalf("IsVersion(%q) = %v, want %v", test.value, got, test.want)
		}
	}
}

func TestVersionHasCommit(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc)", want: true},
		{value: "ceph version 20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc) tentacle", want: true},
		{value: "20.2.2", want: false},
		{value: "", want: false},
	}
	for _, test := range tests {
		if got := VersionHasCommit(test.value); got != test.want {
			t.Fatalf("VersionHasCommit(%q) = %v, want %v", test.value, got, test.want)
		}
	}
}
