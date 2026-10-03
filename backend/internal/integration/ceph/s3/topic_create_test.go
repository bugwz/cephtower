package s3

import "testing"

func TestTopicCreateConfiguration(t *testing.T) {
	valid := TopicCreate{Name: "events", TTL: "None", MaxRetries: "0", RetrySleep: "2147483647", Options: map[string]string{"ca-location": "/etc/ca.pem", "verify-ssl": "true", "kafka-brokers": "a:9092,b:9092"}}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	if valid.EndpointArgs() != "Version=2010-03-31&ca-location=/etc/ca.pem&kafka-brokers=a:9092,b:9092&verify-ssl=true" {
		t.Fatal("wrong native argument ordering")
	}
	for _, change := range []func(*TopicCreate){
		func(p *TopicCreate) { p.Name = "bad:name" }, func(p *TopicCreate) { p.TTL = "2147483648" }, func(p *TopicCreate) { p.Policy = "[]" }, func(p *TopicCreate) { p.Endpoint = "https://user@host/path" }, func(p *TopicCreate) { p.Options = map[string]string{"password": "secret"} }, func(p *TopicCreate) { p.Options = map[string]string{"ca-location": "/a&password=secret"} },
	} {
		p := valid
		change(&p)
		if p.Validate() == nil {
			t.Fatal("invalid configuration accepted")
		}
	}
}
