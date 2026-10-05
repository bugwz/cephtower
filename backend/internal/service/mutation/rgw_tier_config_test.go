package mutation

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTierConfigStringRoundTrip(t *testing.T) {
	for _, value := range []string{"STANDARD", "true", "null", "001", "{}", "{", "}", `"quoted"`, "[]", "中文", "a=b", "comma,name", `back\slash`, ""} {
		encoded := tierConfigString(value)
		if strings.ContainsAny(encoded, ",{}") {
			t.Fatalf("native delimiters leaked: %q", encoded)
		}
		var decoded any
		if json.Unmarshal([]byte(encoded), &decoded) != nil || decoded != value {
			t.Fatalf("literal string lost: %q -> %#v", value, decoded)
		}
	}
}

func TestCloudRestoreLiteralClassAndFollowingGlacierFields(t *testing.T) {
	for _, value := range []string{"STANDARD", "true", "null", "001", "{}", "{", "}", `"quoted"`, "[]", "中文"} {
		p := cloudRestoreParams()
		p["restore_storage_class"] = value
		p["tier_type"] = "cloud-s3-glacier"
		p["glacier_restore_days"] = 3
		p["glacier_restore_tier_type"] = "Expedited"
		spec, err := buildCloudRestore(p)
		if err != nil {
			t.Fatal(value, err)
		}
		// The encoded string contains no raw braces, so native depth stays zero.
		// Each comma must delimit one assignment, including subsequent Glacier keys.
		config := spec.args[len(spec.args)-1]
		parts := strings.Split(config, ",")
		if len(parts) != 6 || strings.ContainsAny(config, "{}") {
			t.Fatal("native token boundaries changed", config)
		}
		fields := map[string]string{}
		for _, part := range parts {
			key, val, ok := strings.Cut(part, "=")
			if !ok {
				t.Fatal(part)
			}
			fields[key] = val
		}
		var decoded any
		if json.Unmarshal([]byte(fields["restore_storage_class"]), &decoded) != nil || decoded != value {
			t.Fatal("class identity changed", value, decoded)
		}
		if fields["glacier_restore_days"] != "3" || fields["glacier_restore_tier_type"] != "Expedited" {
			t.Fatal("following fields swallowed")
		}
	}
}
