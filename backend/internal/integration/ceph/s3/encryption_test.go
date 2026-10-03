package s3

import "testing"

func TestBucketEncryptionNativeNormalization(t *testing.T) {
	wrap := func(inner string) []byte {
		return []byte(`<ServerSideEncryptionConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">` + inner + `</ServerSideEncryptionConfiguration>`)
	}
	for _, tc := range []struct {
		name, inner string
		want        BucketEncryptionConfiguration
	}{
		{"no rule", "", BucketEncryptionConfiguration{}},
		{"empty rule", "<Rule/>", BucketEncryptionConfiguration{RuleExists: true}},
		{"empty defaults", "<Rule><ApplyServerSideEncryptionByDefault/></Rule>", BucketEncryptionConfiguration{RuleExists: true}},
		{"aes", "<Rule><ApplyServerSideEncryptionByDefault><SSEAlgorithm>AES256</SSEAlgorithm></ApplyServerSideEncryptionByDefault></Rule>", BucketEncryptionConfiguration{RuleExists: true, Algorithm: "AES256"}},
		{"kms", "<Rule><ApplyServerSideEncryptionByDefault><SSEAlgorithm>aws:kms</SSEAlgorithm><KMSMasterKeyID> key&amp;id </KMSMasterKeyID></ApplyServerSideEncryptionByDefault><BucketKeyEnabled>true</BucketKeyEnabled></Rule>", BucketEncryptionConfiguration{RuleExists: true, Algorithm: "aws:kms", KMSMasterKeyID: " key&id ", BucketKeyEnabled: true}},
		{"unknown algorithm preserved", "<Rule><ApplyServerSideEncryptionByDefault><SSEAlgorithm>future:algorithm</SSEAlgorithm></ApplyServerSideEncryptionByDefault></Rule>", BucketEncryptionConfiguration{RuleExists: true, Algorithm: "future:algorithm"}},
		{"explicit defaults", "<Rule><ApplyServerSideEncryptionByDefault><SSEAlgorithm/><KMSMasterKeyID/></ApplyServerSideEncryptionByDefault><BucketKeyEnabled>false</BucketKeyEnabled></Rule>", BucketEncryptionConfiguration{RuleExists: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BucketEncryption(wrap(tc.inner))
			if err != nil || got != tc.want {
				t.Fatalf("got %+v, %v; want %+v", got, err, tc.want)
			}
		})
	}
	for value, want := range map[string]bool{"TRUE": true, "False": false, "0": false, "-0": false, "+000": false, " 2\n": true, "-1": true, "2147483647": true, "-2147483648": true} {
		got, err := BucketEncryption(wrap("<Rule><BucketKeyEnabled>" + value + "</BucketKeyEnabled></Rule>"))
		if err != nil || !got.RuleExists || got.BucketKeyEnabled != want {
			t.Errorf("bool %q: %+v, %v", value, got, err)
		}
	}
	for _, value := range []string{"", " true ", "1.0", "yes", "2147483648", "-2147483649", "1_0", "0x1", "1 0", "\u00a01"} {
		if _, err := BucketEncryption(wrap("<Rule><BucketKeyEnabled>" + value + "</BucketKeyEnabled></Rule>")); err == nil {
			t.Errorf("accepted invalid native bool %q", value)
		}
	}
}

func TestBucketEncryptionRejectsLossyDocuments(t *testing.T) {
	for _, inner := range []string{
		"<Rule/><Rule/>", "<Unknown/>", "text", "<Rule>text</Rule>",
		"<Rule><Unknown/></Rule>", "<Rule><BucketKeyEnabled>true</BucketKeyEnabled><BucketKeyEnabled>false</BucketKeyEnabled></Rule>",
		"<Rule><BucketKeyEnabled><Nested/></BucketKeyEnabled></Rule>",
		"<Rule><ApplyServerSideEncryptionByDefault/><ApplyServerSideEncryptionByDefault/></Rule>",
		"<Rule><ApplyServerSideEncryptionByDefault>text</ApplyServerSideEncryptionByDefault></Rule>",
		"<Rule><ApplyServerSideEncryptionByDefault><Unknown/></ApplyServerSideEncryptionByDefault></Rule>",
		"<Rule><ApplyServerSideEncryptionByDefault><SSEAlgorithm/><SSEAlgorithm/></ApplyServerSideEncryptionByDefault></Rule>",
		"<Rule><ApplyServerSideEncryptionByDefault><KMSMasterKeyID/><KMSMasterKeyID/></ApplyServerSideEncryptionByDefault></Rule>",
		"<Rule><ApplyServerSideEncryptionByDefault><SSEAlgorithm><Nested/></SSEAlgorithm></ApplyServerSideEncryptionByDefault></Rule>",
		"<Rule><ApplyServerSideEncryptionByDefault><KMSMasterKeyID><Nested/></KMSMasterKeyID></ApplyServerSideEncryptionByDefault></Rule>",
	} {
		body := []byte("<ServerSideEncryptionConfiguration>" + inner + "</ServerSideEncryptionConfiguration>")
		if err := ValidateBucketConfiguration("encryption", body); err == nil {
			t.Errorf("accepted lossy document %s", inner)
		}
	}
	for _, body := range []string{"<Wrong/>", "<ServerSideEncryptionConfiguration>", "<!DOCTYPE x><ServerSideEncryptionConfiguration/>", "<ServerSideEncryptionConfiguration/><ServerSideEncryptionConfiguration/>"} {
		if _, err := BucketEncryption([]byte(body)); err == nil {
			t.Errorf("accepted malformed document %s", body)
		}
	}
}
