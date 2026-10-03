package ceph

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"unicode"

	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/integration/ceph/s3"
)

// The Dashboard enumerates topic metadata, not a default tenant's topic list.
// Never persist endpoint authentication arguments or URL credentials.
func (p *NativeProvider) collectRGWTopics(ctx context.Context, access ClusterAccess, now time.Time) []Observation {
	var keys []string
	if !p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_topic", []string{"metadata", "list", "topic", "--format", "json"}, &keys) {
		return nil
	}
	seen := map[string]bool{}
	if keys == nil {
		markCollectionUnavailable(ctx, "collect.rgw_topic")
		return nil
	}
	for _, key := range keys {
		if key == "" || strings.TrimSpace(key) != key || strings.IndexFunc(key, unicode.IsControl) >= 0 || seen[key] {
			markCollectionUnavailable(ctx, "collect.rgw_topic")
			return nil
		}
		seen[key] = true
	}
	rows := []Observation{}
	for _, key := range keys {
		var document map[string]any
		if !p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_topic_detail", []string{"metadata", "get", "topic:" + key, "--format", "json"}, &document) {
			continue
		}
		data, ok := document["data"].(map[string]any)
		if !ok || document["key"] != "topic:"+key {
			markCollectionUnavailable(ctx, "collect.rgw_topic_detail")
			continue
		}
		payload, ok := rgwTopicPayload(data)
		scope, name, scoped := strings.Cut(key, ":")
		if !ok || !scoped || name != payload["name"] {
			markCollectionUnavailable(ctx, "collect.rgw_topic_detail")
			continue
		}
		payload["metadata_key"] = key
		payload["scope"] = scope
		if version, valid := RGWTopicMetadataVersion(document["ver"]); valid {
			payload["metadata_version"] = version
		}
		rows = append(rows, observation("rgw_topic", base64.RawURLEncoding.EncodeToString([]byte(key)), payload["name"].(string), "rgw_admin", payload, now))
	}
	return rows
}

// Keep the native uint64 revision in an opaque JSON string across browser DTOs.
func RGWTopicMetadataVersion(value any) (string, bool) {
	version, ok := value.(map[string]any)
	if !ok || len(version) != 2 {
		return "", false
	}
	tag, ok := version["tag"].(string)
	if !ok || tag == "" {
		return "", false
	}
	number, ok := version["ver"].(json.Number)
	if !ok {
		return "", false
	}
	n, err := strconv.ParseUint(number.String(), 10, 64)
	if err != nil || n == 0 {
		return "", false
	}
	body, err := json.Marshal(map[string]any{"tag": tag, "ver": json.Number(strconv.FormatUint(n, 10))})
	return string(body), err == nil
}

func rgwTopicPayload(data map[string]any) (map[string]any, bool) {
	payload := map[string]any{}
	for _, field := range []string{"name", "owner", "arn"} {
		value, ok := data[field].(string)
		if !ok || value == "" {
			return nil, false
		}
		payload[field] = value
	}
	dest, ok := data["dest"].(map[string]any)
	if !ok {
		return nil, false
	}
	// Opaque data and policy are user-facing configuration, not endpoint auth.
	for _, field := range []string{"opaqueData", "policy"} {
		if value, ok := data[field].(string); ok {
			payload[field] = value
		}
	}
	for _, field := range []string{"push_endpoint_topic", "persistent_queue", "time_to_live", "max_retries", "retry_sleep_duration"} {
		if value, ok := dest[field].(string); ok {
			payload[field] = value
		}
	}
	for _, field := range []string{"persistent", "stored_secret"} {
		if value, ok := dest[field].(bool); ok {
			payload[field] = value
		}
	}
	endpoint, ok := dest["push_endpoint"].(string)
	if !ok {
		return nil, false
	}
	if visible, redacted, valid := s3.RedactTopicEndpoint(endpoint); valid {
		payload["push_endpoint"], payload["endpoint_redacted"] = visible, redacted
	} else {
		// Malformed URLs may contain passwords: never fall back to raw text.
		payload["push_endpoint"] = nil
		payload["endpoint_redacted"] = true
	}
	payload["endpoint_args_hidden"] = true
	return payload, true
}
