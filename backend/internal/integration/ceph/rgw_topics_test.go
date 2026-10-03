package ceph

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

type topicExecutor struct {
	t            *testing.T
	list, detail string
	calls        *int
}

func (e topicExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	*e.calls++
	if spec.Binary != executor.BinaryRGWAdmin {
		e.t.Fatal("wrong binary")
	}
	if spec.ID == "collect.rgw_topic" {
		if !reflect.DeepEqual(spec.Args, []string{"metadata", "list", "topic", "--format", "json"}) {
			e.t.Fatal(spec.Args)
		}
		return executor.CommandResult{Stdout: []byte(e.list)}, nil
	}
	if spec.ID != "collect.rgw_topic_detail" || len(spec.Args) != 5 || spec.Args[0] != "metadata" || spec.Args[1] != "get" || spec.Args[3] != "--format" || spec.Args[4] != "json" {
		e.t.Fatal(spec.Args)
	}
	if e.detail == "failure" {
		return executor.CommandResult{}, fmt.Errorf("unavailable")
	}
	body := strings.ReplaceAll(e.detail, "REQUEST_KEY", spec.Args[2])
	return executor.CommandResult{Stdout: []byte(body)}, nil
}

const topicDocument = `{"key":"REQUEST_KEY","data":{"name":"events","owner":"tenant$user","arn":"arn:aws:sns:zone:tenant:events","opaqueData":"note","policy":"{}","unknown_secret":"never-store","dest":{"push_endpoint":"https://user:p%40ss@host.test:443/path?token=never-store#never-store","push_endpoint_args":"password=never-store","push_endpoint_topic":"events","stored_secret":true,"persistent":false,"persistent_queue":"queue","time_to_live":"0","max_retries":"18446744073709551615","retry_sleep_duration":"None","unknown_secret":"never-store"}}}`

func TestRGWTopicMetadataCollection(t *testing.T) {
	for _, tc := range []struct {
		name, list, detail string
		count, calls       int
		unavailable        bool
	}{
		{"valid", `["tenant:events"]`, topicDocument, 1, 2, false},
		{"distinct scopes", `[":events","tenant:events","RGW123:events"]`, topicDocument, 3, 4, false},
		{"empty", `[]`, topicDocument, 0, 1, false},
		{"null", `null`, topicDocument, 0, 1, true},
		{"wrong list", `{}`, topicDocument, 0, 1, true},
		{"duplicate", `["tenant:events","tenant:events"]`, topicDocument, 0, 1, true},
		{"invalid key", `["bad\nkey"]`, topicDocument, 0, 1, true},
		{"wrong identity", `["tenant:other"]`, topicDocument, 0, 2, true},
		{"missing detail", `["tenant:events"]`, `{"key":"REQUEST_KEY"}`, 0, 2, true},
		{"bad wrapper", `["tenant:events"]`, strings.Replace(topicDocument, "REQUEST_KEY", "wrong", 1), 0, 2, true},
		{"failed detail", `["tenant:events"]`, "failure", 0, 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			trace := &collectionTrace{unavailable: map[string]struct{}{}}
			ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
			calls := 0
			p := NativeProvider{Executor: topicExecutor{t, tc.list, tc.detail, &calls}}
			rows := p.collectRGWTopics(ctx, ClusterAccess{}, time.Now())
			_, unavailable := trace.unavailable["rgw_topic"]
			if len(rows) != tc.count || calls != tc.calls || unavailable != tc.unavailable {
				t.Fatalf("rows=%d calls=%d unavailable=%v", len(rows), calls, unavailable)
			}
			keys := map[string]bool{}
			for _, row := range rows {
				if keys[row.NaturalKey] {
					t.Fatal("scope collision")
				}
				keys[row.NaturalKey] = true
				body, _ := json.Marshal(row.Payload)
				if strings.Contains(string(body), "never-store") || strings.Contains(string(body), "p%40ss") || strings.Contains(string(body), "push_endpoint_args") {
					t.Fatalf("secret stored: %s", body)
				}
				data := row.Payload.(map[string]any)
				if data["push_endpoint"] != "https://host.test:443/path" || data["endpoint_redacted"] != true || data["persistent"] != false || data["max_retries"] != "18446744073709551615" || data["time_to_live"] != "0" {
					t.Fatal(data)
				}
			}
		})
	}
}

func TestTopicEndpointRedaction(t *testing.T) {
	for _, endpoint := range []string{"http://user:secret@host/?password=secret#secret", "amqp://user:secret@host/vhost", "http://user:secret@%bad", "user:secret", "", "https://host/path"} {
		payload, ok := rgwTopicPayload(map[string]any{"name": "events", "owner": "u", "arn": "a", "dest": map[string]any{"push_endpoint": endpoint}})
		if !ok {
			t.Fatal("missing optional fields must remain unavailable")
		}
		encoded, _ := json.Marshal(payload)
		if strings.Contains(string(encoded), "secret") || strings.Contains(string(encoded), "user:") {
			t.Fatalf("leak %s", encoded)
		}
		if _, ok := payload["persistent"]; ok {
			t.Fatal("invented persistence")
		}
	}
}

func TestTopicMetadataRevisionPrecision(t *testing.T) {
	for _, revision := range []string{"1", "9007199254740993", "18446744073709551615"} {
		version, ok := RGWTopicMetadataVersion(map[string]any{"tag": "tag", "ver": json.Number(revision)})
		if !ok || version != `{"tag":"tag","ver":`+revision+`}` {
			t.Fatalf("revision rounded %s", version)
		}
	}
	for _, value := range []any{nil, map[string]any{}, map[string]any{"tag": "", "ver": json.Number("1")}, map[string]any{"tag": "t", "ver": float64(1)}, map[string]any{"tag": "t", "ver": json.Number("0")}, map[string]any{"tag": "t", "ver": json.Number("-1")}, map[string]any{"tag": "t", "ver": json.Number("18446744073709551616")}} {
		if _, ok := RGWTopicMetadataVersion(value); ok {
			t.Fatal("invalid revision accepted")
		}
	}
}
