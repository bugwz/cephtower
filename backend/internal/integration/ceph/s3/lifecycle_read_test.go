package s3

import "testing"

func TestBucketLifecycleRead(t *testing.T) {
	body := []byte(`<LifecycleConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Rule><ID>a&amp;b</ID><Status>Enabled</Status><Filter><And><Prefix> x/ </Prefix><Tag><Key>a</Key><Value>1</Value></Tag><Tag><Key>a</Key><Value>2</Value></Tag><ObjectSizeGreaterThan>9007199254740993</ObjectSizeGreaterThan><ArchiveZone/></And></Filter><Expiration><Days>90</Days></Expiration><NoncurrentVersionExpiration><NoncurrentDays>30</NoncurrentDays><NewerNoncurrentVersions>0</NewerNoncurrentVersions></NoncurrentVersionExpiration><AbortIncompleteMultipartUpload><DaysAfterInitiation>7</DaysAfterInitiation></AbortIncompleteMultipartUpload><Transition><Days>0</Days><StorageClass>COLD</StorageClass></Transition><Transition><Days>60</Days><StorageClass>ARCHIVE</StorageClass></Transition><NoncurrentVersionTransition><NoncurrentDays>1</NoncurrentDays><StorageClass>COLD</StorageClass></NoncurrentVersionTransition></Rule><Rule><Status>Disabled</Status><Prefix/><Expiration><ExpiredObjectDeleteMarker>true</ExpiredObjectDeleteMarker></Expiration></Rule></LifecycleConfiguration>`)
	rules, err := BucketLifecycle(body)
	if err != nil || len(rules) != 2 {
		t.Fatalf("rules=%v err=%v", rules, err)
	}
	rule := rules[0]
	if rule.ID != "a&b" || rule.Status != "Enabled" || rule.Selector.Kind != "Filter" || !rule.Selector.And || *rule.Selector.Prefix != " x/ " || !rule.Selector.ArchiveZone || *rule.Selector.ObjectSizeGreaterThan != "9007199254740993" || rule.Selector.ObjectSizeLessThan != nil || len(rule.Selector.Tags) != 2 || rule.Selector.Tags[1].Value != "2" {
		t.Fatalf("lost selector fields: %+v", rule)
	}
	if len(rule.Actions) != 6 || rule.Actions[0].Type != "Expiration" || rule.Actions[1].Fields["NewerNoncurrentVersions"] != "0" || rule.Actions[2].Fields["DaysAfterInitiation"] != "7" || rule.Actions[3].Fields["Days"] != "0" || rule.Actions[4].Fields["StorageClass"] != "ARCHIVE" || rule.Actions[5].Type != "NoncurrentVersionTransition" {
		t.Fatalf("lost actions: %+v", rule.Actions)
	}
	if rules[1].Selector.Kind != "Prefix" || rules[1].Selector.Prefix == nil || *rules[1].Selector.Prefix != "" || rules[1].Selector.Tags == nil || rules[1].Actions[0].Fields["ExpiredObjectDeleteMarker"] != "true" {
		t.Fatalf("lost empty prefix or marker: %+v", rules[1])
	}
	for _, invalid := range []string{"broken", "<LifecycleConfiguration/>", "<Wrong/>"} {
		if _, err := BucketLifecycle([]byte(invalid)); err == nil {
			t.Fatalf("accepted %q", invalid)
		}
	}
}
