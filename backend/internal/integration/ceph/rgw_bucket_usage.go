package ceph

import "encoding/json"

// Preserve usage counters before browser decoding can round native uint64 data.
// Do not infer missing categories or counters as zero.
func preserveRGWBucketUsage(details map[string]any) {
	usage, _ := details["usage"].(map[string]any)
	for _, category := range usage {
		counters, _ := category.(map[string]any)
		for _, field := range []string{"size", "size_actual", "size_utilized", "size_kb", "size_kb_actual", "size_kb_utilized", "num_objects"} {
			if value, ok := counters[field].(json.Number); ok {
				counters[field] = value.String()
			}
		}
	}
}
