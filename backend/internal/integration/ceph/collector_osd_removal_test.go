package ceph

import (
	"encoding/json"
	"testing"
	"time"
)

func TestOSDRemovalRows(t *testing.T) {
	for _, raw := range []string{"No OSD remove/replace operations reported\n", "[]"} {
		rows, valid := osdRemovalRows([]byte(raw), time.Time{})
		if !valid || len(rows) != 0 {
			t.Fatalf("empty queue %q: %v %v", raw, rows, valid)
		}
	}
	for _, raw := range []string{`null`, `{}`, `[null]`, `[{}]`, `[{"osd_id":"0"}]`, `[{"osd_id":-1}]`, `[{"osd_id":2147483648}]`, `[{"osd_id":0},{"osd_id":0}]`, `[] {}`, `No OSD remove/replace operations reported: error`} {
		if rows, valid := osdRemovalRows([]byte(raw), time.Time{}); valid || len(rows) != 0 {
			t.Fatalf("invalid queue accepted: %s", raw)
		}
	}
	rows, valid := osdRemovalRows([]byte(`[{"osd_id":0,"pg_count":18446744073709551615,"drain_status":"draining"}]`), time.Time{})
	if !valid || len(rows) != 1 || rows[0].NaturalKey != "0" {
		t.Fatalf("rows=%v valid=%v", rows, valid)
	}
	if rows[0].Payload.(map[string]any)["pg_count"].(json.Number).String() != "18446744073709551615" {
		t.Fatal("counter precision lost")
	}
}
