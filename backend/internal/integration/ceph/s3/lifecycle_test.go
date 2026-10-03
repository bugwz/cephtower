package s3

import "testing"

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
