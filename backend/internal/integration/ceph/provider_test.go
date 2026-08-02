package ceph

import "testing"

func TestCephVersionFromVersions(t *testing.T) {
	payload := []byte(`{
		"mon": {"ceph version 20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc) tentacle (stable - RelWithDebInfo)": 3},
		"mgr": {"ceph version 20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc) tentacle (stable - RelWithDebInfo)": 2}
	}`)

	if got, want := cephVersionFromVersions(payload), "20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc)"; got != want {
		t.Fatalf("cephVersionFromVersions() = %q, want %q", got, want)
	}
}

func TestCephVersionFromVersionsRejectsInvalidJSON(t *testing.T) {
	if got := cephVersionFromVersions([]byte(`{"mon":`)); got != "" {
		t.Fatalf("cephVersionFromVersions() = %q, want empty string", got)
	}
}
