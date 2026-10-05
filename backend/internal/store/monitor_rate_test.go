package store

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMonitorRateUnknownIsNotZero(t *testing.T) {
	now := time.Now()
	for _, tc := range []struct {
		name, current, prior string
		delta                time.Duration
		want                 any
	}{
		{"first", `{"metric_type":"counter","raw_value":100,"value":100}`, "", time.Second, nil},
		{"reset", `{"metric_type":"counter","raw_value":1,"value":1}`, `{"metric_type":"counter","raw_value":100}`, time.Second, nil},
		{"same time", `{"metric_type":"counter","raw_value":100}`, `{"metric_type":"counter","raw_value":100}`, 0, nil},
		{"backwards", `{"metric_type":"counter","raw_value":100}`, `{"metric_type":"counter","raw_value":50}`, -time.Second, nil},
		{"missing current", `{"metric_type":"counter","value":99}`, `{"metric_type":"counter","raw_value":50}`, time.Second, nil},
		{"bad prior", `{"metric_type":"counter","raw_value":100}`, `{}`, time.Second, nil},
		{"changed type", `{"metric_type":"counter","raw_value":100}`, `{"metric_type":"gauge","raw_value":50}`, time.Second, nil},
		{"changed unit", `{"metric_type":"counter","raw_value":100,"unit":"B/s"}`, `{"metric_type":"counter","raw_value":50,"unit":"/s"}`, time.Second, nil},
		{"invalid unit", `{"metric_type":"counter","raw_value":100,"unit":{}}`, `{"metric_type":"counter","raw_value":50,"unit":{}}`, time.Second, nil},
		{"negative", `{"metric_type":"counter","raw_value":100}`, `{"metric_type":"counter","raw_value":-1}`, time.Second, nil},
		{"zero rate", `{"metric_type":"counter","raw_value":100}`, `{"metric_type":"counter","raw_value":100}`, time.Second, float64(0)},
		{"valid", `{"metric_type":"counter","raw_value":142}`, `{"metric_type":"counter","raw_value":100}`, 10 * time.Second, 4.2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			row := CephEntityRecord{Kind: "mon_perf_counter", DiscoveredData: tc.current, ObservedAt: now.Add(tc.delta)}
			var prior *CephEntityRecord
			if tc.prior != "" {
				prior = &CephEntityRecord{DiscoveredData: tc.prior, ObservedAt: now}
			}
			applyMonitorCounterRate(&row, prior)
			var payload map[string]any
			if json.Unmarshal([]byte(row.DiscoveredData), &payload) != nil {
				t.Fatal("invalid JSON")
			}
			if value, present := payload["value"]; !present || value != tc.want {
				t.Fatalf("got %v want %v", value, tc.want)
			}
		})
	}
}
