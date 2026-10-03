package s3

import (
	"reflect"
	"strings"
	"testing"
)

func TestBucketTagsPreserveNativeEntries(t *testing.T) {
	body := []byte(`<Tagging xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><TagSet><Tag><Key>same</Key><Value>z</Value></Tag><Tag><Key>same</Key><Value></Value></Tag><Tag><Key>same</Key><Value>z</Value></Tag><Tag><Key> 中文 </Key><Value>&lt;&amp;&gt;</Value></Tag></TagSet></Tagging>`)
	tags, err := BucketTags(body)
	want := []BucketTag{{" 中文 ", "<&>"}, {"same", ""}, {"same", "z"}, {"same", "z"}}
	if err != nil || !reflect.DeepEqual(tags, want) {
		t.Fatalf("lost tag entries: %+v %v", tags, err)
	}
	tags, err = BucketTags([]byte(`<Tagging><TagSet/></Tagging>`))
	if err != nil || tags == nil || len(tags) != 0 {
		t.Fatalf("empty set lost: %+v %v", tags, err)
	}
	entry := `<Tag><Key>` + strings.Repeat("中", 42) + `aa</Key><Value>` + strings.Repeat("x", 256) + `</Value></Tag>`
	if _, err := BucketTags([]byte(`<Tagging><TagSet>` + strings.Repeat(entry, 50) + `</TagSet></Tagging>`)); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []string{
		`<Tagging/>`, `<Tagging><TagSet/><TagSet/></Tagging>`, `<Tagging><Other/></Tagging>`,
		`<Tagging><TagSet><Tag><Key>a</Key></Tag></TagSet></Tagging>`,
		`<Tagging><TagSet><Tag><Key/><Value/></Tag></TagSet></Tagging>`,
		`<Tagging><TagSet><Tag><Key>a</Key><Key>b</Key><Value/></Tag></TagSet></Tagging>`,
		`<Tagging><TagSet><Tag><Key><Nested/>a</Key><Value/></Tag></TagSet></Tagging>`,
		`<Tagging><TagSet><Other/></TagSet></Tagging>`,
		`<Tagging><TagSet>` + strings.Repeat(entry, 51) + `</TagSet></Tagging>`,
		`<Tagging><TagSet><Tag><Key>` + strings.Repeat("中", 43) + `</Key><Value/></Tag></TagSet></Tagging>`,
		`<Tagging><TagSet><Tag><Key>a</Key><Value>` + strings.Repeat("x", 257) + `</Value></Tag></TagSet></Tagging>`,
		`<Tagging><TagSet>garbage</TagSet></Tagging>`, `<Tagging><TagSet/></Tagging><Tagging/>`,
	} {
		if _, err := BucketTags([]byte(invalid)); err == nil {
			t.Fatalf("accepted %.120s", invalid)
		}
	}
	if !IsConfigurationMissing("tagging", responseError(404, strings.NewReader(`<Error><Code>NoSuchTagSet</Code></Error>`))) {
		t.Fatal("missing tags not recognized")
	}
	if IsConfigurationMissing("tagging", responseError(404, strings.NewReader(`<Error><Code>NoSuchBucket</Code></Error>`))) {
		t.Fatal("missing bucket misclassified")
	}
}
