package handler

import "testing"

func TestUpgradeTargetRequestContracts(t *testing.T) {
	for _, action := range []string{"upgrade.check", "upgrade.action"} {
		for _, field := range []string{"version", "image"} {
			body := map[string]any{"cluster_id": float64(1), field: "target"}
			if action == "upgrade.action" {
				body["action"] = "start"
			}
			if err := ValidateMutationRequest(action, body); err != nil {
				t.Fatal(action, field, err)
			}
			body[field] = true
			if err := ValidateMutationRequest(action, body); err == nil {
				t.Fatal("accepted non-string target")
			}
		}
	}
}
