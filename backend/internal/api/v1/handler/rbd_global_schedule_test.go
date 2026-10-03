package handler

import "testing"

func TestGlobalMirrorScheduleContractAndIdentity(t *testing.T) {
	const action = "rbd_mirroring.global_schedule"
	request := map[string]any{"cluster_id": float64(1), "action": "mirror-schedule-add", "interval": "1h"}
	if err := ValidateMutationRequest(action, request); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"pool", "namespace", "image", "image_spec"} {
		request[key] = "images"
		if err := ValidateMutationRequest(action, request); err == nil {
			t.Fatal("scope field accepted", key)
		}
		delete(request, key)
	}
	key := resourceKey("rbd_mirroring", action, nil, request)
	if key != "global/snapshot-schedule" || resourceLookupKey("rbd_mirroring", key) != key || resourceLookupKey("rbd_mirroring", "snapshot-schedule") == key {
		t.Fatal("global lock identity collides with pool", key)
	}
}
