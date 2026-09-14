package mutation

import "testing"

func TestZoneGroupUpdateReadback(t *testing.T) {
	p := map[string]any{"new_name": "east", "zonegroup": "group", "read_only": false, "tier_type": "archive", "sync_from_all": true, "master": true, "endpoints": "https://a,https://b"}
	valid := `{"name":"group","master_zone":"id","zones":[{"id":"id","name":"east","read_only":false,"tier_type":"archive","sync_from_all":true,"endpoints":["https://b","https://a"]}]}`
	if !zoneGroupUpdateMatches(p, []byte(valid)) {
		t.Fatal("valid native state rejected")
	}
	for _, raw := range []string{`null`, `{}`, `{"name":"other","zones":[]}`, `{"name":"group","zones":[{"id":"id","name":"old"}]}`, `{"name":"group","master_zone":"id","zones":[{"id":"id","name":"east","read_only":true}]}`} {
		if zoneGroupUpdateMatches(p, []byte(raw)) {
			t.Fatal("unverified state accepted")
		}
	}
	for _, key := range []string{"read_only", "tier_type", "sync_from_all", "master", "endpoints"} {
		copy := map[string]any{}
		for k, v := range p {
			copy[k] = v
		}
		switch key {
		case "read_only":
			copy[key] = true
		case "tier_type":
			copy[key] = ""
		case "sync_from_all":
			copy[key] = false
		case "master":
			continue
		case "endpoints":
			copy[key] = "https://other"
		}
		if zoneGroupUpdateMatches(copy, []byte(valid)) {
			t.Fatalf("missed %s mismatch", key)
		}
	}
}

func TestZoneSyncSourceReadback(t *testing.T) {
	for _, tc := range []struct {
		all     bool
		sources string
		want    bool
	}{
		{false, `["z1","z2","existing"]`, true},
		{false, `["z1"]`, false},
		{true, `["existing"]`, true},
		{true, `["z2"]`, false},
		{true, `null`, false},
		{false, `[42]`, false},
	} {
		mode := "false"
		if tc.all {
			mode = "true"
		}
		raw := `{"name":"group","zones":[{"name":"east","id":"id","sync_from_all":` + mode + `,"sync_from":` + tc.sources + `}]}`
		p := map[string]any{"new_name": "east", "zonegroup": "group", "sync_from_all": tc.all, "sync_from": "z1,z2"}
		if got := zoneGroupUpdateMatches(p, []byte(raw)); got != tc.want {
			t.Fatalf("all=%v sources=%s got=%v", tc.all, tc.sources, got)
		}
	}
}
