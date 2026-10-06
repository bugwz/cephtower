package ceph

import (
	"encoding/json"
	"testing"
)

func TestOSDTreeUnclassifiedDevice(t *testing.T) {
	for _, tc := range []struct {
		name, node string
		known      bool
		class      string
	}{
		{"omitted", `{"id":0,"name":"osd.0","type":"osd"}`, true, ""},
		{"empty", `{"id":0,"name":"osd.0","type":"osd","device_class":""}`, true, ""},
		{"assigned", `{"id":0,"name":"osd.0","type":"osd","device_class":"ssd"}`, true, "ssd"},
		{"null", `{"id":0,"name":"osd.0","type":"osd","device_class":null}`, false, ""},
		{"bucket", `{"id":-1,"name":"default","type":"root"}`, false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var tree osdTreeWire
			if err := json.Unmarshal([]byte(`{"nodes":[`+tc.node+`]}`), &tree); err != nil {
				t.Fatal(err)
			}
			class := tree.Nodes[0].DeviceClass
			if (class != nil) != tc.known || (class != nil && *class != tc.class) {
				t.Fatalf("unexpected class: %v", class)
			}
		})
	}
}
