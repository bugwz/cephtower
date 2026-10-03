package s3

import "testing"

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
