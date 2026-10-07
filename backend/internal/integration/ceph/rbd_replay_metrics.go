package ceph

import (
	"encoding/json"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

var replayNumber = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)

// Descriptions originate from journal/ReplayStatusFormatter and snapshot/Replayer.
// Keep validated numbers as text so inventory clients do not round uint64 values.
func mirrorReplayMetrics(row map[string]any) map[string]string {
	description, ok := row["description"].(string)
	if !ok || row["state"] != "up+replaying" || !strings.HasPrefix(description, "replaying, ") || len(description) > 65536 {
		return nil
	}
	var data map[string]json.RawMessage
	if json.Unmarshal([]byte(strings.TrimPrefix(description, "replaying, ")), &data) != nil {
		return nil
	}
	result := map[string]string{}
	for _, key := range []string{"bytes_per_second", "seconds_until_synced", "syncing_percent", "entries_behind_primary"} {
		raw := string(data[key])
		if len(raw) > 128 || !replayNumber.MatchString(raw) {
			continue
		}
		if key == "seconds_until_synced" || key == "entries_behind_primary" {
			if _, err := strconv.ParseUint(raw, 10, 64); err != nil {
				continue
			}
		} else {
			number, err := strconv.ParseFloat(raw, 64)
			if err != nil || math.IsInf(number, 0) || number == 0 && strings.ContainsAny(strings.Split(strings.ToLower(raw), "e")[0], "123456789") {
				continue
			}
			if key == "syncing_percent" && number != 0 {
				ratio, ok := new(big.Rat).SetString(raw)
				if !ok || ratio.Cmp(big.NewRat(100, 1)) > 0 {
					continue
				}
			}
		}
		result[key] = raw
	}
	return result
}

func enrichMirrorReplayMetrics(value any) {
	rows, _ := value.([]any)
	for _, value := range rows {
		row, ok := value.(map[string]any)
		if !ok {
			continue
		}
		if metrics := mirrorReplayMetrics(row); len(metrics) > 0 {
			row["replay_metrics"] = metrics
		}
		enrichMirrorReplayMetrics(row["peer_sites"])
	}
}
