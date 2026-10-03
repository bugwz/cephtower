package s3

import (
	"strings"
	"testing"
)

func TestLifecycleRuleActionConflicts(t *testing.T) {
	transition := func(kind, timing, class string) string {
		return "<" + kind + ">" + timing + "<StorageClass>" + class + "</StorageClass></" + kind + ">"
	}
	days := "<Days>30</Days>"
	date := "<Date>2030-01-01T00:00:00Z</Date>"
	current := transition("Transition", days, "COLD")
	noncurrent := transition("NoncurrentVersionTransition", "<NoncurrentDays>30</NoncurrentDays>", "COLD")
	markerFalse := "<Expiration><ExpiredObjectDeleteMarker>false</ExpiredObjectDeleteMarker></Expiration>"
	for _, tc := range []struct {
		name, fields string
		valid        bool
	}{
		{"duplicate current class", current + current, false},
		{"duplicate noncurrent class", noncurrent + noncurrent, false},
		{"separate class namespaces", current + noncurrent, true},
		{"distinct current classes", current + transition("Transition", days, "ARCHIVE"), true},
		{"mixed transitions", current + transition("Transition", date, "ARCHIVE"), false},
		{"mixed expiration", current + "<Expiration>" + date + "</Expiration>", false},
		{"matching expiration", current + "<Expiration>" + days + "</Expiration>", true},
		{"noncurrent days with current date", noncurrent + "<Expiration>" + date + "</Expiration>", true},
		{"false marker alone", markerFalse, false},
		{"false marker with action", markerFalse + current, true},
		{"true marker alone", "<Expiration><ExpiredObjectDeleteMarker>true</ExpiredObjectDeleteMarker></Expiration>", true},
		{"id boundary", "<ID>" + strings.Repeat("a", 255) + "</ID>" + current, true},
		{"id byte overflow", "<ID>" + strings.Repeat("界", 86) + "</ID>" + current, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte("<LifecycleConfiguration><Rule><Status>Disabled</Status><Filter/>" + tc.fields + "</Rule></LifecycleConfiguration>")
			if err := ValidateBucketConfiguration("lifecycle", body); (err == nil) != tc.valid {
				t.Fatalf("valid=%v err=%v", tc.valid, err)
			}
		})
	}
}

func TestLifecycleDates(t *testing.T) {
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{"1970", true}, {"2030-02", true}, {"2028-02-29", true},
		{"2030-01-01T00Z", true}, {"2030-01-01T00:00Z", true},
		{"2030-01-01T00:00:00Z", true}, {"2030-01-01T00:00:00.000000000Z", true},
		{"2030-01-01T00:00:00.0Z", true},
		{"1969-12-31", false}, {"3000", false}, {"2030-02-29", false},
		{"2030-13", false}, {"2030-00", false}, {"2030-01-00", false},
		{"2030-01-01T24Z", false}, {"2030-01-01T00:01Z", false},
		{"2030-01-01T00:00:01Z", false}, {"2030-01-01T00:00:00.000000001Z", false},
		{"2030-01-01T00:00:00.0000000000Z", false},
		{"2030-01-01T00:00:00+00:00", false}, {"2030-01-01T00:00:00z", false},
		{"2030 garbage", false}, {"2030-01-01T00Z trailing", false},
		{"", false}, {" 2030", false}, {"2030-01-01T00:00:00", false},
	} {
		for _, action := range []string{"Expiration", "Transition"} {
			extra := ""
			if action == "Transition" {
				extra = "<StorageClass>COLD</StorageClass>"
			}
			body := []byte("<LifecycleConfiguration><Rule><Status>Enabled</Status><Filter/><" + action + "><Date>" + tc.value + "</Date>" + extra + "</" + action + "></Rule></LifecycleConfiguration>")
			if err := ValidateBucketConfiguration("lifecycle", body); (err == nil) != tc.valid {
				t.Fatalf("%s date=%q valid=%v err=%v", action, tc.value, tc.valid, err)
			}
		}
	}
}

