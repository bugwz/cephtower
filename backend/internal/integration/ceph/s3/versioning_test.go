package s3

import "testing"

func TestBucketVersioningStatus(t *testing.T) {
	for _, status := range []string{"", "Enabled", "Suspended"} {
		inner := ""
		if status != "" {
			inner = "<Status>" + status + "</Status><MfaDelete>Disabled</MfaDelete>"
		}
		got, err := BucketVersioningStatus([]byte(`<VersioningConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">` + inner + "</VersioningConfiguration>"))
		if err != nil || got != status {
			t.Fatalf("%q: %q %v", status, got, err)
		}
	}
	for _, inner := range []string{"<Status/>", "<Status>enabled</Status>", "<Status>Enabled</Status><Status>Suspended</Status>", "<Status><Enabled/></Status>", "<MfaDelete>Enabled</MfaDelete>", "<Status>Enabled</Status><MfaDelete>unknown</MfaDelete>", "<Unknown/>", "text"} {
		if _, err := BucketVersioningStatus([]byte("<VersioningConfiguration>" + inner + "</VersioningConfiguration>")); err == nil {
			t.Fatalf("accepted %s", inner)
		}
	}
	for _, body := range []string{"broken", "<Wrong/>", "<VersioningConfiguration>", "<!DOCTYPE x><VersioningConfiguration/>"} {
		if _, err := BucketVersioningStatus([]byte(body)); err == nil {
			t.Fatalf("accepted %s", body)
		}
	}
}
