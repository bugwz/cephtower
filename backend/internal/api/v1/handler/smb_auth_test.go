package handler

import "testing"

func TestSMBAuthDeletionIdentity(t *testing.T) {
	for kind, expected := range map[string]string{"smb_join_auth": "smb/join/auth/target", "smb_usersgroups": "smb/usersgroup/target"} {
		body := map[string]any{"cluster_id": float64(1), "name": "target"}
		if err := ValidateMutationRequest(kind+".delete", body); err != nil {
			t.Fatal(err)
		}
		key := resourceKey(kind, kind+".delete", nil, body)
		if key != expected || resourceLookupKey(kind, key) != "target" {
			t.Fatal(key)
		}
		if err := ValidateMutationRequest(kind+".delete", map[string]any{"cluster_id": float64(1)}); err == nil {
			t.Fatal("missing identity accepted")
		}
	}
}