func TestLifecycleNumericActions(t *testing.T) {
	wrap := func(action string) []byte {
		return []byte("<LifecycleConfiguration><Rule><Status>Enabled</Status><Filter/>" + action + "</Rule></LifecycleConfiguration>")
	}
	for _, tc := range []struct {
		action, field, extra string
		zero                 bool
	}{
		{"Expiration", "Days", "", false},
		{"NoncurrentVersionExpiration", "NoncurrentDays", "", false},
		{"AbortIncompleteMultipartUpload", "DaysAfterInitiation", "", false},
		{"Transition", "Days", "<StorageClass>COLD</StorageClass>", true},
		{"NoncurrentVersionTransition", "NoncurrentDays", "<StorageClass>COLD</StorageClass>", true},
		{"NoncurrentVersionExpiration", "NewerNoncurrentVersions", "<NoncurrentDays>1</NoncurrentDays>", true},
	} {
		for _, value := range []string{"0", "1", "001", "2147483647", "2147483648", "-1", "1x", "1.5", "", " 1"} {
			body := wrap("<" + tc.action + "><" + tc.field + ">" + value + "</" + tc.field + ">" + tc.extra + "</" + tc.action + ">")
			err := ValidateBucketConfiguration("lifecycle", body)
			valid := value == "1" || value == "001" || value == "2147483647" || (value == "0" && tc.zero)
			if (err == nil) != valid {
				t.Fatalf("%s=%q err=%v valid=%v", tc.field, value, err, valid)
			}
		}
	}
	for _, value := range []string{"TRUE", "yes", "", "1"} {
		if err := ValidateBucketConfiguration("lifecycle", wrap("<Expiration><ExpiredObjectDeleteMarker>"+value+"</ExpiredObjectDeleteMarker></Expiration>")); err == nil {
			t.Fatalf("accepted marker %q", value)
		}
	}
}

