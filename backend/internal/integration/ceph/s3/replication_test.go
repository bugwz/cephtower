package s3

import (
	"strings"
	"testing"
)

const replicationExample = `<ReplicationConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Role></Role><Rule><ID>&lt;rule&gt;</ID><Status>Enabled</Status><Priority>-2147483648</Priority><DeleteMarkerReplication><Status>Disabled</Status></DeleteMarkerReplication><Source><Zone>a</Zone><Zone>b</Zone></Source><Destination><Bucket>arn:aws:s3:::target</Bucket><Zone>c</Zone><Zone>d</Zone><StorageClass>COLD</StorageClass><AccessControlTranslation><Owner>Destination</Owner></AccessControlTranslation></Destination><Filter><And><Prefix>prefix/</Prefix><Tag><Key>k</Key><Value>v</Value></Tag></And></Filter></Rule><Rule><ID>other</ID><Status>Disabled</Status><Destination><Bucket>arn:aws:s3:::other</Bucket></Destination></Rule></ReplicationConfiguration>`

func TestBucketReplicationNativeSummary(t *testing.T) {
	data, err := BucketReplication([]byte(replicationExample))
	if err != nil {
		t.Fatal(err)
	}
	if data.Role != "" || len(data.Rules) != 2 || data.Rules[0].ID != "<rule>" || data.Rules[0].Status != "Enabled" || *data.Rules[0].Priority != "-2147483648" || data.Rules[1].Priority != nil || data.Rules[1].DestinationBucket != "arn:aws:s3:::other" {
		t.Fatalf("lost data: %+v", data)
	}
	for _, body := range []string{"<ReplicationConfiguration><Role/></ReplicationConfiguration>", "<ReplicationConfiguration/>"} {
		empty, err := BucketReplication([]byte(body))
		if err != nil || empty.Rules == nil || len(empty.Rules) != 0 {
			t.Fatal("native empty configuration rejected")
		}
	}
	emptyTarget, err := BucketReplication([]byte(strings.Replace(replicationExample, "<Bucket>arn:aws:s3:::target</Bucket>", "<Bucket/>", 1)))
	if err != nil || emptyTarget.Rules[0].DestinationBucket != "" {
		t.Fatal("native empty destination lost")
	}
	unknown, err := BucketReplication([]byte(strings.Replace(replicationExample, ">Enabled<", ">FutureState<", 1)))
	if err != nil || unknown.Rules[0].Status != "FutureState" {
		t.Fatal("unknown status lost")
	}
}

func TestBucketReplicationRejectsAmbiguousSummaries(t *testing.T) {
	for _, body := range []string{
		"", "<Error/>", replicationExample + replicationExample, "<!DOCTYPE x>" + replicationExample,
		strings.Replace(replicationExample, "<Role></Role>", "<Role/><Role/>", 1),
		strings.Replace(replicationExample, "<Role></Role>", "<Role><nested/></Role>", 1),
		strings.Replace(replicationExample, "<Status>Enabled</Status>", "", 1),
		strings.Replace(replicationExample, "<Status>Enabled</Status>", "<Status/>", 1),
		strings.Replace(replicationExample, "<Status>Enabled</Status>", "<Status>Enabled</Status><Status>Disabled</Status>", 1),
		strings.Replace(replicationExample, "<Priority>-2147483648</Priority>", "<Priority><nested/></Priority>", 1),
		strings.Replace(replicationExample, "<Bucket>arn:aws:s3:::target</Bucket>", "<Bucket>a</Bucket><Bucket>b</Bucket>", 1),
		strings.Replace(replicationExample, "<Bucket>arn:aws:s3:::target</Bucket>", "", 1),
		"<ReplicationConfiguration><Rule><Status>Enabled</Status></Rule></ReplicationConfiguration>",
	} {
		if _, err := BucketReplication([]byte(body)); err == nil {
			t.Errorf("ambiguous response accepted: %s", body)
		}
	}
	if ValidateBucketConfiguration("replication", []byte(replicationExample)) == nil || !DeletableBucketConfiguration("replication") {
		t.Fatal("replication must allow deletion but not XML writes")
	}
}
