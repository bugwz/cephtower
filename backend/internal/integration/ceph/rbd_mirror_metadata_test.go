package ceph

import (
	"context"
	"testing"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestMirrorImageMetadataIdentity(t *testing.T) {
	primary := false
	base := cephdomain.RBDImage{Name: "image", MirrorGlobalID: "global", MirrorMode: "snapshot", MirrorState: "enabled", Primary: &primary}
	for _, scenario := range []string{"success", "namespace", "recreated", "missing_id", "duplicate", "missing", "unknown_fields"} {
		t.Run(scenario, func(t *testing.T) {
			image := base
			images := []cephdomain.RBDImage{image}
			switch scenario {
			case "namespace":
				images[0].Namespace = "other"
			case "recreated":
				images[0].MirrorGlobalID = "new-global"
			case "missing_id":
				images[0].MirrorGlobalID = ""
			case "duplicate":
				images = append(images, image)
			case "missing":
				images = nil
			case "unknown_fields":
				images[0].MirrorMode = "future"
				images[0].MirrorState = "future"
				images[0].Primary = nil
			}
			row := map[string]any{"name": "image", "global_id": "global"}
			enrichMirrorImageMetadata([]any{row, nil}, images)
			if scenario == "success" {
				if row["mirror_mode"] != "snapshot" || row["mirror_primary"] != false || row["mirror_image_state"] != "enabled" {
					t.Fatalf("metadata=%#v", row)
				}
			} else if len(row) != 2 {
				t.Fatalf("unverified metadata joined: %#v", row)
			}
		})
	}
	row := map[string]any{"name": "image", "global_id": "global"}
	base.MirrorState = "disabled"
	enrichMirrorImageMetadata([]any{row}, []cephdomain.RBDImage{base})
	if _, ok := row["mirror_image_state"]; ok {
		t.Fatal("default disabled state treated as observed configuration")
	}
	for _, state := range []string{"creating", "disabling"} {
		base.MirrorState = state
		enrichMirrorImageMetadata([]any{row}, []cephdomain.RBDImage{base})
		if row["mirror_image_state"] != state {
			t.Fatalf("missing native transition state: %#v", row)
		}
	}
}

func TestMirrorImageMetadataCollection(t *testing.T) {
	provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rbd_image_detail":     []byte(`[{"image":"image","name":"image"}]`),
		"collect.rbd_namespace":        []byte(`[]`),
		"collect.rbd_image_info":       []byte(`{"features":[],"mirroring":{"mode":"journal","state":"enabled","global_id":"global","primary":true}}`),
		"collect.rbd_mirroring":        []byte(`{"mode":"image","peers":[]}`),
		"collect.rbd_mirroring_status": []byte(`{"images":[{"name":"image","global_id":"global"}]}`),
	}}}
	rows := provider.collectStorageOptional(context.Background(), ClusterAccess{}, []poolWire{{PoolName: "pool-a"}}, fsDumpWire{}, time.Now())
	for _, row := range rows {
		if row.Kind == "rbd_mirroring" {
			image := row.Payload.(map[string]any)["images"].([]any)[0].(map[string]any)
			if image["mirror_mode"] != "journal" || image["mirror_primary"] != true || image["mirror_image_state"] != "enabled" {
				t.Fatalf("missing collected metadata: %#v", image)
			}
			return
		}
	}
	t.Fatal("missing mirroring inventory")
}