func TestLifecycleObjectSizeBounds(t *testing.T) {
	wrap := func(bounds string) []byte {
		return []byte("<LifecycleConfiguration><Rule><Status>Enabled</Status><Filter>" + bounds + "</Filter><Expiration><Days>30</Days></Expiration></Rule></LifecycleConfiguration>")
	}
	for _, bounds := range []string{"<ObjectSizeGreaterThan/>", "<ObjectSizeLessThan>18446744073709551615</ObjectSizeLessThan>", "<ObjectSizeGreaterThan>9007199254740992</ObjectSizeGreaterThan><ObjectSizeLessThan>9007199254740993</ObjectSizeLessThan>", "<ObjectSizeGreaterThan>0</ObjectSizeGreaterThan><ObjectSizeLessThan>100</ObjectSizeLessThan>"} {
		if err := ValidateBucketConfiguration("lifecycle", wrap(bounds)); err != nil {
			t.Fatalf("%s: %v", bounds, err)
		}
	}
	for _, value := range []string{"18446744073709551616", "-1", "1x", " 1", "1.2", "1e3", "+1"} {
		if err := ValidateBucketConfiguration("lifecycle", wrap("<ObjectSizeLessThan>"+value+"</ObjectSizeLessThan>")); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
	for _, pair := range [][2]string{{"10", "10"}, {"20", "10"}, {"9", "100"}, {"100", "9"}} {
		if err := ValidateBucketConfiguration("lifecycle", wrap("<ObjectSizeGreaterThan>"+pair[0]+"</ObjectSizeGreaterThan><ObjectSizeLessThan>"+pair[1]+"</ObjectSizeLessThan>")); err == nil {
			t.Fatalf("accepted invalid or native-rejected bounds %v", pair)
		}
	}
}

func TestLifecycleFilterStructure(t *testing.T) {
	wrap := func(filter string) []byte {
		return []byte("<LifecycleConfiguration><Rule><Status>Enabled</Status><Filter>" + filter + "</Filter><Expiration><Days>30</Days></Expiration></Rule></LifecycleConfiguration>")
	}
	for _, filter := range []string{"", "<And/>", "<Prefix>p/</Prefix><Tag><Key>a</Key><Value>b</Value></Tag>", "<And><Prefix/><Tag/><Tag><Key>a</Key></Tag><ArchiveZone/><ObjectSizeGreaterThan>0</ObjectSizeGreaterThan><ObjectSizeLessThan>100</ObjectSizeLessThan></And>"} {
		if err := ValidateBucketConfiguration("lifecycle", wrap(filter)); err != nil {
			t.Fatalf("%s: %v", filter, err)
		}
	}
	for _, filter := range []string{"<And/><Prefix/>", "<And/><And/>", "<And><And/></And>", "<Prefix/><Prefix/>", "<Unknown/>", "<Prefix><Nested/></Prefix>", "<ArchiveZone>false</ArchiveZone>", "<Tag><Key>a</Key><Key>b</Key></Tag>", "<Tag><Value><Nested/></Value></Tag>", "<Tag>text</Tag>", "<Tag><Unknown/></Tag>", "<ObjectSizeGreaterThan>0</ObjectSizeGreaterThan><ObjectSizeGreaterThan>1</ObjectSizeGreaterThan>"} {
		if err := ValidateBucketConfiguration("lifecycle", wrap(filter)); err == nil {
			t.Fatalf("accepted %s", filter)
		}
	}
}

func TestLifecycleActionStructure(t *testing.T) {
	wrap := func(action string) []byte {
		return []byte("<LifecycleConfiguration><Rule><Status>Enabled</Status><Filter/>" + action + "</Rule></LifecycleConfiguration>")
	}
	for _, action := range []string{
		"<Expiration><Date>2030-01-01T00:00:00Z</Date></Expiration>",
		"<Expiration><ExpiredObjectDeleteMarker>true</ExpiredObjectDeleteMarker></Expiration>",
		"<NoncurrentVersionExpiration><NoncurrentDays>2</NoncurrentDays><NewerNoncurrentVersions>3</NewerNoncurrentVersions></NoncurrentVersionExpiration>",
		"<Transition><Date>2030-01-01T00:00:00Z</Date><StorageClass>custom-class</StorageClass></Transition>",
	} {
		if err := ValidateBucketConfiguration("lifecycle", wrap(action)); err != nil {
			t.Fatalf("%s: %v", action, err)
		}
	}
	for _, action := range []string{
		"<Expiration><Days>1</Days><Date>2030-01-01T00:00:00Z</Date></Expiration>",
		"<Expiration><Days>1</Days><Days>2</Days></Expiration>",
		"<Expiration><Days><Nested/></Days></Expiration>",
		"<Expiration><Unknown>1</Unknown></Expiration>",
		"<Transition><Days>1</Days></Transition>",
		"<Transition><StorageClass>COLD</StorageClass></Transition>",
		"<Transition><Days>1</Days><Date>2030-01-01T00:00:00Z</Date><StorageClass>COLD</StorageClass></Transition>",
		"<NoncurrentVersionTransition><StorageClass>COLD</StorageClass></NoncurrentVersionTransition>",
		"<NoncurrentVersionExpiration><NewerNoncurrentVersions>1</NewerNoncurrentVersions></NoncurrentVersionExpiration>",
		"<AbortIncompleteMultipartUpload><Days>1</Days></AbortIncompleteMultipartUpload>",
	} {
		if err := ValidateBucketConfiguration("lifecycle", wrap(action)); err == nil {
			t.Fatalf("accepted %s", action)
		}
	}
}

func TestLifecycleRuleEnvelope(t *testing.T) {
	for _, action := range []string{"<Expiration><Days>30</Days></Expiration>", "<NoncurrentVersionExpiration><NoncurrentDays>30</NoncurrentDays></NoncurrentVersionExpiration>", "<AbortIncompleteMultipartUpload><DaysAfterInitiation>2</DaysAfterInitiation></AbortIncompleteMultipartUpload>", "<Transition><Days>1</Days><StorageClass>COLD</StorageClass></Transition><Transition><Days>2</Days><StorageClass>ARCHIVE</StorageClass></Transition>", "<NoncurrentVersionTransition><NoncurrentDays>2</NoncurrentDays><StorageClass>COLD</StorageClass></NoncurrentVersionTransition>"} {
		for _, filter := range []string{"<Filter/>", "<Prefix/>", "<Filter><And><Prefix>p/</Prefix><ArchiveZone/></And></Filter>"} {
			body := []byte("<LifecycleConfiguration><Rule><Status>Enabled</Status>" + filter + action + "</Rule></LifecycleConfiguration>")
			if err := ValidateBucketConfiguration("lifecycle", body); err != nil {
				t.Fatalf("%s: %v", body, err)
			}
		}
	}
	for _, rule := range []string{"", "<Status>enabled</Status><Filter/><Expiration><Days>1</Days></Expiration>", "<Status>Enabled</Status><Filter/>", "<Status>Enabled</Status><Expiration><Days>1</Days></Expiration>", "<Status>Enabled</Status><Status>Disabled</Status><Filter/><Expiration><Days>1</Days></Expiration>", "<Status>Enabled</Status><Filter/><Prefix/><Expiration><Days>1</Days></Expiration>", "<Status>Enabled</Status><Filter/><Expiration/>", "<Status>Enabled</Status><Filter/><Unknown/>"} {
		if err := ValidateBucketConfiguration("lifecycle", []byte("<LifecycleConfiguration><Rule>"+rule+"</Rule></LifecycleConfiguration>")); err == nil {
			t.Fatalf("accepted %s", rule)
		}
	}
	if err := ValidateBucketConfiguration("lifecycle", []byte("<LifecycleConfiguration/>")); err == nil {
		t.Fatal("accepted empty config")
	}
}
