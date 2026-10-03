package s3

import (
	"strings"
	"testing"
)

func TestBucketNotifications(t *testing.T) {
	const rule = `<TopicConfiguration><Id>same</Id><Topic>arn:aws:sns:east:tenant:topic</Topic><Event>s3:ObjectCreated:*</Event><Event>future</Event><Filter><S3Key><FilterRule><Name>regex</Name><Value> a&amp;b.* </Value></FilterRule></S3Key><S3Metadata><FilterRule><Name>x-amz-meta-key</Name><Value/></FilterRule></S3Metadata><S3Tags><FilterRule><Name>标签</Name><Value>&lt;script&gt;</Value></FilterRule></S3Tags></Filter></TopicConfiguration>`
	data, err := BucketNotifications([]byte(`<NotificationConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">` + rule + rule + `</NotificationConfiguration>`))
	if err != nil || len(data) != 2 || data[0].ID != data[1].ID || len(data[0].Events) != 2 || data[0].Events[1] != "future" || len(data[0].Filters) != 3 || data[0].Filters[0].Value != " a&b.* " || data[0].Filters[1].Value != "" || data[0].Filters[2].Value != "<script>" {
		t.Fatalf("native values lost: %+v %v", data, err)
	}
	for _, document := range []string{`<NotificationConfiguration/>`, `<NotificationConfiguration><TopicConfiguration><Id/><Topic/></TopicConfiguration></NotificationConfiguration>`} {
		data, err := BucketNotifications([]byte(document))
		if err != nil || data == nil || (len(data) > 0 && (data[0].Events == nil || data[0].Filters == nil)) {
			t.Fatalf("empty data not preserved: %+v %v", data, err)
		}
	}
	for _, bad := range []string{
		"", "broken", `<Other/>`, `<NotificationConfiguration/><NotificationConfiguration/>`, `<!DOCTYPE x><NotificationConfiguration/>`, `<NotificationConfiguration>text</NotificationConfiguration>`, `<NotificationConfiguration><QueueConfiguration/></NotificationConfiguration>`,
		strings.Replace(rule, `<Id>same</Id>`, ``, 1),
		strings.Replace(rule, `<Id>same</Id>`, `<Id>a</Id><Id>b</Id>`, 1),
		strings.Replace(rule, `<Event>future</Event>`, `<Event><Nested/></Event>`, 1),
		strings.Replace(rule, `<Filter>`, `<Filter>text`, 1),
		strings.Replace(rule, `<S3Tags>`, `<S3Metadata/> <S3Tags>`, 1),
		strings.Replace(rule, `<Value/>`, `<Name>duplicate</Name>`, 1),
		strings.Replace(rule, `<Value/>`, `<Value/><Unknown/>`, 1),
		strings.Replace(rule, `<Name>regex</Name>`, `<Name><Nested/></Name>`, 1),
		strings.Replace(rule, `<FilterRule>`, `<FilterRule>text`, 1),
		strings.ReplaceAll(rule, "S3Tags", "UnknownFilter"),
		strings.Repeat("x", (4<<20)+1),
	} {
		if strings.HasPrefix(bad, "<TopicConfiguration>") {
			bad = `<NotificationConfiguration>` + bad + `</NotificationConfiguration>`
		}
		if _, err := BucketNotifications([]byte(bad)); err == nil {
			t.Fatal("accepted malformed or ambiguous notification")
		}
	}
	if ValidateBucketConfiguration("notification", []byte(`<NotificationConfiguration/>`)) == nil || DeletableBucketConfiguration("notification") {
		t.Fatal("read-only notification exposed through generic mutations")
	}
}
