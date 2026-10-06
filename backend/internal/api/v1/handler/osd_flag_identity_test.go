package handler

import "testing"

func TestOSDFlagMutationUsesInventoryIdentity(t *testing.T) {
	for _, verb := range []string{"set", "unset"} {
		body := map[string]any{"cluster_id": float64(1), "action": verb, "flag": "noout"}
		if err := ValidateMutationRequest("osd_flag.update", body); err != nil {
			t.Fatal(err)
		}
		key := resourceKey("osd_flag", "osd_flag.update", nil, body)
		if key != "osd-flag" {
			t.Fatalf("unexpected command resource key %q", key)
		}
		lookup := resourceLookupKey("osd_flag", key)
		if lookup != "flags" || lookup != readResourceKey("osd_flag", body) {
			t.Fatalf("mutation version/lock key %q differs from inventory key", lookup)
		}
	}
	// Individual mutations must still lock and validate the selected OSD.
	body := map[string]any{"osd_id": "17"}
	key := resourceKey("osd", "osd.individual_flag", nil, body)
	if lookup := resourceLookupKey("osd", key); lookup != "17" {
		t.Fatalf("individual flag lookup changed: %q", lookup)
	}
}
