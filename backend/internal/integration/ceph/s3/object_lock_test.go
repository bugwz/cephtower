package s3

import "testing"

func TestBucketObjectLock(t *testing.T) {
	for _, tc := range []struct {
		rule        string
		mode        string
		days, years *string
	}{
		{}, {`<Rule><DefaultRetention><Mode>GOVERNANCE</Mode><Days>30</Days></DefaultRetention></Rule>`, "GOVERNANCE", stringPointer("30"), nil},
		{`<Rule><DefaultRetention><Mode>COMPLIANCE</Mode><Years>2</Years></DefaultRetention></Rule>`, "COMPLIANCE", nil, stringPointer("2")},
		{`<Rule><DefaultRetention><Mode>FUTURE</Mode><Days>0</Days></DefaultRetention></Rule>`, "FUTURE", stringPointer("0"), nil},
	} {
		body := []byte(`<ObjectLockConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><ObjectLockEnabled>Enabled</ObjectLockEnabled>` + tc.rule + `</ObjectLockConfiguration>`)
		configuration, err := BucketObjectLock(body)
		if err != nil || !configuration.Enabled {
			t.Fatalf("configuration=%+v err=%v", configuration, err)
		}
		if tc.rule == "" {
			if configuration.DefaultRetention != nil {
				t.Fatal("invented retention")
			}
			continue
		}
		retention := configuration.DefaultRetention
		if retention == nil || retention.Mode != tc.mode {
			t.Fatal("lost mode")
		}
		for _, pair := range [][2]*string{{retention.Days, tc.days}, {retention.Years, tc.years}} {
			if (pair[0] == nil) != (pair[1] == nil) || (pair[0] != nil && *pair[0] != *pair[1]) {
				t.Fatal("lost period")
			}
		}
	}
	for _, body := range []string{`broken`, `<ObjectLockConfiguration/>`, `<ObjectLockConfiguration><ObjectLockEnabled>Disabled</ObjectLockEnabled></ObjectLockConfiguration>`, `<ObjectLockConfiguration><ObjectLockEnabled>Enabled</ObjectLockEnabled><Rule/></ObjectLockConfiguration>`, `<ObjectLockConfiguration><ObjectLockEnabled>Enabled</ObjectLockEnabled><Rule><DefaultRetention><Mode>GOVERNANCE</Mode><Days>1</Days><Years>2</Years></DefaultRetention></Rule></ObjectLockConfiguration>`, `<!DOCTYPE x><ObjectLockConfiguration><ObjectLockEnabled>Enabled</ObjectLockEnabled></ObjectLockConfiguration>`} {
		if _, err := BucketObjectLock([]byte(body)); err == nil {
			t.Fatalf("accepted %q", body)
		}
	}
	if ValidateBucketConfiguration("object-lock", []byte(`<ObjectLockConfiguration/>`)) == nil || DeletableBucketConfiguration("object-lock") {
		t.Fatal("invalid configuration or unsupported delete accepted")
	}
}

func TestObjectLockWriteValidation(t *testing.T) {
	for _, period := range []string{"0", "-1", "2147483648", "", "1x", " 1", "1.5"} {
		body := []byte(`<ObjectLockConfiguration><ObjectLockEnabled>Enabled</ObjectLockEnabled><Rule><DefaultRetention><Mode>GOVERNANCE</Mode><Days>` + period + `</Days></DefaultRetention></Rule></ObjectLockConfiguration>`)
		if ValidateBucketConfiguration("object-lock", body) == nil {
			t.Fatalf("accepted %q", period)
		}
	}
	for _, mode := range []string{"GOVERNANCE", "COMPLIANCE", "future"} {
		body := []byte(`<ObjectLockConfiguration><ObjectLockEnabled>Enabled</ObjectLockEnabled><Rule><DefaultRetention><Mode>` + mode + `</Mode><Years>002</Years></DefaultRetention></Rule></ObjectLockConfiguration>`)
		configuration, err := ValidatedBucketObjectLock(body)
		if mode == "future" {
			if err == nil {
				t.Fatal("unknown write mode accepted")
			}
			continue
		}
		if err != nil || *configuration.DefaultRetention.Years != "2" {
			t.Fatalf("native integer normalization failed: %v", err)
		}
	}
	if err := ValidateBucketConfiguration("object-lock", []byte(`<ObjectLockConfiguration><ObjectLockEnabled>Enabled</ObjectLockEnabled></ObjectLockConfiguration>`)); err != nil {
		t.Fatal(err)
	}
}

func stringPointer(value string) *string { return &value }
