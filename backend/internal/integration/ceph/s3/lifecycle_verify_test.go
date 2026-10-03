package s3

import (
	"strings"
	"testing"
)

func TestBucketLifecycleMatches(t *testing.T) {
	wrap := func(rules string) []byte {
		return []byte("<LifecycleConfiguration>" + rules + "</LifecycleConfiguration>")
	}
	rule := func(id, filter, actions string) string {
		return "<Rule><ID>" + id + "</ID><Status>Enabled</Status>" + filter + actions + "</Rule>"
	}
	expiration := "<Expiration><Days>30</Days></Expiration>"
	transition := "<Transition><Days>0</Days><StorageClass>COLD</StorageClass></Transition>"
	anonymous := rule("", "<Filter/>", expiration)
	generated := rule("generated", "<Prefix/>", expiration)
	tags := "<Tag><Key>a</Key><Value>1</Value></Tag><Tag><Key>a</Key><Value>2</Value></Tag>"
	reversedTags := "<Tag><Key>a</Key><Value>2</Value></Tag><Tag><Key>a</Key><Value>1</Value></Tag>"
	for _, tc := range []struct {
		name, wanted, actual string
		match                bool
	}{
		{"generated id and empty filter", anonymous, generated, true},
		{"missing generated id", anonymous, anonymous, false},
		{"explicit id changed", rule("explicit", "<Filter/>", expiration), generated, false},
		{"explicit before anonymous matching", anonymous + rule("explicit", "<Prefix/>", expiration), rule("explicit", "<Filter/>", expiration) + generated, true},
		{"duplicate returned ids", anonymous + anonymous, generated + generated, false},
		{"rule missing", anonymous + anonymous, generated, false},
		{"distinct generated ids", anonymous + anonymous, generated + strings.Replace(generated, "generated", "second", 1), true},
		{"changed days", anonymous, strings.Replace(generated, "30", "31", 1), false},
		{"changed status", anonymous, strings.Replace(generated, "Enabled", "Disabled", 1), false},
		{"changed prefix", anonymous, strings.Replace(generated, "<Prefix/>", "<Prefix>x</Prefix>", 1), false},
		{"false marker omission", rule("a", "<Filter/>", "<Expiration><ExpiredObjectDeleteMarker>false</ExpiredObjectDeleteMarker></Expiration>"+transition), rule("a", "<Prefix/>", transition), true},
		{"action ordering", rule("a", "<Filter/>", expiration+transition), rule("a", "<Prefix/>", transition+expiration), true},
		{"changed storage class", rule("a", "<Filter/>", transition), rule("a", "<Filter/>", strings.Replace(transition, "COLD", "ARCHIVE", 1)), false},
		{"and and tag ordering", rule("a", "<Filter>"+tags+"</Filter>", expiration), rule("a", "<Filter><And>"+reversedTags+"</And></Filter>", expiration), true},
		{"tag lost", rule("a", "<Filter>"+tags+"</Filter>", expiration), rule("a", "<Filter><Tag><Key>a</Key><Value>1</Value></Tag></Filter>", expiration), false},
		{"empty scalar omission", rule("a", "<Filter><Prefix/><ObjectSizeLessThan/></Filter>", expiration), rule("a", "<Prefix/>", expiration), true},
		{"precise size changed", rule("a", "<Filter><ObjectSizeGreaterThan>9007199254740993</ObjectSizeGreaterThan></Filter>", expiration), rule("a", "<Filter><ObjectSizeGreaterThan>9007199254740992</ObjectSizeGreaterThan></Filter>", expiration), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			matches, err := BucketLifecycleMatches(wrap(tc.wanted), wrap(tc.actual))
			if err != nil || matches != tc.match {
				t.Fatalf("match=%v expected=%v err=%v", matches, tc.match, err)
			}
		})
	}
	if _, err := BucketLifecycleMatches(wrap(anonymous), []byte("broken")); err == nil {
		t.Fatal("accepted corrupt response")
	}
}
