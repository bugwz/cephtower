package mutation

import (
	"strings"
	"testing"
)

func TestSubvolumeGroupCreationCaseSensitivity(t *testing.T) {
	for _, value := range []any{nil, true, false, "false"} {
		params := map[string]any{"pool": "data", "normalization": "nfc"}
		if value != nil {
			params["case_sensitive"] = value
		}
		args, err := subvolumeGroupCreateArgs("cephfs", "team", params)
		if value == "false" {
			if err == nil {
				t.Fatal("non-boolean value accepted")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		flag := ""
		for _, arg := range args {
			if strings.HasPrefix(arg, "--casesensitive") {
				flag = arg
			}
		}
		want := ""
		if value == true {
			want = "--casesensitive=true"
		}
		if value == false {
			want = "--casesensitive=false"
		}
		if flag != want {
			t.Fatalf("value %v: got %q want %q", value, flag, want)
		}
	}
}
